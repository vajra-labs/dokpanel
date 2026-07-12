package shell

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/semaphore"
)

var pool sync.Map // map[string]*SSHClient

const (
	timeout     = 10 * time.Second
	maxSessions = 5 // keep below OpenSSH MaxSessions
)

type SSHConfig struct {
	Host       string
	Port       string
	User       string
	PrivateKey string
}

type SSHClient struct {
	serverId string
	cfg      SSHConfig
	mu       sync.RWMutex // guards conn field
	conn     *ssh.Client
	sem      *semaphore.Weighted // limits concurrent sessions to maxSessions
}

// dial opens an authenticated SSH connection to the server described by cfg.
func dial(cfg SSHConfig) (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
	if err != nil {
		return nil, err
	}
	return ssh.Dial("tcp", cfg.Host+":"+cfg.Port, &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	})
}

// SetSSHClient connects to the server and registers it in the pool under serverId.
// Calling again with the same serverId replaces the existing connection.
func SetSSHClient(serverId string, cfg SSHConfig) error {
	client, err := dial(cfg)
	if err != nil {
		return newSSHConnError(serverId, err)
	}
	newClient := &SSHClient{
		serverId: serverId,
		cfg:      cfg,
		conn:     client,
		sem:      semaphore.NewWeighted(maxSessions),
	}
	if old, loaded := pool.Swap(serverId, newClient); loaded {
		// close the previous connection for this serverId
		oldClient := old.(*SSHClient)
		oldClient.mu.Lock()
		_ = oldClient.conn.Close()
		oldClient.mu.Unlock()
	}
	return nil
}

// GetSSHClient retrieves a registered SSHClient from the pool.
func GetSSHClient(serverId string) (*SSHClient, error) {
	val, ok := pool.Load(serverId)
	if !ok {
		return nil, &ExecError{Message: fmt.Sprintf("server %q not found in ssh pool", serverId)}
	}
	return val.(*SSHClient), nil
}

// DelSSHClient removes a server from the pool and closes its connection.
func DelSSHClient(serverId string) error {
	val, ok := pool.Load(serverId)
	if !ok {
		return &ExecError{Message: fmt.Sprintf("server %q not found in ssh pool", serverId)}
	}
	client := val.(*SSHClient)
	pool.Delete(client.serverId)
	client.mu.Lock()
	defer client.mu.Unlock()
	return client.conn.Close()
}

// DelAllSSHClient closes every connection in the pool.
func DelAllSSHClient() {
	pool.Range(func(key, val any) bool {
		client := val.(*SSHClient)
		client.mu.Lock()
		_ = client.conn.Close()
		client.mu.Unlock()
		pool.Delete(key)
		return true
	})
}

// Exec runs a command on the remote server and returns a channel that receives
// the result once execution completes. Blocks if maxSessions slots are in use.
func (c *SSHClient) Exec(ctx context.Context, command string, onData func(string)) <-chan ExecResult {
	ch := make(chan ExecResult, 1)
	go func() {
		defer close(ch)
		// block until a session slot is free; respect ctx cancellation
		if err := c.sem.Acquire(ctx, 1); err != nil {
			ch <- ExecResult{Err: &ExecError{
				Message:  fmt.Sprintf("session limit reached: %v", err),
				Command:  command,
				ServerID: &c.serverId,
				Err:      err,
			}}
			return
		}
		defer c.sem.Release(1)
		// snapshot conn under RLock before opening a session
		c.mu.RLock()
		deadConn := c.conn
		session, err := c.conn.NewSession()
		c.mu.RUnlock()
		if err != nil {
			// OpenChannelError = server explicitly rejected (e.g. MaxSessions hit)
			// reconnecting won't help in this case
			if oe, ok := errors.AsType[*ssh.OpenChannelError](err); ok {
				ch <- ExecResult{Err: oe}
				return
			}
			// connection is stale, reconnect once and retry
			c.mu.Lock()
			reconnErr := c.reconnect(deadConn)
			if reconnErr != nil {
				c.mu.Unlock()
				ch <- ExecResult{Err: reconnErr}
				return
			}
			session, err = c.conn.NewSession()
			c.mu.Unlock()
			if err != nil {
				ch <- ExecResult{Err: newSSHExecError(command, "", "", err, c.serverId)}
				return
			}
		}
		defer func() { _ = session.Close() }()
		// signal SIGKILL to the remote process if ctx is canceled
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
			ch <- sshExecStream(session, c.serverId, command, onData)
		} else {
			ch <- sshExecSimple(session, c.serverId, command)
		}
	}()
	return ch
}

// reconnect replaces a stale connection with a fresh one.
// Skips silently if another goroutine already reconnected (c.conn != deadConn).
// Caller must hold c.mu.Lock().
func (c *SSHClient) reconnect(deadConn *ssh.Client) error {
	if c.conn != deadConn {
		return nil
	}
	newConn, err := dial(c.cfg)
	if err != nil {
		return newSSHConnError(c.serverId, err)
	}
	_ = c.conn.Close() // best effort — may already be dead
	c.conn = newConn
	return nil
}

// sshExecSimple runs the command and captures stdout/stderr into buffers.
func sshExecSimple(session *ssh.Session, serverId, command string) ExecResult {
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if err := session.Run(command); err != nil {
		return ExecResult{Err: newSSHExecError(command, stdout.String(), stderr.String(), err, serverId)}
	}
	return ExecResult{Stdout: stdout.String(), Stderr: stderr.String()}
}

// sshExecStream runs the command and forwards output chunks to onData in real-time.
func sshExecStream(session *ssh.Session, serverId, command string, onData func(string)) ExecResult {
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
		return ExecResult{Err: newSSHExecError(command, "", "", err, serverId)}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(stdoutWriter, stdoutPipe) }()
	go func() { defer wg.Done(); _, _ = io.Copy(stderrWriter, stderrPipe) }()

	err = session.Wait()
	wg.Wait()

	if err != nil {
		return ExecResult{Err: newSSHExecError(command, stdout.String(), stderr.String(), err, serverId)}
	}
	return ExecResult{Stdout: stdout.String(), Stderr: stderr.String()}
}
