package shellx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"goploy/src/db/repos"

	"go.uber.org/fx"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

const (
	timeout     = 10 * time.Second
	maxSessions = 5 // keep below OpenSSH MaxSessions
)

// SSHClient wraps a single SSH connection with a session semaphore
// to limit concurrent sessions per server.
type SSHClient struct {
	client *ssh.Client
	sem    *semaphore.Weighted
}

// Conn returns the underlying ssh.Client — used for tunneling (e.g. remote Docker).
func (c *SSHClient) Client() *ssh.Client {
	return c.client
}

// SSHPool holds SSH connections for multiple servers.
// Uses singleflight so only one dial happens per server at a time.
type SSHPool struct {
	mu      sync.RWMutex
	pools   map[int64]*SSHClient
	group   singleflight.Group
	queries *repos.Queries
}

// NewSSHPool creates the pool and registers a cleanup hook on app shutdown.
func NewSSHPool(lc fx.Lifecycle, query *repos.Queries) *SSHPool {
	p := &SSHPool{
		pools:   make(map[int64]*SSHClient),
		queries: query,
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			p.CloseAll()
			return nil
		},
	})
	return p
}

// Get returns a live connection for the server, dialing fresh if missing or stale.
func (p *SSHPool) Get(ctx context.Context, serverId int64) (*SSHClient, error) {
	p.mu.RLock()
	sc, ok := p.pools[serverId]
	p.mu.RUnlock()

	if ok {
		// quick ping to check if the connection is still alive
		_, _, err := sc.client.SendRequest("keepalive@golang.org", true, nil)
		if err == nil {
			return sc, nil
		}
		p.Close(serverId)
	}

	// singleflight ensures only one dial per server runs at a time
	v, err, _ := p.group.Do(fmt.Sprint(serverId), func() (any, error) {
		client, err := p.dial(ctx, serverId)
		if err != nil {
			return nil, err
		}
		sshClient := &SSHClient{
			client: client,
			sem:    semaphore.NewWeighted(maxSessions),
		}
		p.mu.Lock()
		p.pools[serverId] = sshClient
		p.mu.Unlock()
		return sshClient, nil
	})
	if err != nil {
		return nil, err
	}

	return v.(*SSHClient), nil
}

// dial fetches credentials from DB and opens a fresh SSH connection.
func (p *SSHPool) dial(
	ctx context.Context,
	serverId int64,
) (*ssh.Client, error) {
	serverIdStr := fmt.Sprint(serverId)

	cfg, err := p.queries.GetServerSSHCredentials(ctx, serverId)
	if err != nil {
		return nil, &ExecError{
			Message:  fmt.Sprintf("Failed to fetch SSH credentials: %v", err),
			ServerID: &serverIdStr,
			Err:      err,
		}
	}

	// private key is required — bail early with a clear message
	if cfg.PrivateKey == nil || *cfg.PrivateKey == "" {
		return nil, &ExecError{
			Message:  "SSH private key is missing for this server",
			ServerID: &serverIdStr,
		}
	}

	signer, err := ssh.ParsePrivateKey([]byte(*cfg.PrivateKey))
	if err != nil {
		return nil, newSSHConnError(serverIdStr, err)
	}

	client, err := ssh.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", cfg.IpAddress, cfg.Port),
		&ssh.ClientConfig{
			User:            cfg.Username,
			Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         10 * time.Second,
		},
	)
	if err != nil {
		return nil, newSSHConnError(serverIdStr, err)
	}

	return client, nil
}

// Close removes and closes one server's connection from the pool.
func (p *SSHPool) Close(serverId int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if sc, ok := p.pools[serverId]; ok {
		_ = sc.client.Close()
		delete(p.pools, serverId)
	}
}

// CloseAll shuts down every connection. Called on app shutdown.
func (p *SSHPool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for id, sc := range p.pools {
		_ = sc.client.Close()
		delete(p.pools, id)
	}
}

// Exec runs a command on a remote server.
// Pass onData to stream output live, or nil to get it all in ExecResult.
func (p *SSHPool) Exec(
	ctx context.Context, serverId int64,
	command string, onData func(string),
) <-chan ExecResult {
	serverIdStr := fmt.Sprint(serverId)
	ch := make(chan ExecResult, 1)

	go func() {
		defer close(ch)

		sc, err := p.Get(ctx, serverId)
		if err != nil {
			ch <- ExecResult{Err: err}
			return
		}

		// wait for a free session slot — respects ctx cancellation
		if err := sc.sem.Acquire(ctx, 1); err != nil {
			ch <- ExecResult{Err: &ExecError{
				Message:  fmt.Sprintf("session limit reached: %v", err),
				Command:  command,
				ServerID: &serverIdStr,
				Err:      err,
			}}
			return
		}
		defer sc.sem.Release(1)

		session, err := sc.client.NewSession()
		if err != nil {
			ch <- ExecResult{
				Err: newSSHExecError(command, "", "", err, serverIdStr),
			}
			return
		}
		defer func() { _ = session.Close() }()
		// if ctx is canceled, kill the remote process
		done := make(chan struct{})
		defer close(done)
		go func() {
			select {
			case <-ctx.Done():
				_ = session.Signal(ssh.SIGKILL)
				_ = session.Close()
			case <-done:
			}
		}()
		if onData != nil {
			ch <- sshExecStream(session, serverIdStr, command, onData)
		} else {
			ch <- sshExecSimple(session, serverIdStr, command)
		}
	}()

	return ch
}

// sshExecSimple runs the command and captures stdout/stderr into buffers.
func sshExecSimple(session *ssh.Session, serverId, command string) ExecResult {
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if err := session.Run(command); err != nil {
		return ExecResult{
			Err: newSSHExecError(
				command, stdout.String(),
				stderr.String(), err, serverId,
			),
		}
	}
	return ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
}

// sshExecStream runs the command and forwards output chunks to onData in real-time.
func sshExecStream(
	session *ssh.Session, serverId,
	command string, onData func(string),
) ExecResult {
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return ExecResult{Err: fmt.Errorf("ssh stdout pipe: %w", err)}
	}
	stderrPipe, err := session.StderrPipe()
	if err != nil {
		return ExecResult{Err: fmt.Errorf("ssh stderr pipe: %w", err)}
	}

	var stdout, stderr bytes.Buffer
	stdoutWriter := &streamWriter{buf: &stdout, onData: onData}
	stderrWriter := &streamWriter{buf: &stderr, onData: onData}

	if err := session.Start(command); err != nil {
		return ExecResult{
			Err: newSSHExecError(command, "", "", err, serverId),
		}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(stdoutWriter, stdoutPipe) }()
	go func() { defer wg.Done(); _, _ = io.Copy(stderrWriter, stderrPipe) }()

	err = session.Wait()
	wg.Wait()

	if err != nil {
		return ExecResult{
			Err: newSSHExecError(
				command, stdout.String(),
				stderr.String(), err, serverId,
			),
		}
	}
	return ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
}
