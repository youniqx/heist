package mount

import (
	"github.com/youniqx/heist/pkg/vault/core"
)

type API interface {
	HasEngine(engine core.MountPathEntity) (bool, error)
	ListMounts() ([]*Mount, error)
	DeleteEngine(engine core.MountPathEntity) error
	MountEngine(engine Entity) error
	ReadMount(engine core.MountPathEntity) (*Mount, error)
	ReloadPluginBackends(plugin Plugin) error
	TuneEngine(engine core.MountPathEntity, config *TuneConfig) error
	ReadTuneConfig(engine core.MountPathEntity) (*TuneConfig, error)
}

type Type string

const (
	// TypeKVV2 configures the secret engine to have the type kv-v2.
	TypeKVV2 Type = "kv-v2"
	// TypeKVV1 configures the secret engine to have the type kv.
	TypeKVV1 Type = "kv"
	// TypeTransit configures the secret engine to have the type transit.
	TypeTransit Type = "transit"
	// TypePKI configures the secret engine to have the type pki.
	TypePKI Type = "pki"
)

type Visibility string

const (
	// VisibilityHidden specifies to hide this mount
	// in the UI-specific listing endpoint.
	// Behaves like "hidden" if not set.
	VisibilityHidden Visibility = "hidden"
	// VisibilityUnauth specifies to show this mount
	// in the UI-specific listing endpoint.
	VisibilityUnauth Visibility = "unauth"
)

type TuneConfig struct {
	DefaultLeaseTTL           *core.VaultTTL `json:"default_lease_ttl,omitempty"`
	MaxLeaseTTL               *core.VaultTTL `json:"max_lease_ttl,omitempty"`
	Description               string         `json:"description,omitempty"`
	AuditNonHmacRequestKeys   []string       `json:"audit_non_hmac_request_keys,omitempty"`
	AuditNonHmacResponseKeys  []string       `json:"audit_non_hmac_response_keys,omitempty"`
	ListingVisiblity          Visibility     `json:"listing_visiblity,omitempty"`
	PassthroughRequestHeaders []string       `json:"passthrough_request_headers,omitempty"`
	AllowedResponseHeaders    []string       `json:"allowed_response_headers,omitempty"`
}

type Entity interface {
	core.MountPathEntity
	GetMountType() (Type, error)
	GetMountConfig() (*TuneConfig, error)
	GetMountOptions() (map[string]string, error)
}

type Plugin string

const (
	// PluginTransit defines the name of the transit plugin.
	PluginTransit = "transit"
)

type Mount struct {
	Path    string
	Type    Type
	Options map[string]string
	Config  *TuneConfig
}

func (e *Mount) GetMountOptions() (map[string]string, error) {
	return e.Options, nil
}

func (e *Mount) GetMountPath() (string, error) {
	return e.Path, nil
}

func (e *Mount) GetMountType() (Type, error) {
	return e.Type, nil
}

func (e *Mount) GetMountConfig() (*TuneConfig, error) {
	return e.Config, nil
}

type mountAPI struct {
	Core core.API
}

func NewAPI(core core.API) API {
	return &mountAPI{Core: core}
}
