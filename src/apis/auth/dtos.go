package auth

// RegisterDto is the request body for user registration.
type RegisterDto struct {
	Email     string `json:"email" validate:"required,email" doc:"User email address"`
	FirstName string `json:"firstName" validate:"required,min=2,max=50" doc:"User first name"`
	LastName  string `json:"lastName" validate:"required,min=2,max=50" doc:"User last name"`
	Password  string `json:"password" validate:"required,min=8" doc:"User password"`
}

// LoginDto is the request body for user login.
type LoginDto struct {
	Email    string `json:"email" validate:"required,email" doc:"User email address"`
	Password string `json:"password" validate:"required" doc:"User password"`
}

// TokenDto holds a JWT token and its expiry.
type TokenDto struct {
	Token   string `json:"token" doc:"JWT token"`
	Expires int64  `json:"expires" doc:"Token expiration time in Unix seconds"`
}

// LoginRes is the response body for user login.
type LoginRes struct {
	Access  TokenDto `json:"access" doc:"Access token details"`
	Refresh TokenDto `json:"refresh" doc:"Refresh token details"`
}
