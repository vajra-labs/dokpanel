package dtos

type CreateSshKeyDto struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255" doc:"Name of the SSH Key"`
	Description *string `json:"description" validate:"omitempty,max=1000"     doc:"Optional description"`
	PrivateKey  string  `json:"privateKey"  validate:"required"               doc:"SSH Private Key content"`
	PublicKey   string  `json:"publicKey"   validate:"required"               doc:"SSH Public Key content"`
}

type UpdateSshKeyDto struct {
	Name        string  `json:"name"        validate:"required,min=1,max=255" doc:"Name of the SSH Key"`
	Description *string `json:"description" validate:"omitempty,max=1000"     doc:"Optional description"`
	PrivateKey  string  `json:"privateKey"  validate:"required"               doc:"SSH Private Key content"`
	PublicKey   string  `json:"publicKey"   validate:"required"               doc:"SSH Public Key content"`
}

type SshKeyResDto struct {
	ID          int64   `json:"id"          doc:"SSH Key ID"`
	Name        string  `json:"name"        doc:"Name of the SSH Key"`
	Description *string `json:"description" doc:"Optional description"`
	PrivateKey  string  `json:"privateKey"  doc:"SSH Private Key content"`
	PublicKey   string  `json:"publicKey"   doc:"SSH Public Key content"`
	LastUsedAt  *int64  `json:"lastUsedAt"  doc:"Unix timestamp when the key was last used"`
	CreatedAt   int64   `json:"createdAt"   doc:"Unix timestamp when the key was created"`
	UpdatedAt   int64   `json:"updatedAt"   doc:"Unix timestamp when the key was updated"`
}

type SshKeyMiniResDto struct {
	ID   int64  `json:"id"   doc:"SSH Key ID"`
	Name string `json:"name" doc:"Name of the SSH Key"`
}

type GenSshKeyDto struct {
	Type string `json:"type" validate:"required,oneof=rsa ed25519" doc:"SSH Key type" enums:"rsa,ed25519"`
}

type GenSshKeyResDto struct {
	PrivateKey string `json:"privateKey" doc:"Generated SSH Private Key"`
	PublicKey  string `json:"publicKey"  doc:"Generated SSH Public Key (authorized_keys format)"`
}
