package service

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"

	"goploy/src/apis/dtos"
	"goploy/src/core/throw"
	"goploy/src/db/repos"

	"golang.org/x/crypto/ssh"
)

type SshKeyService struct {
	queries *repos.Queries
}

func NewSshKeyService(queries *repos.Queries) *SshKeyService {
	return &SshKeyService{queries}
}

// CreateSshKey inserts a new SSH Key and returns it.
func (s *SshKeyService) CreateSshKey(
	ctx context.Context,
	dto *dtos.CreateSshKeyDto,
) (*repos.SshKey, error) {
	key, err := s.queries.CreateSSHKey(ctx, repos.CreateSSHKeyParams{
		Name:        dto.Name,
		Description: dto.Description,
		PrivateKey:  dto.PrivateKey,
		PublicKey:   dto.PublicKey,
		LastUsedAt:  nil,
	})
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to create SSH Key",
			"SSH_KEY_CREATE_ERROR",
			throw.WithCause(err),
		)
	}
	return &key, nil
}

// DeleteSshKey deletes the SSH Key by ID and returns the deleted key.
func (s *SshKeyService) DeleteSshKey(
	ctx context.Context,
	id int64,
) (*repos.SshKey, error) {
	key, err := s.queries.DeleteSSHKey(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"SSH Key not found",
				"SSH_KEY_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to delete SSH Key",
			"SSH_KEY_DELETE_ERROR",
			throw.WithCause(err),
		)
	}
	return &key, nil
}

// UpdateSshKey updates SSH Key fields.
func (s *SshKeyService) UpdateSshKey(
	ctx context.Context,
	id int64,
	dto *dtos.UpdateSshKeyDto,
) (*repos.SshKey, error) {
	key, err := s.queries.UpdateSSHKey(ctx, repos.UpdateSSHKeyParams{
		ID:          id,
		Name:        dto.Name,
		Description: dto.Description,
		PrivateKey:  dto.PrivateKey,
		PublicKey:   dto.PublicKey,
		LastUsedAt:  nil,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"SSH Key not found",
				"SSH_KEY_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to update SSH Key",
			"SSH_KEY_UPDATE_ERROR",
			throw.WithCause(err),
		)
	}
	return &key, nil
}

// GetSshKeyByID fetches an SSH Key by ID.
func (s *SshKeyService) GetSshKeyByID(
	ctx context.Context,
	id int64,
) (*repos.SshKey, error) {
	key, err := s.queries.GetSSHKeyByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, throw.NotFoundError(
				"SSH Key not found",
				"SSH_KEY_NOT_FOUND",
			)
		}
		return nil, throw.InternalServerError(
			"Failed to fetch SSH Key",
			"SSH_KEY_FETCH_ERROR",
			throw.WithCause(err),
		)
	}
	return &key, nil
}

// ListSshKeys fetches all SSH Keys.
func (s *SshKeyService) ListSshKeys(
	ctx context.Context,
) ([]repos.SshKey, error) {
	keys, err := s.queries.ListSSHKeys(ctx)
	if err != nil {
		return nil, throw.InternalServerError(
			"Failed to list SSH Keys",
			"SSH_KEYS_LIST_ERROR",
			throw.WithCause(err),
		)
	}
	return keys, nil
}

// GenSshKeyPair generates a new secure private and public key pair based on type (rsa or ed25519).
func (s *SshKeyService) GenSshKeyPair(
	keyType string,
) (*string, *string, error) {
	if keyType == "ed25519" {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, throw.InternalServerError(
				"Failed to generate ED25519 private key",
				"ED25519_GEN_ERROR",
				throw.WithCause(err),
			)
		}
		privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
		if err != nil {
			return nil, nil, throw.InternalServerError(
				"Failed to marshal private key",
				"PRIV_KEY_MARSHAL_ERROR",
				throw.WithCause(err),
			)
		}
		privateKeyPEM := &pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privBytes,
		}
		privateKeyStr := string(pem.EncodeToMemory(privateKeyPEM))
		publicKeySSH, err := ssh.NewPublicKey(pubKey)
		if err != nil {
			return nil, nil, throw.InternalServerError(
				"Failed to format SSH public key",
				"SSH_PUB_KEY_FORMAT_ERROR",
				throw.WithCause(err),
			)
		}
		publicKeyStr := string(ssh.MarshalAuthorizedKey(publicKeySSH))
		return &privateKeyStr, &publicKeyStr, nil
	}
	// Default to RSA
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, throw.InternalServerError(
			"Failed to generate RSA private key",
			"RSA_GEN_ERROR",
			throw.WithCause(err),
		)
	}
	// PEM encode Private Key
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyStr := string(pem.EncodeToMemory(privateKeyPEM))
	// Marshal Public Key for SSH
	publicKeySSH, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, nil, throw.InternalServerError(
			"Failed to format SSH public key",
			"SSH_PUB_KEY_FORMAT_ERROR",
			throw.WithCause(err),
		)
	}
	publicKeyStr := string(ssh.MarshalAuthorizedKey(publicKeySSH))
	return &privateKeyStr, &publicKeyStr, nil
}
