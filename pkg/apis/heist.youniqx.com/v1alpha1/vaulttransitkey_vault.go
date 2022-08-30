package v1alpha1

import "github.com/youniqx/heist/pkg/vault/transit"

var (
	_ transit.KeyEntity     = &VaultTransitKey{}
	_ transit.KeyNameEntity = &VaultTransitKey{}
)

func (r *VaultTransitKey) GetTransitKeyName() (string, error) {
	return r.Name, nil
}

func (r *VaultTransitKey) GetTransitKeyType() (transit.KeyType, error) {
	return r.Spec.Type, nil
}

func (r *VaultTransitKey) GetTransitKeyConfig() (*transit.KeyConfig, error) {
	return &transit.KeyConfig{
		MinimumDecryptionVersion: r.Spec.MinimumDecryptionVersion,
		MinimumEncryptionVersion: r.Spec.MinimumEncryptionVersion,
		DeletionAllowed:          true,
		Exportable:               r.Spec.Exportable,
		AllowPlaintextBackup:     r.Spec.AllowPlaintextBackup,
	}, nil
}
