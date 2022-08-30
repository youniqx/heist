package operator

import (
	"errors"
	"strconv"

	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/vault"
	"k8s.io/client-go/rest"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var log = controllerruntime.Log.WithName("setup-manager")

type Component interface {
	Register(api vault.API, mgr manager.Manager) error
}

type Builder interface {
	WithVaultAPI(api vault.API) Builder
	WithRestConfig(config *rest.Config) Builder
	WithOptions(options controllerruntime.Options) Builder
	Register(component ...Component) Builder
	Complete() (manager.Manager, error)
}

type builder struct {
	API        vault.API
	Config     *rest.Config
	Options    controllerruntime.Options
	Components []Component
}

func (b *builder) Register(components ...Component) Builder {
	b.Components = append(b.Components, components...)
	return b
}

func (b *builder) WithVaultAPI(api vault.API) Builder {
	b.API = api
	return b
}

func (b *builder) WithRestConfig(config *rest.Config) Builder {
	b.Config = config
	return b
}

func (b *builder) WithOptions(options controllerruntime.Options) Builder {
	b.Options = options
	return b
}

func (b *builder) Complete() (manager.Manager, error) {
	if err := managed.UpdateManagedTransitEngine(b.API); err != nil {
		log.Error(err, "failed to updated managed components for operator")
		return nil, err
	}

	if b.API == nil {
		return nil, errors.New("no Vault API set")
	}

	if b.Config == nil {
		return nil, errors.New("no Rest Config set")
	}

	mgr, err := b.createManager()
	if err != nil {
		log.Error(err, "unable to create manager")
		return nil, err
	}

	for i, component := range b.Components {
		if err := component.Register(b.API, mgr); err != nil {
			log.Error(err, "unable to register component "+strconv.Itoa(i))
			return nil, err
		}
	}

	return mgr, nil
}

func Create() Builder {
	return &builder{}
}

func (b *builder) createManager() (manager.Manager, error) {
	mgr, err := controllerruntime.NewManager(b.Config, b.Options)
	if err != nil {
		log.Error(err, "unable to start manager")
		return nil, err
	}

	if err := mgr.AddHealthzCheck("health", healthz.Ping); err != nil {
		log.Error(err, "unable to set up health check")
		return nil, err
	}
	if err := mgr.AddReadyzCheck("check", healthz.Ping); err != nil {
		log.Error(err, "unable to set up ready check")
		return nil, err
	}

	return mgr, nil
}
