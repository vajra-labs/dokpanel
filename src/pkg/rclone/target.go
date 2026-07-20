package rclone

import (
	"fmt"
	"strings"
)

// Target represents a rclone source or destination.
// Each variant compiles itself into a path string + RCLONE_CONFIG_* env vars
// so no rclone config file is needed.
type Target interface {
	// compile returns the rclone path string (e.g. "src_sftp:/remote/path")
	// and the RCLONE_CONFIG_* env vars needed to authenticate it.
	// prefix is either "src" or "dest" — used to namespace the env vars.
	compile(prefix string) (path string, envs map[string]string)
}

// LocalTarget is a path on the local filesystem. No env vars are needed.
type LocalTarget struct {
	Path string
}

func (t *LocalTarget) compile(_ string) (string, map[string]string) {
	return t.Path, nil
}

// S3Target holds credentials and config for an S3-compatible storage backend.
type S3Target struct {
	Provider        S3Provider // e.g. S3AWS, S3Minio, S3Cloudflare
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Region          string
	Endpoint        string // Leave empty for AWS
	Path            string // Path inside the bucket
	ForcePathStyle  bool   // Required for MinIO and some S3-compatible services
	NoCheckBucket   bool   // Skip bucket existence check (useful for restricted IAM)
}

func (t *S3Target) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "s3")
	c.set("TYPE", "s3")
	c.set("PROVIDER", string(t.Provider))
	c.set("ACCESS_KEY_ID", t.AccessKeyID)
	c.set("SECRET_ACCESS_KEY", t.SecretAccessKey)
	c.set("REGION", t.Region)
	c.set("ENDPOINT", t.Endpoint)
	if t.ForcePathStyle {
		c.set("FORCE_PATH_STYLE", "true")
	}
	if t.NoCheckBucket {
		c.set("NO_CHECK_BUCKET", "true")
	}
	return c.bucketPath(t.Bucket, t.Path), c.envs
}

// SftpAuth is the authentication method for SFTP.
type SftpAuth struct {
	Pass        *string // password (will be obscured via rclone obscure)
	KeyFile     *string // path to private key file
	KeyUseAgent bool    // use SSH agent
}

// SftpPassword creates password-based SFTP auth.
func SftpPassword(pass string) SftpAuth { return SftpAuth{Pass: &pass} }

// SftpKeyFile creates key-file-based SFTP auth.
func SftpKeyFile(keyFile string) SftpAuth { return SftpAuth{KeyFile: &keyFile} }

// SftpAgent creates SSH-agent-based SFTP auth.
func SftpAgent() SftpAuth { return SftpAuth{KeyUseAgent: true} }

// SftpTarget holds credentials for an SFTP backend.
type SftpTarget struct {
	Host string
	Port uint16 // 0 means use default (22)
	User string
	Auth SftpAuth
	Path string
}

func (t *SftpTarget) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "sftp")
	c.set("TYPE", "sftp")
	c.set("HOST", t.Host)
	c.set("USER", t.User)
	if t.Port != 0 {
		c.set("PORT", fmt.Sprintf("%d", t.Port))
	}
	if t.Auth.Pass != nil {
		c.set("PASS", obscurePassword(*t.Auth.Pass))
	}
	if t.Auth.KeyFile != nil {
		c.set("KEY_FILE", *t.Auth.KeyFile)
	}
	if t.Auth.KeyUseAgent {
		c.set("KEY_USE_AGENT", "true")
	}
	return c.path(t.Path), c.envs
}

// FtpTarget holds credentials for an FTP/FTPS backend.
type FtpTarget struct {
	Host string
	Port uint16 // 0 means use default (21)
	User string
	Pass string
	Path string
	TLS  bool // enable explicit FTPS
}

func (t *FtpTarget) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "ftp")
	c.set("TYPE", "ftp")
	c.set("HOST", t.Host)
	c.set("USER", t.User)
	c.set("PASS", obscurePassword(t.Pass))
	if t.Port != 0 {
		c.set("PORT", fmt.Sprintf("%d", t.Port))
	}
	if t.TLS {
		c.set("TLS", "true")
	}
	return c.path(t.Path), c.envs
}

// B2Target holds credentials for Backblaze B2.
type B2Target struct {
	AccountID      string
	ApplicationKey string
	Bucket         string
	Path           string
}

func (t *B2Target) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "b2")
	c.set("TYPE", "b2")
	c.set("ACCOUNT", t.AccountID)
	c.set("KEY", t.ApplicationKey)
	return c.bucketPath(t.Bucket, t.Path), c.envs
}

// GcsTarget holds credentials for Google Cloud Storage.
type GcsTarget struct {
	ServiceAccountCredentials string // JSON key content or file path
	Bucket                    string
	Path                      string
	ProjectNumber             string // optional
}

func (t *GcsTarget) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "gcs")
	c.set("TYPE", "google cloud storage")
	c.set("SERVICE_ACCOUNT_CREDENTIALS", t.ServiceAccountCredentials)
	c.set("PROJECT_NUMBER", t.ProjectNumber)
	return c.bucketPath(t.Bucket, t.Path), c.envs
}

// AzureBlobTarget holds credentials for Azure Blob Storage.
type AzureBlobTarget struct {
	Account   string
	Key       string
	Container string
	Path      string
}

func (t *AzureBlobTarget) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "azureblob")
	c.set("TYPE", "azureblob")
	c.set("ACCOUNT", t.Account)
	c.set("KEY", t.Key)
	return c.bucketPath(t.Container, t.Path), c.envs
}

// WebdavTarget holds credentials for a WebDAV server.
type WebdavTarget struct {
	URL    string
	User   string
	Pass   string
	Vendor WebdavVendor // optional: e.g. WebdavNextcloud, WebdavOwncloud
	Path   string
}

func (t *WebdavTarget) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "webdav")
	c.set("TYPE", "webdav")
	c.set("URL", t.URL)
	c.set("USER", t.User)
	c.set("PASS", obscurePassword(t.Pass))
	c.set("VENDOR", string(t.Vendor))
	return c.path(t.Path), c.envs
}

// DropboxTarget holds credentials for Dropbox.
type DropboxTarget struct {
	Token string
	Path  string
}

func (t *DropboxTarget) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "dropbox")
	c.set("TYPE", "dropbox")
	c.set("TOKEN", t.Token)
	return c.path(t.Path), c.envs
}

// GoogleDriveTarget holds credentials for Google Drive.
type GoogleDriveTarget struct {
	Token        string
	ClientID     string // optional
	ClientSecret string // optional
	Path         string
}

func (t *GoogleDriveTarget) compile(prefix string) (string, map[string]string) {
	c := newConfigBuilder(prefix, "drive")
	c.set("TYPE", "drive")
	c.set("TOKEN", t.Token)
	c.set("CLIENT_ID", t.ClientID)
	c.set("CLIENT_SECRET", t.ClientSecret)
	return c.path(t.Path), c.envs
}

type configBuilder struct {
	name string
	envs map[string]string
}

func newConfigBuilder(prefix, backend string) *configBuilder {
	return &configBuilder{
		name: prefix + "_" + backend,
		envs: make(map[string]string),
	}
}

func (c *configBuilder) set(key, val string) {
	if val != "" {
		k := fmt.Sprintf(
			"RCLONE_CONFIG_%s_%s",
			strings.ToUpper(c.name),
			strings.ToUpper(key),
		)
		c.envs[k] = val
	}
}

func (c *configBuilder) path(path string) string {
	return fmt.Sprintf("%s:%s", c.name, path)
}

func (c *configBuilder) bucketPath(bucket, path string) string {
	return fmt.Sprintf(
		"%s:%s/%s",
		c.name,
		bucket,
		strings.TrimPrefix(path, "/"),
	)
}
