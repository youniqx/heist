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
	"context"

	v1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVaultSyncSecrets implements VaultSyncSecretInterface
type FakeVaultSyncSecrets struct {
	Fake *FakeHeistV1alpha1
	ns   string
}

var vaultsyncsecretsResource = v1alpha1.SchemeGroupVersion.WithResource("vaultsyncsecrets")

var vaultsyncsecretsKind = v1alpha1.SchemeGroupVersion.WithKind("VaultSyncSecret")

// Get takes name of the vaultSyncSecret, and returns the corresponding vaultSyncSecret object, and an error if there is any.
func (c *FakeVaultSyncSecrets) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VaultSyncSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(vaultsyncsecretsResource, c.ns, name), &v1alpha1.VaultSyncSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSyncSecret), err
}

// List takes label and field selectors, and returns the list of VaultSyncSecrets that match those selectors.
func (c *FakeVaultSyncSecrets) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VaultSyncSecretList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(vaultsyncsecretsResource, vaultsyncsecretsKind, c.ns, opts), &v1alpha1.VaultSyncSecretList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VaultSyncSecretList{ListMeta: obj.(*v1alpha1.VaultSyncSecretList).ListMeta}
	for _, item := range obj.(*v1alpha1.VaultSyncSecretList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested vaultSyncSecrets.
func (c *FakeVaultSyncSecrets) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(vaultsyncsecretsResource, c.ns, opts))

}

// Create takes the representation of a vaultSyncSecret and creates it.  Returns the server's representation of the vaultSyncSecret, and an error, if there is any.
func (c *FakeVaultSyncSecrets) Create(ctx context.Context, vaultSyncSecret *v1alpha1.VaultSyncSecret, opts v1.CreateOptions) (result *v1alpha1.VaultSyncSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(vaultsyncsecretsResource, c.ns, vaultSyncSecret), &v1alpha1.VaultSyncSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSyncSecret), err
}

// Update takes the representation of a vaultSyncSecret and updates it. Returns the server's representation of the vaultSyncSecret, and an error, if there is any.
func (c *FakeVaultSyncSecrets) Update(ctx context.Context, vaultSyncSecret *v1alpha1.VaultSyncSecret, opts v1.UpdateOptions) (result *v1alpha1.VaultSyncSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(vaultsyncsecretsResource, c.ns, vaultSyncSecret), &v1alpha1.VaultSyncSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSyncSecret), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVaultSyncSecrets) UpdateStatus(ctx context.Context, vaultSyncSecret *v1alpha1.VaultSyncSecret, opts v1.UpdateOptions) (*v1alpha1.VaultSyncSecret, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(vaultsyncsecretsResource, "status", c.ns, vaultSyncSecret), &v1alpha1.VaultSyncSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSyncSecret), err
}

// Delete takes name of the vaultSyncSecret and deletes it. Returns an error if one occurs.
func (c *FakeVaultSyncSecrets) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(vaultsyncsecretsResource, c.ns, name, opts), &v1alpha1.VaultSyncSecret{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVaultSyncSecrets) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(vaultsyncsecretsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.VaultSyncSecretList{})
	return err
}

// Patch applies the patch and returns the patched vaultSyncSecret.
func (c *FakeVaultSyncSecrets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VaultSyncSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(vaultsyncsecretsResource, c.ns, name, pt, data, subresources...), &v1alpha1.VaultSyncSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSyncSecret), err
}
