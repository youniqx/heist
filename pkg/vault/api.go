package vault

import (
	"github.com/youniqx/heist/pkg/vault/auth"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kubernetesauth"
	"github.com/youniqx/heist/pkg/vault/kvengine"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
	"github.com/youniqx/heist/pkg/vault/mount"
	"github.com/youniqx/heist/pkg/vault/pki"
	"github.com/youniqx/heist/pkg/vault/policy"
	"github.com/youniqx/heist/pkg/vault/random"
	"github.com/youniqx/heist/pkg/vault/transit"
)

type API interface {
	kvsecret.API
	kvengine.API
	transit.API
	policy.API
	mount.API
	random.API
	auth.API
	kubernetesauth.API
	pki.API
	GetAddress() string
	GetCACerts() []string
}

type (
	kvSecretAPI       = kvsecret.API
	kvEngineAPI       = kvengine.API
	transitAPI        = transit.API
	policyAPI         = policy.API
	mountAPI          = mount.API
	randomAPI         = random.API
	authAPI           = auth.API
	pkiAPI            = pki.API
	kubernetesAuthAPI = kubernetesauth.API
)

type vaultAPI struct {
	kvSecretAPI
	kvEngineAPI
	transitAPI
	policyAPI
	mountAPI
	randomAPI
	authAPI
	kubernetesAuthAPI
	pkiAPI

	Address string
	CACerts []string
}

func (v *vaultAPI) GetCACerts() []string {
	return v.CACerts
}

func (v *vaultAPI) GetAddress() string {
	return v.Address
}

type Builder interface {
	WithAddressFrom(source core.StringSource) Builder
	WithTokenFrom(source core.StringSource) Builder
	WithCAsFrom(source ...core.StringSource) Builder
	WithAuthProvider(provider core.AuthProvider) Builder
	Complete() (API, error)
}

func NewAPI() Builder {
	return &builder{}
}

type builder struct {
	Address      core.StringSource
	AuthOption   authOptionFactory
	CACertOption func() ([]string, error)
}

type authOptionFactory func() (core.Option, error)

func (b *builder) WithAddressFrom(source core.StringSource) Builder {
	b.Address = source
	return b
}

func (b *builder) WithTokenFrom(source core.StringSource) Builder {
	b.AuthOption = func() (core.Option, error) {
		token, err := source.FetchStringValue()
		if err != nil {
			return nil, core.ErrAPIError.WithDetails("failed to fetch Vault token").WithCause(err)
		}

		return core.WithToken(token), nil
	}

	return b
}

func (b *builder) WithCAsFrom(sources ...core.StringSource) Builder {
	b.CACertOption = func() ([]string, error) {
		caCerts := make([]string, 0, len(sources))
		for _, source := range sources {
			ca, err := source.FetchStringValue()
			if err != nil {
				return nil, core.ErrAPIError.WithDetails("failed to fetch ca").WithCause(err)
			}

			caCerts = append(caCerts, ca)
		}
		return caCerts, nil
	}
	return b
}

func (b *builder) WithAuthProvider(provider core.AuthProvider) Builder {
	b.AuthOption = func() (core.Option, error) {
		return core.WithAuthProvider(provider), nil
	}

	return b
}

// DefaultKubernetesTokenPath is the path to the default mount point for kubernetes service account tokens.
// #nosec G101
const DefaultKubernetesTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

func (b *builder) Complete() (API, error) {
	address, err := b.Address.FetchStringValue()
	if err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to fetch Vault address").WithCause(err)
	}

	authOption, err := b.AuthOption()
	if err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to fetch Vault auth option").WithCause(err)
	}

	var caCerts []string
	if b.CACertOption != nil {
		caCerts, err = b.CACertOption()
		if err != nil {
			return nil, core.ErrAPIError.WithDetails("failed to fetch Vault ca cert option").WithCause(err)
		}
	}

	coreAPI, err := core.NewCoreAPI(address, core.WithCACerts(caCerts...), authOption)
	if err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to create core api instance").WithCause(err)
	}

	mountAPI := mount.NewAPI(coreAPI)
	authAPI := auth.NewAPI(coreAPI)

	return &vaultAPI{
		kvSecretAPI:       kvsecret.NewAPI(coreAPI),
		kvEngineAPI:       kvengine.NewAPI(coreAPI, mountAPI),
		policyAPI:         policy.NewAPI(coreAPI),
		transitAPI:        transit.NewAPI(coreAPI, mountAPI),
		randomAPI:         random.NewAPI(coreAPI),
		kubernetesAuthAPI: kubernetesauth.NewAPI(coreAPI, authAPI),
		authAPI:           authAPI,
		mountAPI:          mountAPI,
		pkiAPI:            pki.NewAPI(coreAPI, mountAPI),
		Address:           address,
		CACerts:           caCerts,
	}, nil
}
