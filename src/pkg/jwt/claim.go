package jwt

import (
	"time"

	"goploy/src/types"

	"github.com/golang-jwt/jwt/v5"
)

// Payload holds standard JWT claims.
type Payload struct {
	jwt.RegisteredClaims
	TokenType types.TOKEN `json:"token_type"`
}

// Token holds a signed JWT string and its expiry time.
type Token struct {
	Value   string
	Expires time.Time
}
