package transit

import (
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
)

type API interface {
	UpdateTransitEngine(engine EngineEntity) error
	ReadTransitEngine(engine core.MountPathEntity) (*Engine, error)
	ListKeys(engine core.MountPathEntity) ([]KeyName, error)
	UpdateTransitKey(engine core.MountPathEntity, key KeyEntity) error
	ReadTransitKey(engine core.MountPathEntity, key KeyNameEntity) (*Key, error)
	DeleteTransitKey(engine core.MountPathEntity, key KeyNameEntity) error
	RotateTransitKey(engine core.MountPathEntity, key KeyNameEntity) error
	TransitEncrypt(engine core.MountPathEntity, key KeyNameEntity, plainText []byte) (string, error)
	TransitDecrypt(engine core.MountPathEntity, key KeyNameEntity, cipherText string) ([]byte, error)
	TransitSign(engine core.MountPathEntity, key KeyNameEntity, input []byte) (string, error)
	TransitVerify(engine core.MountPathEntity, key KeyNameEntity, input []byte, signature string) (bool, error)
}

type EngineEntity interface {
	core.MountPathEntity
	GetPluginName() (string, error)
	GetTransitEngineConfig() (*EngineConfig, error)
}

type EngineConfig struct {
	Cache EngineCacheConfig
}

type EngineCacheConfig struct {
	Size int `json:"size"`
}

type KeyNameEntity interface {
	GetTransitKeyName() (string, error)
}

type KeyEntity interface {
	KeyNameEntity
	GetTransitKeyType() (KeyType, error)
	GetTransitKeyConfig() (*KeyConfig, error)
}

type Key struct {
	Name   string
	Type   KeyType
	Config *KeyConfig
}

type KeyConfig struct {
	MinimumDecryptionVersion int  `json:"min_decryption_version,omitempty"`
	MinimumEncryptionVersion int  `json:"min_encryption_version,omitempty"`
	DeletionAllowed          bool `json:"deletion_allowed,omitempty"`
	Exportable               bool `json:"exportable,omitempty"`
	AllowPlaintextBackup     bool `json:"allow_plaintext_backup,omitempty"`
}

func (t *Key) GetTransitKeyName() (string, error) {
	return t.Name, nil
}

func (t *Key) GetTransitKeyType() (KeyType, error) {
	return t.Type, nil
}

func (t *Key) GetTransitKeyConfig() (*KeyConfig, error) {
	return t.Config, nil
}

// KeyType are defined keys that the vault transit engine supports.
// Details: https://www.vaultproject.io/docs/secrets/transit#key-types.
type KeyType string

const (
	// TypeAes128Gcm96 AES-GCM with a 128-bit AES key and a 96-bit nonce; supports encryption,
	// decryption, key derivation, and convergent encryption.
	TypeAes128Gcm96 KeyType = "aes128-gcm96"
	// TypeAes256Gcm96 AES-GCM with a 256-bit AES key and a 96-bit nonce; supports encryption,
	// decryption, key derivation, and convergent encryption.
	TypeAes256Gcm96 KeyType = "aes256-gcm96"
	// TypeChacha20Poly1305 ChaCha20-Poly1305 with a 256-bit key; supports encryption, decryption,
	// key derivation, and convergent encryption.
	TypeChacha20Poly1305 KeyType = "chacha20-poly1305"
	// TypeED25519 Ed25519; supports signing, signature verification, and key derivation.
	TypeED25519 KeyType = "ed25519"
	// TypeEcdsaP256 ECDSA using curve P-256; supports signing and signature verification.
	TypeEcdsaP256 KeyType = "ecdsa-p256"
	// TypeEcdsaP384 ECDSA using curve P-384; supports signing and signature verification.
	TypeEcdsaP384 KeyType = "ecdsa-p384"
	// TypeEcdsaP521 ECDSA using curve P-521; supports signing and signature verification.
	TypeEcdsaP521 KeyType = "ecdsa-p521"
	// TypeRSA2048 2048-bit RSA key; supports encryption, decryption, signing, and signature verification.
	TypeRSA2048 KeyType = "rsa-2048"
	// TypeRSA3072 3072-bit RSA key; supports encryption, decryption, signing, and signature verification.
	TypeRSA3072 KeyType = "rsa-3072"
	// TypeRSA4096 4096-bit RSA key; supports encryption, decryption, signing, and signature verification.
	TypeRSA4096 KeyType = "rsa-4096"
)

type Engine struct {
	Path       string
	PluginName string
	Config     *EngineConfig
}

func (t *Engine) GetMountPath() (string, error) {
	return t.Path, nil
}

func (t *Engine) GetPluginName() (string, error) {
	return t.PluginName, nil
}

func (t *Engine) GetTransitEngineConfig() (*EngineConfig, error) {
	return t.Config, nil
}

type KeyName string

func (t KeyName) GetTransitKeyName() (string, error) {
	return string(t), nil
}

type transitAPI struct {
	Core  core.API
	Mount mount.API
}

func NewAPI(coreAPI core.API, mountAPI mount.API) API {
	return &transitAPI{
		Core:  coreAPI,
		Mount: mountAPI,
	}
}
