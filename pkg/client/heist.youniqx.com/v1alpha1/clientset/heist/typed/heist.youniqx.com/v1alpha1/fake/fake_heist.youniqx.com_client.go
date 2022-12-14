/*
Copyright 2022 youniqx Identity AG.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/youniqx/heist/pkg/client/heist.youniqx.com/v1alpha1/clientset/heist/typed/heist.youniqx.com/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeHeistV1alpha1 struct {
	*testing.Fake
}

func (c *FakeHeistV1alpha1) VaultBindings(namespace string) v1alpha1.VaultBindingInterface {
	return &FakeVaultBindings{c, namespace}
}

func (c *FakeHeistV1alpha1) VaultCertificateAuthorities(namespace string) v1alpha1.VaultCertificateAuthorityInterface {
	return &FakeVaultCertificateAuthorities{c, namespace}
}

func (c *FakeHeistV1alpha1) VaultCertificateRoles(namespace string) v1alpha1.VaultCertificateRoleInterface {
	return &FakeVaultCertificateRoles{c, namespace}
}

func (c *FakeHeistV1alpha1) VaultClientConfigs(namespace string) v1alpha1.VaultClientConfigInterface {
	return &FakeVaultClientConfigs{c, namespace}
}

func (c *FakeHeistV1alpha1) VaultKVSecrets(namespace string) v1alpha1.VaultKVSecretInterface {
	return &FakeVaultKVSecrets{c, namespace}
}

func (c *FakeHeistV1alpha1) VaultKVSecretEngines(namespace string) v1alpha1.VaultKVSecretEngineInterface {
	return &FakeVaultKVSecretEngines{c, namespace}
}

func (c *FakeHeistV1alpha1) VaultSyncSecrets(namespace string) v1alpha1.VaultSyncSecretInterface {
	return &FakeVaultSyncSecrets{c, namespace}
}

func (c *FakeHeistV1alpha1) VaultTransitEngines(namespace string) v1alpha1.VaultTransitEngineInterface {
	return &FakeVaultTransitEngines{c, namespace}
}

func (c *FakeHeistV1alpha1) VaultTransitKeys(namespace string) v1alpha1.VaultTransitKeyInterface {
	return &FakeVaultTransitKeys{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeHeistV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
