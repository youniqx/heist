package v1alpha1

import (
	"fmt"

	"github.com/youniqx/heist/pkg/vault/transit"
)

func (r *VaultTransitEngine) GetMountPath() (string, error) {
	return fmt.Sprintf("managed/transit_engine/%s/%s", r.Namespace, r.Name), nil
}

func (r *VaultTransitEngine) GetTransitEngineConfig() (*transit.EngineConfig, error) {
	return &transit.EngineConfig{}, nil
}

func (r *VaultTransitEngine) GetPluginName() (string, error) {
	return r.Spec.Plugin, nil
}
