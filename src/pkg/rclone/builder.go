package rclone

import (
	"fmt"
	"strings"
)

// Builder accumulates rclone flags and targets via method chaining.
// Call String() to preview the final command, or Run/RunWithOutput to execute it.
type Builder struct {
	command     Command
	source      Target
	destination Target

	// Performance & Optimization
	transfers  *uint32 // --transfers: number of file transfers to run in parallel
	checkers   *uint32 // --checkers: number of checkers to run in parallel
	bufferSize string  // --buffer-size: size of in-memory buffer e.g. "16M"
	bwlimit    string  // --bwlimit: bandwidth limit e.g. "10M" or "1M:off"
	fastList   bool    // --fast-list: use fewer API calls (S3/GCS, uses more memory)
	useMmap    bool    // --use-mmap: use mmap allocator (may reduce memory spikes)

	// Timeouts & Retries
	retries         *uint32 // --retries: number of times to retry the entire operation
	lowLevelRetries *uint32 // --low-level-retries: retries for individual operations
	timeout         string  // --timeout: I/O idle timeout e.g. "5m"
	connectTimeout  string  // --contimeout: connection timeout e.g. "10s"
	retryDelay      string  // --retry-delay: delay between retries e.g. "10s"

	// Safety & Behaviour
	dryRun         bool // --dry-run: show what would be done without doing it
	checkFirst     bool // --check-first: check all files before starting transfers
	ignoreErrors   bool // --ignore-errors: delete even if there are errors
	ignoreExisting bool // --ignore-existing: skip files that already exist on destination
	update         bool // --update: skip files newer on the destination
	inplace        bool // --inplace: write files directly to final location (no temp file)

	// Logging
	logFile  string // --log-file: path to log file
	logLevel string // --log-level: DEBUG, INFO, NOTICE, ERROR
	stats    string // --stats: interval for printing stats e.g. "1m"

	// Extra raw flags passed through verbatim
	extraFlags []string
}

// NewBuilder creates a new Builder for the given rclone command.
func NewBuilder(cmd Command) *Builder {
	return &Builder{command: cmd}
}

// Source sets the source target.
func (b *Builder) Source(t Target) *Builder {
	b.source = t
	return b
}

// Destination sets the destination target.
func (b *Builder) Destination(t Target) *Builder {
	b.destination = t
	return b
}

// Transfers sets the number of file transfers to run in parallel (--transfers).
func (b *Builder) Transfers(n uint32) *Builder {
	b.transfers = &n
	return b
}

// Checkers sets the number of checkers to run in parallel (--checkers).
func (b *Builder) Checkers(n uint32) *Builder {
	b.checkers = &n
	return b
}

// BufferSize sets the in-memory buffer size per transfer (--buffer-size).
// e.g. "16M", "64M"
func (b *Builder) BufferSize(size string) *Builder {
	b.bufferSize = size
	return b
}

// Bwlimit sets a bandwidth limit (--bwlimit).
// e.g. "10M" (10 MB/s), "1M:off" (1 MB/s during day, off at night)
func (b *Builder) Bwlimit(limit string) *Builder {
	b.bwlimit = limit
	return b
}

// FastList uses fewer API calls to list files, trading memory for speed (--fast-list).
// Recommended for large S3/GCS buckets.
func (b *Builder) FastList() *Builder {
	b.fastList = true
	return b
}

// UseMmap uses mmap for memory allocation, reducing GC pressure (--use-mmap).
func (b *Builder) UseMmap() *Builder {
	b.useMmap = true
	return b
}

// Retries sets the number of times to retry the entire operation (--retries).
func (b *Builder) Retries(n uint32) *Builder {
	b.retries = &n
	return b
}

// LowLevelRetries sets the number of retries for individual low-level operations (--low-level-retries).
func (b *Builder) LowLevelRetries(n uint32) *Builder {
	b.lowLevelRetries = &n
	return b
}

// Timeout sets the I/O idle timeout (--timeout).
// e.g. "5m", "30s"
func (b *Builder) Timeout(t string) *Builder {
	b.timeout = t
	return b
}

// ConnectTimeout sets the connection timeout (--contimeout).
// e.g. "10s", "1m"
func (b *Builder) ConnectTimeout(t string) *Builder {
	b.connectTimeout = t
	return b
}

// RetryDelay sets the delay between retries (--retry-delay).
// e.g. "10s"
func (b *Builder) RetryDelay(d string) *Builder {
	b.retryDelay = d
	return b
}

// DryRun prints what would be done without actually doing it (--dry-run).
func (b *Builder) DryRun() *Builder {
	b.dryRun = true
	return b
}

// CheckFirst checks all files before starting any transfers (--check-first).
func (b *Builder) CheckFirst() *Builder {
	b.checkFirst = true
	return b
}

// IgnoreErrors continues and deletes even when errors occur (--ignore-errors).
func (b *Builder) IgnoreErrors() *Builder {
	b.ignoreErrors = true
	return b
}

// IgnoreExisting skips files that already exist at the destination (--ignore-existing).
func (b *Builder) IgnoreExisting() *Builder {
	b.ignoreExisting = true
	return b
}

// Update skips files that are newer at the destination (--update).
func (b *Builder) Update() *Builder {
	b.update = true
	return b
}

// Inplace writes files directly to their final location instead of a temp file (--inplace).
func (b *Builder) Inplace() *Builder {
	b.inplace = true
	return b
}

// LogFile writes all log output to the given file path (--log-file).
func (b *Builder) LogFile(path string) *Builder {
	b.logFile = path
	return b
}

// LogLevel sets the log verbosity (--log-level).
func (b *Builder) LogLevel(level LogLevel) *Builder {
	b.logLevel = string(level)
	return b
}

// Stats sets the interval for printing transfer statistics (--stats).
// e.g. "1m", "10s". Set to "0" to disable.
func (b *Builder) Stats(interval string) *Builder {
	b.stats = interval
	return b
}

// Flag appends a raw flag string verbatim (for flags not yet covered by a method).
// e.g. b.Flag("--no-traverse")
func (b *Builder) Flag(flag string) *Builder {
	b.extraFlags = append(b.extraFlags, flag)
	return b
}

// build assembles the final args slice and env map.
func (b *Builder) build() (args []string, envs map[string]string) {
	envs = make(map[string]string)
	args = append(args, string(b.command))

	if b.transfers != nil {
		args = append(args, fmt.Sprintf("--transfers=%d", *b.transfers))
	}
	if b.checkers != nil {
		args = append(args, fmt.Sprintf("--checkers=%d", *b.checkers))
	}
	if b.bufferSize != "" {
		args = append(args, "--buffer-size="+b.bufferSize)
	}
	if b.bwlimit != "" {
		args = append(args, "--bwlimit="+b.bwlimit)
	}
	if b.fastList {
		args = append(args, "--fast-list")
	}
	if b.useMmap {
		args = append(args, "--use-mmap")
	}
	if b.retries != nil {
		args = append(args, fmt.Sprintf("--retries=%d", *b.retries))
	}
	if b.lowLevelRetries != nil {
		args = append(
			args,
			fmt.Sprintf("--low-level-retries=%d", *b.lowLevelRetries),
		)
	}
	if b.timeout != "" {
		args = append(args, "--timeout="+b.timeout)
	}
	if b.connectTimeout != "" {
		args = append(args, "--contimeout="+b.connectTimeout)
	}
	if b.retryDelay != "" {
		args = append(args, "--retry-delay="+b.retryDelay)
	}
	if b.dryRun {
		args = append(args, "--dry-run")
	}
	if b.checkFirst {
		args = append(args, "--check-first")
	}
	if b.ignoreErrors {
		args = append(args, "--ignore-errors")
	}
	if b.ignoreExisting {
		args = append(args, "--ignore-existing")
	}
	if b.update {
		args = append(args, "--update")
	}
	if b.inplace {
		args = append(args, "--inplace")
	}
	if b.logFile != "" {
		args = append(args, "--log-file="+b.logFile)
	}
	if b.logLevel != "" {
		args = append(args, "--log-level="+b.logLevel)
	}
	if b.stats != "" {
		args = append(args, "--stats="+b.stats)
	}

	args = append(args, b.extraFlags...)

	if b.source != nil {
		path, srcEnvs := b.source.compile("src")
		args = append(args, path)
		for k, v := range srcEnvs {
			envs[k] = v
		}
	}
	if b.destination != nil {
		path, destEnvs := b.destination.compile("dest")
		args = append(args, path)
		for k, v := range destEnvs {
			envs[k] = v
		}
	}

	return args, envs
}

// Builder returns the environment variables map and the fully assembled shell command string.
//
//	envs, cmd := b.Builder()
func (b *Builder) Builder() (map[string]string, string) {
	args, envs := b.build()
	cmd := "rclone " + strings.Join(args, " ")
	return envs, cmd
}
