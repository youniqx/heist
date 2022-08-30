package pki

import (
	"reflect"

	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
)

type EngineEntity interface {
	core.MountPathEntity
	GetPluginName() (string, error)
	GetPKIEngineConfig() (*mount.TuneConfig, error)
}

type Engine struct {
	Path       string
	PluginName string
	Config     *mount.TuneConfig
}

func (e *Engine) GetMountPath() (string, error) {
	return e.Path, nil
}

func (e *Engine) GetPluginName() (string, error) {
	return e.PluginName, nil
}

func (e *Engine) GetPKIEngineConfig() (*mount.TuneConfig, error) {
	return e.Config, nil
}

func (p *pkiAPI) ReadPKIEngine(engine core.MountPathEntity) (*Engine, error) {
	panic("implement me")
}

func (p *pkiAPI) UpdatePKIEngine(engine EngineEntity) error {
	log := p.Core.Log().WithValues("method", "UpdatePKIEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		log.Info("failed to get engine path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get pki engine path").WithCause(err)
	}

	log = log.WithValues("path", path)

	exists, err := p.Mount.HasEngine(engine)
	if err != nil {
		log.Info("failed to check if engine exists", "error", err)
		return core.ErrAPIError.WithDetails("failed to check if pki engine exists").WithCause(err)
	}

	log = log.WithValues("exists", exists)

	desiredConfig, err := engine.GetPKIEngineConfig()
	if err != nil {
		log.Info("failed to get desired pki mount config", "error", err)
		return core.ErrAPIError.WithDetails("failed to get desired pki mount config").WithCause(err)
	}

	pluginName, err := engine.GetPluginName()
	if err != nil {
		log.Info("failed to fetch plugin name", "error", err)
		return core.ErrAPIError.WithDetails("failed to fetch plugin name").WithCause(err)
	}

	if pluginName == "" {
		pluginName = string(mount.TypePKI)
	}

	if !exists {
		mountRequest := &mount.Mount{
			Path:   path,
			Type:   mount.Type(pluginName),
			Config: desiredConfig,
		}

		log.Info("creating new pki engine")

		if err := p.Mount.MountEngine(mountRequest); err != nil {
			log.Info("failed to create pki engine", "error", err)
			return core.ErrAPIError.WithDetails("failed to create pki engine").WithCause(err)
		}
	}

	log.Info("updating pki engine")

	if err := p.tunePKIEngine(engine); err != nil {
		log.Info("failed to update pki engine config", "error", err)
		return core.ErrAPIError.WithDetails("failed to update pki engine config").WithCause(err)
	}

	return nil
}

func (p *pkiAPI) tunePKIEngine(entity EngineEntity) error {
	log := p.Core.Log().WithValues("method", "tunePKIEngine")

	desiredConfig, err := entity.GetPKIEngineConfig()
	if err != nil {
		log.Info("failed to get desired pki mount config", "error", err)
		return core.ErrAPIError.WithDetails("failed to get desired pki mount config").WithCause(err)
	}

	currentConfig, err := p.Mount.ReadTuneConfig(entity)
	if err != nil {
		log.Info("failed to read pki tune mount config", "error", err)
		return core.ErrAPIError.WithDetails("failed to read pki tune mount config").WithCause(err)
	}

	if reflect.DeepEqual(desiredConfig, currentConfig) {
		return nil
	}

	if err := p.Mount.TuneEngine(entity, desiredConfig); err != nil {
		log.Info("failed to write desired pki tune configuration", "error", err)
		return core.ErrAPIError.WithDetails("failed to write desired pki tune configuration").WithCause(err)
	}

	return nil
}
