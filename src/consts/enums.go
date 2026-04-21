package consts

// Declare Types
type TOKEN string

const (
	// Env
	DEV  string = "dev"
	PROD string = "prod"
	TEST string = "test"

	// JWT
	ACC_TOKEN TOKEN = "access_token"
	REF_TOKEN TOKEN = "refresh_token"
)
