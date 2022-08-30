/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vaultbinding

import (
	"github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/vault/kubernetesauth"
)

func (r *Reconciler) updateVaultRole(info *BindingInfo) error {
	role := &kubernetesauth.Role{
		Name:                 info.VaultRoleName,
		BoundServiceAccounts: []string{info.Spec.Subject.Name},
		BoundNamespaces:      []string{info.Binding.Namespace},
		Policies:             info.Policies,
	}

	if len(info.Policies) == 0 {
		if err := r.VaultAPI.DeleteKubernetesAuthRole(managed.KubernetesAuth, role); err != nil {
			return err
		}
	} else {
		if err := r.VaultAPI.UpdateKubernetesAuthRole(managed.KubernetesAuth, role); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) deleteVaultRole(binding *v1alpha1.VaultBinding, spec *v1alpha1.VaultBindingSpec) error {
	return r.VaultAPI.DeleteKubernetesAuthRole(managed.KubernetesAuth, &kubernetesauth.Role{
		Name: getVaultRoleName(binding, spec),
	})
}
