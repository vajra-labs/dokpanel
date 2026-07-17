package handler

import (
	"strconv"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/db/repos"
	"goploy/src/service"

	"github.com/gofiber/fiber/v3"
)

type SshKeyHandler struct {
	sshKey *service.SshKeyService
}

func NewSshKeyHandler(sshKey *service.SshKeyService) *SshKeyHandler {
	return &SshKeyHandler{sshKey: sshKey}
}

// Create handles POST /api/ssh-keys.
func (h *SshKeyHandler) Create(ctx fiber.Ctx) error {
	var body dtos.CreateSshKeyDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	res, err := h.sshKey.CreateSshKey(ctx.Context(), &body)
	if err != nil {
		return err
	}
	return ctx.Status(fiber.StatusCreated).JSON(mapSshKeyResponse(res))
}

// List handles GET /api/ssh-keys.
func (h *SshKeyHandler) List(ctx fiber.Ctx) error {
	keys, err := h.sshKey.ListSshKeys(ctx.Context())
	if err != nil {
		return err
	}
	res := make([]dtos.SshKeyResDto, len(keys))
	for i, key := range keys {
		res[i] = *mapSshKeyResponse(&key)
	}
	return ctx.JSON(res)
}

// ListMini handles GET /api/ssh-keys/mini (allForApps).
func (h *SshKeyHandler) ListMini(ctx fiber.Ctx) error {
	keys, err := h.sshKey.ListSshKeys(ctx.Context())
	if err != nil {
		return err
	}
	res := make([]dtos.SshKeyMiniResDto, len(keys))
	for i, key := range keys {
		res[i] = dtos.SshKeyMiniResDto{
			ID:   key.ID,
			Name: key.Name,
		}
	}
	return ctx.JSON(res)
}

// Generate handles POST /api/ssh-keys/generate.
func (h *SshKeyHandler) Generate(ctx fiber.Ctx) error {
	var body dtos.GenSshKeyDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	privateKey, publicKey, err := h.sshKey.GenSshKeyPair(body.Type)
	if err != nil {
		return err
	}
	return ctx.JSON(dtos.GenSshKeyResDto{
		PrivateKey: *privateKey,
		PublicKey:  *publicKey,
	})
}

// Get handles GET /api/ssh-keys/:id.
func (h *SshKeyHandler) Get(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError("Invalid SSH Key ID", "INVALID_SSH_KEY_ID")
	}
	res, err := h.sshKey.GetSshKeyByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(mapSshKeyResponse(res))
}

// Update handles PUT /api/ssh-keys/:id.
func (h *SshKeyHandler) Update(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError("Invalid SSH Key ID", "INVALID_SSH_KEY_ID")
	}
	var body dtos.UpdateSshKeyDto
	if err := ctx.Bind().Body(&body); err != nil {
		return err
	}
	res, err := h.sshKey.UpdateSshKey(ctx.Context(), id, &body)
	if err != nil {
		return err
	}
	return ctx.JSON(mapSshKeyResponse(res))
}

// Delete handles DELETE /api/ssh-keys/:id.
func (h *SshKeyHandler) Delete(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return throw.BadRequestError("Invalid SSH Key ID", "INVALID_SSH_KEY_ID")
	}
	res, err := h.sshKey.DeleteSshKey(ctx.Context(), id)
	if err != nil {
		return err
	}
	return ctx.JSON(mapSshKeyResponse(res))
}

// mapSshKeyResponse maps database SSHKey model to DTO response.
func mapSshKeyResponse(key *repos.SshKey) *dtos.SshKeyResDto {
	return &dtos.SshKeyResDto{
		ID:          key.ID,
		Name:        key.Name,
		Description: key.Description,
		PrivateKey:  key.PrivateKey,
		PublicKey:   key.PublicKey,
		LastUsedAt:  key.LastUsedAt,
		CreatedAt:   key.CreatedAt,
		UpdatedAt:   key.UpdatedAt,
	}
}
