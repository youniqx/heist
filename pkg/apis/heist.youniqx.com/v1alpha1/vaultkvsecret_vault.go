package v1alpha1

import "path/filepath"

func (r *VaultKVSecret) GetSecretPath() (string, error) {
	return filepath.Join(r.Spec.Path, r.Name), nil
}
