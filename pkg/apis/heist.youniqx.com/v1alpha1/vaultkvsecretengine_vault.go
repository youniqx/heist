package v1alpha1

import (
	"fmt"

	"github.com/youniqx/heist/pkg/vault/kvengine"
)

func (r *VaultKVSecretEngine) GetMountPath() (string, error) {
	return fmt.Sprintf("managed/kv/%s/%s", r.Namespace, r.Name), nil
}

func (r *VaultKVSecretEngine) GetKvEngineConfig() (*kvengine.Config, error) {
	var maxVersions int
	if r.Spec.MaxVersions != 0 {
		maxVersions = r.Spec.MaxVersions
	} else {
		maxVersions = 10
	}

	return &kvengine.Config{
		MaxVersions:        maxVersions,
		CasRequired:        true,
		DeleteVersionAfter: "0s",
	}, nil
}
