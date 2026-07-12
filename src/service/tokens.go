package service

import (
	"context"
	"strconv"
	"time"

	"dokpanel/src/conf"
	"dokpanel/src/db/repos"
	"dokpanel/src/errx"
	"dokpanel/src/lib/jwt"
	"dokpanel/src/types"
)

// createOptions holds options for creating a token.
type createOptions struct {
	userID    int64
	role      string
	tokenType types.TOKEN
	exp       time.Duration
	save      bool
}

// CreatedToken holds the signed token value and its max-age duration.
type CreatedToken struct {
	Token  jwt.Token
	MaxAge time.Duration
}

// AuthTokens holds the access and refresh tokens returned on login/register.
type AuthTokens struct {
	Access  *CreatedToken
	Refresh *CreatedToken
}

// TokenService handles JWT creation, verification, and blacklisting.
type TokenService struct {
	cfg     *conf.Config
	jwt     *jwt.JwtToken
	queries *repos.Queries
}

// NewTokenService creates a new TokenService instance.
func NewTokenService(cfg *conf.Config, jwtToken *jwt.JwtToken, queries *repos.Queries) *TokenService {
	return &TokenService{cfg: cfg, jwt: jwtToken, queries: queries}
}

// create generates a signed JWT and optionally persists it in the database.
func (s *TokenService) create(ctx context.Context, opts createOptions) (*CreatedToken, error) {
	sub := strconv.FormatInt(opts.userID, 10)
	payload := s.jwt.Payload(sub, opts.tokenType)

	token, err := s.jwt.Sign(payload, opts.exp)
	if err != nil {
		return nil, err
	}

	if opts.save {
		expiredAt := time.Now().Add(opts.exp).Unix()
		_, err = s.queries.CreateJwtToken(ctx, repos.CreateJwtTokenParams{
			Jti:       payload.ID,
			Role:      opts.role,
			UserID:    opts.userID,
			ExpiredAt: &expiredAt,
		})
		if err != nil {
			return nil, errx.InternalServerError(
				"Failed to save token", "TOKEN_SAVE_ERROR",
				errx.WithCause(err),
			)
		}
	}

	return &CreatedToken{Token: token, MaxAge: opts.exp}, nil
}

// Verify parses and validates a JWT string, and checks the token type.
func (s *TokenService) Verify(tokenStr string, tokenType types.TOKEN) (*jwt.Payload, error) {
	payload, err := s.jwt.Verify(tokenStr)
	if err != nil {
		return nil, err
	}
	if payload.TokenType != tokenType {
		return nil, errx.UnauthorizedError("Token type is invalid", "INVALID_TOKEN_TYPE")
	}
	return payload, nil
}

// Generate creates a new access + refresh token pair for the given user.
func (s *TokenService) Generate(ctx context.Context, userID int64, role string) (*AuthTokens, error) {
	access, err := s.create(ctx, createOptions{
		userID:    userID,
		role:      role,
		tokenType: types.ACC_TOKEN,
		exp:       s.cfg.JWT_ACCESS_EXP,
		save:      false,
	})
	if err != nil {
		return nil, err
	}

	refresh, err := s.create(ctx, createOptions{
		userID:    userID,
		role:      role,
		tokenType: types.REF_TOKEN,
		exp:       s.cfg.JWT_REFRESH_EXP,
		save:      true,
	})
	if err != nil {
		return nil, err
	}

	return &AuthTokens{Access: access, Refresh: refresh}, nil
}

// AddBlacklist blacklists a refresh token. If many is true, all tokens for
// that user are blacklisted (e.g. on logout-all / password change).
func (s *TokenService) AddBlacklist(ctx context.Context, tokenStr string, many bool) (string, error) {
	payload, err := s.Verify(tokenStr, types.REF_TOKEN)
	if err != nil {
		return "", err
	}

	record, err := s.queries.GetJwtTokenByJti(ctx, payload.ID)
	if err != nil {
		return "", errx.UnauthorizedError(
			"Token not found", "TOKEN_NOT_FOUND",
			errx.WithCause(err),
		)
	}

	isBlacklisted := record.IsBlacklist != nil && *record.IsBlacklist == 1
	if isBlacklisted {
		return "", errx.BadRequestError("Token is already blacklisted", "TOKEN_ALREADY_BLACKLISTED")
	}

	now := time.Now().Unix()
	blacklisted := int64(1)

	if many {
		userID, err := strconv.ParseInt(payload.Subject, 10, 64)
		if err != nil {
			return "", errx.InternalServerError("Invalid subject in token", "INVALID_TOKEN_SUB")
		}
		err = s.queries.UpdateJwtTokensByUserID(ctx, repos.UpdateJwtTokensByUserIDParams{
			IsBlacklist: &blacklisted,
			BlacklistAt: &now,
			UserID:      userID,
		})
	} else {
		err = s.queries.UpdateJwtTokenByJti(ctx, repos.UpdateJwtTokenByJtiParams{
			IsBlacklist: &blacklisted,
			BlacklistAt: &now,
			Jti:         payload.ID,
		})
	}
	if err != nil {
		return "", errx.InternalServerError(
			"Failed to blacklist token", "TOKEN_BLACKLIST_ERROR",
			errx.WithCause(err),
		)
	}

	return payload.Subject, nil
}

// RefreshAccess issues a new access token using a valid, non-blacklisted refresh token.
func (s *TokenService) RefreshAccess(ctx context.Context, refreshToken string, role string) (*CreatedToken, error) {
	payload, err := s.Verify(refreshToken, types.REF_TOKEN)
	if err != nil {
		return nil, err
	}

	notBlacklisted := int64(0)
	_, err = s.queries.GetJwtTokenByJtiAndBlacklist(ctx, repos.GetJwtTokenByJtiAndBlacklistParams{
		Jti:         payload.ID,
		IsBlacklist: &notBlacklisted,
	})
	if err != nil {
		return nil, errx.UnauthorizedError(
			"Refresh token is blacklisted or invalid", "TOKEN_BLACKLISTED",
			errx.WithCause(err),
		)
	}

	userID, err := strconv.ParseInt(payload.Subject, 10, 64)
	if err != nil {
		return nil, errx.InternalServerError("Invalid subject in token", "INVALID_TOKEN_SUB")
	}

	return s.create(ctx, createOptions{
		userID:    userID,
		role:      role,
		tokenType: types.ACC_TOKEN,
		exp:       s.cfg.JWT_ACCESS_EXP,
		save:      false,
	})
}
