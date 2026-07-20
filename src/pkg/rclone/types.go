package rclone

// Command is the rclone sub-command to run.
type Command string

const (
	Rcat   Command = "rcat"   // Stream stdin to a remote file
	Cat    Command = "cat"    // Concatenate files and output to stdout
	Lsf    Command = "lsf"    // List files in a path in a machine-friendly format
	Copyto Command = "copyto" // Copy source to destination, skipping identical files
	Sync   Command = "sync"   // Make destination identical to source (deletes extra dest files)
	Delete Command = "delete" // Remove files in a path
	Purge  Command = "purge"  // Remove a path and all its contents
	Check  Command = "check"  // Check whether source and destination match
)

// S3Provider identifies the S3-compatible storage provider.
// Pass to S3Target.Provider or use with NewBuilder.
type S3Provider string

const (
	S3AWS          S3Provider = "AWS"
	S3Minio        S3Provider = "Minio"
	S3Cloudflare   S3Provider = "Cloudflare"
	S3DigitalOcean S3Provider = "DigitalOcean"
	S3Alibaba      S3Provider = "Alibaba"
	S3Tencent      S3Provider = "TencentCOS"
	S3Scaleway     S3Provider = "Scaleway"
	S3Wasabi       S3Provider = "Wasabi"
	S3IBMCOS       S3Provider = "IBMCOS"
	S3Linode       S3Provider = "Linode"
	S3Ceph         S3Provider = "Ceph"
	S3Other        S3Provider = "Other"
)

// WebdavVendor identifies the WebDAV server implementation.
// Some vendors have quirks that rclone handles specially.
type WebdavVendor string

const (
	WebdavNextcloud      WebdavVendor = "nextcloud"
	WebdavOwncloud       WebdavVendor = "owncloud"
	WebdavSharepoint     WebdavVendor = "sharepoint"
	WebdavSharepointNTLM WebdavVendor = "sharepoint-ntlm"
	WebdavOther          WebdavVendor = "other"
)

// LogLevel controls rclone's log verbosity.
type LogLevel string

const (
	LogDebug  LogLevel = "DEBUG"
	LogInfo   LogLevel = "INFO"
	LogNotice LogLevel = "NOTICE"
	LogError  LogLevel = "ERROR"
)
