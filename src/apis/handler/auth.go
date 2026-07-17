package handler

import (
	"goploy/src/apis/dtos"
	"goploy/src/conf"
	"goploy/src/core/throw"
	"goploy/src/service"
	"goploy/src/types"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	auth  *service.AuthService
	token *service.TokenService
	cfg   *conf.Config
}

func NewAuthHandler(
	auth *service.AuthService,
	token *service.TokenService,
	cfg *conf.Config,
) *AuthHandler {
	return &AuthHandler{auth, token, cfg}
}

// Setup handles GET /api/auth/setup.
func (h *AuthHandler) Setup(ctx fiber.Ctx) error {
	present, err := h.auth.IsOwnerPresent(ctx.Context())
	if err != nil {
		return err
	}
	return ctx.JSON(dtos.SetupStatusResDto{IsOwnerPresent: present})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(ctx fiber.Ctx) error {
	var body dtos.LoginDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	tokens, err := h.auth.Login(ctx.Context(), &body)
	if err != nil {
		return err
	}
	h.setTokenCookies(ctx, tokens)
	return ctx.JSON(h.toLoginRes(tokens))
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(ctx fiber.Ctx) error {
	var body dtos.RegisterDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	user, err := h.auth.RegisterOwner(ctx.Context(), &body)
	if err != nil {
		return err
	}
	tokens, err := h.token.Generate(ctx.Context(), user.ID)
	if err != nil {
		return err
	}
	h.setTokenCookies(ctx, tokens)
	return ctx.Status(fiber.StatusCreated).JSON(h.toLoginRes(tokens))
}

// Refresh handles POST /api/auth/refresh.
func (h *AuthHandler) Refresh(ctx fiber.Ctx) error {
	refreshToken := ctx.Cookies(string(types.REF_TOKEN))
	if refreshToken == "" {
		return throw.BadRequestError(
			"Refresh token is required",
			"REFRESH_TOKEN_REQUIRED",
		)
	}
	access, err := h.token.RefreshAccess(ctx.Context(), refreshToken)
	if err != nil {
		return err
	}
	ctx.Cookie(&fiber.Cookie{
		Name:     string(types.ACC_TOKEN),
		Value:    access.Value,
		Expires:  access.Expires,
		HTTPOnly: true,
		Secure:   !h.cfg.IS_DEV,
		SameSite: "Lax",
	})
	return ctx.JSON(dtos.TokenDto{
		Token:   access.Value,
		Expires: access.Expires.Unix(),
	})
}

// Logout handles POST /api/auth/logout.
func (h *AuthHandler) Logout(ctx fiber.Ctx) error {
	refreshToken := ctx.Cookies(string(types.REF_TOKEN))
	if refreshToken == "" {
		return throw.BadRequestError(
			"Refresh token is required in cookies",
			"REFRESH_TOKEN_REQUIRED",
		)
	}
	var query dtos.LogoutDto
	if err := ctx.Bind().Query(&query); err != nil {
		return err
	}
	if err := h.token.AddBlacklist(
		ctx.Context(),
		refreshToken,
		query.All,
	); err != nil {
		return err
	}
	// clear both cookies
	ctx.ClearCookie(string(types.ACC_TOKEN))
	ctx.ClearCookie(string(types.REF_TOKEN))
	return ctx.SendStatus(fiber.StatusOK)
}

// setTokenCookies sets access and refresh tokens as HTTP-only cookies.
func (h *AuthHandler) setTokenCookies(
	ctx fiber.Ctx,
	tokens *service.AuthTokens,
) {
	ctx.Cookie(&fiber.Cookie{
		Name:     string(types.ACC_TOKEN),
		Value:    tokens.Access.Value,
		Expires:  tokens.Access.Expires,
		HTTPOnly: true,
		Secure:   !h.cfg.IS_DEV,
		SameSite: "Lax",
	})
	ctx.Cookie(&fiber.Cookie{
		Name:     string(types.REF_TOKEN),
		Value:    tokens.Refresh.Value,
		Expires:  tokens.Refresh.Expires,
		HTTPOnly: true,
		Secure:   !h.cfg.IS_DEV,
		SameSite: "Lax",
	})
}

// toLoginRes converts AuthTokens to LoginRes DTO.
func (h *AuthHandler) toLoginRes(tokens *service.AuthTokens) dtos.LoginRes {
	return dtos.LoginRes{
		Access: dtos.TokenDto{
			Token:   tokens.Access.Value,
			Expires: tokens.Access.Expires.Unix(),
		},
		Refresh: dtos.TokenDto{
			Token:   tokens.Refresh.Value,
			Expires: tokens.Refresh.Expires.Unix(),
		},
	}
}
