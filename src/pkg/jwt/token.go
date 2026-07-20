package jwt

import (
	"errors"
	"time"

	"goploy/src/conf"
	"goploy/src/core/throw"
	"goploy/src/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	iss = "goploy"
	aud = "goploy-client"
)

// JwtToken provides.
type JwtToken struct {
	secret []byte
}

// New creates a new JwtToken instance.
func NewJwtToken(cfg *conf.Config) *JwtToken {
	return &JwtToken{secret: []byte(cfg.SECRET)}
}

// Payload creates a standard JWT payload for a given subject and token type.
func (j *JwtToken) Payload(sub string, tokenType types.TOKEN) Payload {
	return Payload{
		Subject:   sub,
		ID:        uuid.NewString(),
		Issuer:    iss,
		Audience:  jwt.ClaimStrings{aud},
		TokenType: tokenType,
	}
}

// Sign creates a signed JWT token with the given payload and expiry duration.
func (j *JwtToken) Sign(payload Payload, exp time.Duration) (Token, error) {
	now := time.Now()
	expires := now.Add(exp)

	payload.IssuedAt = jwt.NewNumericDate(now)
	payload.ExpiresAt = jwt.NewNumericDate(expires)

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).
		SignedString(j.secret)
	if err != nil {
		return Token{}, throw.InternalServerError(
			"Failed to sign JWT token", "JWT_SIGN_ERROR",
			throw.WithCause(err),
		)
	}

	return Token{Value: signed, Expires: expires}, nil
}

// Verify parses and validates a JWT token string, returning the claims.
func (j *JwtToken) Verify(tokenStr string) (*Payload, error) {
	var payload Payload

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&payload,
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, throw.UnauthorizedError(
					"Unexpected signing method",
					"INVALID_JWT",
				)
			}
			return j.secret, nil
		},
		jwt.WithIssuer(iss),
		jwt.WithAudience(aud),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, throw.UnauthorizedError(
				"JWT token has expired", "JWT_EXPIRED",
				throw.WithCause(err),
			)
		}
		return nil, throw.UnauthorizedError(
			"Invalid JWT token", "INVALID_JWT",
			throw.WithCause(err),
		)
	}

	if !token.Valid {
		return nil, throw.UnauthorizedError("Invalid JWT token", "INVALID_JWT")
	}

	return &payload, nil
}
