package vaultkvsecret

func (r *Reconciler) performVaultReconciliation(desired *deployedSecret, current *deployedSecret) error {
	if err := r.performVaultPolicyReconciliation(desired, current); err != nil {
		return err
	}

	err := r.performVaultKVSecretReconciliation(desired, current)
	return err
}

func (r *Reconciler) performVaultKVSecretReconciliation(desired *deployedSecret, current *deployedSecret) error {
	var deleteCurrent bool
	if current.Provisioned {
		if desired.Engine != current.Engine {
			deleteCurrent = true
		} else if desired.Secret.Path != current.Secret.Path {
			deleteCurrent = true
		}
	}

	if deleteCurrent {
		if err := r.VaultAPI.DeleteKvSecret(current.Engine, current.Secret); err != nil {
			return err
		}
	}

	err := r.VaultAPI.UpdateKvSecret(desired.Engine, desired.Secret)
	return err
}

func (r *Reconciler) performVaultPolicyReconciliation(desired *deployedSecret, current *deployedSecret) error {
	var deleteCurrent bool
	if current.Provisioned {
		if desired.Policy.Name != current.Policy.Name {
			deleteCurrent = true
		}
	}

	if deleteCurrent {
		if err := r.VaultAPI.DeletePolicy(current.Policy); err != nil {
			return err
		}
	}

	err := r.VaultAPI.UpdatePolicy(desired.Policy)
	return err
}
