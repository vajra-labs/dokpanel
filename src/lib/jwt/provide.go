package jwt

import (
	"errors"
	"time"

	"dokpanel/src/conf"
	"dokpanel/src/errx"
	"dokpanel/src/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	iss = "dokpanel"
	aud = "dokpanel-client"
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

// JwtToken provides.
type JwtToken struct {
	secret []byte
}

// New creates a new JwtToken instance.
func New(cfg *conf.Config) *JwtToken {
	return &JwtToken{secret: []byte(cfg.SECRET)}
}

// Payload creates a standard JWT payload for a given subject and token type.
func (j *JwtToken) Payload(sub string, tokenType types.TOKEN) Payload {
	return Payload{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  sub,
			ID:       uuid.NewString(),
			Issuer:   iss,
			Audience: jwt.ClaimStrings{aud},
		},
		TokenType: tokenType,
	}
}

// Sign creates a signed JWT token with the given payload and expiry duration.
func (j *JwtToken) Sign(payload Payload, exp time.Duration) (Token, error) {
	now := time.Now()
	expires := now.Add(exp)

	payload.IssuedAt = jwt.NewNumericDate(now)
	payload.ExpiresAt = jwt.NewNumericDate(expires)

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString(j.secret)
	if err != nil {
		return Token{}, errx.InternalServerError(
			"Failed to sign JWT token", "JWT_SIGN_ERROR",
			errx.WithCause(err),
		)
	}

	return Token{Value: signed, Expires: expires}, nil
}

// Verify parses and validates a JWT token string, returning the claims.
func (j *JwtToken) Verify(tokenStr string) (*Payload, error) {
	var payload Payload

	token, err := jwt.ParseWithClaims(tokenStr, &payload, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errx.UnauthorizedError("Unexpected signing method", "INVALID_JWT")
		}
		return j.secret, nil
	}, jwt.WithIssuer(iss), jwt.WithAudience(aud), jwt.WithExpirationRequired())
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errx.UnauthorizedError(
				"JWT token has expired", "JWT_EXPIRED",
				errx.WithCause(err),
			)
		}
		return nil, errx.UnauthorizedError(
			"Invalid JWT token", "INVALID_JWT",
			errx.WithCause(err),
		)
	}

	if !token.Valid {
		return nil, errx.UnauthorizedError("Invalid JWT token", "INVALID_JWT")
	}

	return &payload, nil
}
