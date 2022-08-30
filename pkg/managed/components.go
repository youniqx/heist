package managed

import (
	"path/filepath"

	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/policy"
	"github.com/youniqx/heist/pkg/vault/transit"
)

var managedTransitEngine = &transit.Engine{
	Path:   TransitEnginePath,
	Config: nil,
}

var managedTransitKey = &transit.Key{
	Name: TransitKeyName,
	Type: transit.TypeAes256Gcm96,
	Config: &transit.KeyConfig{
		MinimumDecryptionVersion: 1,
		MinimumEncryptionVersion: 1,
		DeletionAllowed:          false,
		Exportable:               false,
		AllowPlaintextBackup:     false,
	},
}

var managedEncryptPolicy = &policy.Policy{
	Name: EncryptPolicyName,
	Rules: []*policy.Rule{
		{
			Path: filepath.Join(TransitEnginePath, "encrypt", TransitKeyName),
			Capabilities: []policy.Capability{
				policy.UpdateCapability,
			},
		},
	},
}

const (
	PathPrefix         = "managed"
	TransitEnginePath  = "managed/transit"
	TransitKeyName     = "encryption-key"
	EncryptPolicyName  = "managed.encrypt"
	KubernetesAuthPath = "managed/kubernetes"
)

var (
	TransitEngine  core.MountPathEntity  = managedTransitEngine
	TransitKey     transit.KeyNameEntity = managedTransitKey
	KubernetesAuth core.MountPathEntity  = core.MountPath(KubernetesAuthPath)
	EncryptPolicy  core.PolicyName       = EncryptPolicyName
)

func UpdateManagedTransitEngine(api vault.API) error {
	if err := api.UpdateTransitEngine(managedTransitEngine); err != nil {
		return err
	}

	if err := api.UpdateTransitKey(managedTransitEngine, managedTransitKey); err != nil {
		return err
	}

	return api.UpdatePolicy(managedEncryptPolicy)
}
