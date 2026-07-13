package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// EnvMap is a convenience alias for env vars.
type EnvMap = map[string]string

// ExecOptions configures how a command is executed.
type ExecOptions struct {
	// Dir sets the working directory. Defaults to current directory.
	Dir string
	// Env adds extra environment variables.
	// Merged on top of the current process environment.
	Env []string
	// Shell sets the shell executable (e.g. "/bin/bash", "/bin/sh").
	// When set, enables $VARIABLE expansion, pipes (|), &&, || etc.
	// When empty, command runs as a direct binary (faster, safer).
	Shell string
	// Stdin optionally feeds data into the command's standard input.
	// Useful for tools like psql, mysql, openssl that read from stdin.
	Stdin io.Reader
	// Called whenever stdout/stderr receives new data.
	// The callback receives raw chunks exactly as they are produced.
	OnData func(string)
}

type ExecOption func(*ExecOptions)

// WithDir sets the working directory.
func WithDir(dir string) ExecOption {
	return func(opts *ExecOptions) {
		opts.Dir = dir
	}
}

// WithEnv adds a single environment variable.
func WithEnv(key, value string) ExecOption {
	return func(opts *ExecOptions) {
		opts.Env = append(opts.Env, key+"="+value)
	}
}

// WithEnvMap adds multiple environment variables at once.
func WithEnvMap(env EnvMap) ExecOption {
	return func(opts *ExecOptions) {
		for k, v := range env {
			opts.Env = append(opts.Env, k+"="+v)
		}
	}
}

// WithShell sets the shell executable.
func WithShell(shell string) ExecOption {
	return func(opts *ExecOptions) {
		opts.Shell = shell
	}
}

// WithStdin sets the stdin reader for the command.
func WithInput(str string) ExecOption {
	return func(opts *ExecOptions) {
		opts.Stdin = strings.NewReader(str)
	}
}

// WithOnData sets a callback that receives real-time output chunks.
func WithOnData(fn func(string)) ExecOption {
	return func(o *ExecOptions) {
		o.OnData = fn
	}
}

// ExecResult holds the output of a completed command.
type ExecResult struct {
	Stdout string
	Stderr string
	Err    error
}

// Exec runs a command asynchronously and returns a channel that will
// receive the command result once execution completes.
//
// By default, the command is executed directly without a shell, so shell
// features like environment variable expansion ($VAR), pipes (|),
// redirection (>), &&, ||, and wildcards (*) are not supported.
//
// To enable shell syntax, specify a shell using WithShell.
// To stream output in real-time, specify a callback using WithOnData.
func Exec(ctx context.Context, command string, options ...ExecOption) <-chan ExecResult {
	ch := make(chan ExecResult, 1)

	opts := &ExecOptions{}
	for _, option := range options {
		option(opts)
	}

	go func() {
		defer close(ch)
		cmd := buildCmd(ctx, command, opts)
		if opts.OnData != nil {
			ch <- execStream(cmd, command, opts.OnData)
		} else {
			ch <- execSimple(cmd, command)
		}
	}()

	return ch
}

// execSimple runs the command and captures output into buffers.
// Used when no streaming callback is provided.
func execSimple(cmd *exec.Cmd, command string) ExecResult {
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		execErr := newExecError(
			command, stdout.String(),
			stderr.String(), err,
		)
		return ExecResult{Err: execErr}
	}

	return ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
}

// execStream runs the command with stdout/stderr pipes and forwards
// each chunk to the onData callback in real-time.
func execStream(cmd *exec.Cmd, command string, onData func(string)) ExecResult {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return ExecResult{Err: fmt.Errorf("stdout pipe: %w", err)}
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return ExecResult{Err: fmt.Errorf("stderr pipe: %w", err)}
	}

	var stdout, stderr bytes.Buffer
	stdoutWriter := &streamWriter{buf: &stdout, onData: onData}
	stderrWriter := &streamWriter{buf: &stderr, onData: onData}

	if err := cmd.Start(); err != nil {
		execErr := newExecError(command, "", "", err)
		return ExecResult{Err: execErr}
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(stdoutWriter, stdoutPipe) }()
	go func() { defer wg.Done(); _, _ = io.Copy(stderrWriter, stderrPipe) }()

	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		execErr := newExecError(
			command, stdout.String(),
			stderr.String(), err,
		)
		return ExecResult{Err: execErr}
	}

	return ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}
}

type streamWriter struct {
	buf    *bytes.Buffer
	onData func(string)
}

func (w *streamWriter) Write(p []byte) (int, error) {
	n, err := w.buf.Write(p)
	if err != nil {
		return n, err
	}
	if w.onData != nil {
		w.onData(string(p))
	}
	return n, nil
}

// buildCmd constructs an *exec.Cmd from a command string and options.
func buildCmd(ctx context.Context, command string, opts *ExecOptions) *exec.Cmd {
	var cmd *exec.Cmd
	if opts.Shell == "" {
		// Direct binary — no shell, faster and safer
		parts := strings.Fields(command)
		cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
	} else {
		// -e: exit immediately on error (fail fast)
		// -c: execute the command string
		cmd = exec.CommandContext(ctx, opts.Shell, "-ec", command)
	}
	cmd.Dir = opts.Dir
	if opts.Stdin != nil {
		cmd.Stdin = opts.Stdin
	}
	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	return cmd
}
