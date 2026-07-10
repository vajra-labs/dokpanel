package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Global connection pool
var pool sync.Map // map[string]*SSHClient

// SSHConfig holds the connection parameters for a remote server.
type SSHConfig struct {
	Host       string
	Port       string
	User       string
	PrivateKey string
}

// SSHClient is a lightweight handle for a registered server.
// Obtain one via GetSSHClient after registering with NewSSH.
type SSHClient struct {
	serverId string
	mu       sync.RWMutex
	conn     *ssh.Client
	cfg      *SSHConfig
}

// NewSSH dials the remote server, establishes an SSH connection, and stores
// it in the global pool under serverId. Calling NewSSH again with the same
// serverId replaces the existing connection.
func NewSSH(serverId string, cfg SSHConfig) error {
	signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
	if err != nil {
		return &ExecError{
			Message: fmt.Sprintf("Invalid SSH private key: %v", err),
			Err:     err,
		}
	}
	client, err := ssh.Dial(
		"tcp",
		cfg.Host+":"+cfg.Port, &ssh.ClientConfig{
			User: cfg.User,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         30 * time.Second,
		})
	if err != nil {
		return newSSHConnError(serverId, err)
	}
	pool.Store(serverId, &SSHClient{
		serverId: serverId,
		cfg:      &cfg,
		conn:     client,
	})
	return nil
}

// GetSSHClient retrieves a registered SSHClient handle from the pool.
// Returns an error if NewSSH has not been called for this serverId.
func GetSSHClient(serverId string) (*SSHClient, error) {
	val, ok := pool.Load(serverId)
	if !ok {
		return nil, &ExecError{
			Message: fmt.Sprintf("Server %q not connected", serverId),
		}
	}
	return val.(*SSHClient), nil
}

// Exec runs a command on the remote server and returns a channel that receives
// the result once execution completes.
//
// Each call opens a new SSH session on the existing TCP connection — lightweight
// and safe to call concurrently. If onData is set, output chunks are forwarded
// in real-time. The context is honoured — cancelling it closes the session.
func (c *SSHClient) Exec(ctx context.Context, command string, onData func(string)) <-chan ExecResult {
	ch := make(chan ExecResult, 1)
	go func() {
		defer close(ch)

		session, err := c.newSession()
		if err != nil {
			ch <- ExecResult{Err: err}
			return
		}
		defer session.Close()

		done := make(chan struct{})
		defer close(done)
		go func() {
			select {
			case <-ctx.Done():
				session.Close()
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

// reconnect dials a fresh SSH connection using the stored config.
func (c *SSHClient) reconnect() (*ssh.Client, error) {
	signer, err := ssh.ParsePrivateKey([]byte(c.cfg.PrivateKey))
	if err != nil {
		return nil, &ExecError{
			Message: fmt.Sprintf("Parse private key: %v", err),
			Err:     err,
		}
	}
	client, err := ssh.Dial(
		"tcp",
		c.cfg.Host+":"+c.cfg.Port, &ssh.ClientConfig{
			User: c.cfg.User,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         30 * time.Second,
		})
	if err != nil {
		return nil, newSSHConnError(c.serverId, err)
	}
	// Update pool with fresh connection
	pool.Store(c.serverId, c)
	return client, nil
}

// newSession opens a new SSH session, reconnecting once if the connection is stale.
func (c *SSHClient) newSession() (*ssh.Session, error) {
	// Multiple goroutines can read conn simultaneously
	c.mu.RLock()
	session, err := c.conn.NewSession()
	c.mu.RUnlock()
	if err == nil {
		return session, nil
	}
	// Full lock for reconnect
	// Only one goroutine writes conn
	c.mu.Lock()
	defer c.mu.Unlock()
	// Double-check
	// Another goroutine may have already reconnected
	session, err = c.conn.NewSession()
	if err == nil {
		return session, nil
	}
	// Attempt reconnect
	newConn, err := c.reconnect()
	if err != nil {
		return nil, err
	}
	c.conn = newConn
	session, err = c.conn.NewSession()
	if err != nil {
		return nil, &ExecError{
			Message: fmt.Sprintf("SSH new session: %v", err),
			Err:     err,
		}
	}
	return session, nil
}

// Close removes the client from the pool and closes the underlying connection.
func (c *SSHClient) Close() error {
	pool.Delete(c.serverId)
	return c.conn.Close()
}

// SSHCloseAll closes every connection in the pool.
func SSHCloseAll() {
	pool.Range(func(key, val any) bool {
		_ = val.(*SSHClient).conn.Close()
		pool.Delete(key)
		return true
	})
}

// sshExecSimple runs the command and captures stdout/stderr into buffers.
// Used when no streaming callback is provided.
func sshExecSimple(session *ssh.Session, serverId, command string) ExecResult {
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return ExecResult{Err: newSSHExecError(
			command, stdout.String(),
			stderr.String(), err,
			serverId,
		)}
	}
	return ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
}

// sshExecStream runs the command and forwards output to onData if set.
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
		return ExecResult{Err: newSSHExecError(
			command, "", "",
			err, serverId,
		)}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); io.Copy(stdoutWriter, stdoutPipe) }()
	go func() { defer wg.Done(); io.Copy(stderrWriter, stderrPipe) }()

	err = session.Wait()
	wg.Wait()

	if err != nil {
		return ExecResult{Err: newSSHExecError(
			command, stdout.String(),
			stderr.String(), err,
			serverId,
		)}
	}
	return ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
}
