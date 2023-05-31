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

// FakeVaultKVSecrets implements VaultKVSecretInterface
type FakeVaultKVSecrets struct {
	Fake *FakeHeistV1alpha1
	ns   string
}

var vaultkvsecretsResource = v1alpha1.SchemeGroupVersion.WithResource("vaultkvsecrets")

var vaultkvsecretsKind = v1alpha1.SchemeGroupVersion.WithKind("VaultKVSecret")

// Get takes name of the vaultKVSecret, and returns the corresponding vaultKVSecret object, and an error if there is any.
func (c *FakeVaultKVSecrets) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VaultKVSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(vaultkvsecretsResource, c.ns, name), &v1alpha1.VaultKVSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultKVSecret), err
}

// List takes label and field selectors, and returns the list of VaultKVSecrets that match those selectors.
func (c *FakeVaultKVSecrets) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VaultKVSecretList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(vaultkvsecretsResource, vaultkvsecretsKind, c.ns, opts), &v1alpha1.VaultKVSecretList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VaultKVSecretList{ListMeta: obj.(*v1alpha1.VaultKVSecretList).ListMeta}
	for _, item := range obj.(*v1alpha1.VaultKVSecretList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested vaultKVSecrets.
func (c *FakeVaultKVSecrets) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(vaultkvsecretsResource, c.ns, opts))

}

// Create takes the representation of a vaultKVSecret and creates it.  Returns the server's representation of the vaultKVSecret, and an error, if there is any.
func (c *FakeVaultKVSecrets) Create(ctx context.Context, vaultKVSecret *v1alpha1.VaultKVSecret, opts v1.CreateOptions) (result *v1alpha1.VaultKVSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(vaultkvsecretsResource, c.ns, vaultKVSecret), &v1alpha1.VaultKVSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultKVSecret), err
}

// Update takes the representation of a vaultKVSecret and updates it. Returns the server's representation of the vaultKVSecret, and an error, if there is any.
func (c *FakeVaultKVSecrets) Update(ctx context.Context, vaultKVSecret *v1alpha1.VaultKVSecret, opts v1.UpdateOptions) (result *v1alpha1.VaultKVSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(vaultkvsecretsResource, c.ns, vaultKVSecret), &v1alpha1.VaultKVSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultKVSecret), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVaultKVSecrets) UpdateStatus(ctx context.Context, vaultKVSecret *v1alpha1.VaultKVSecret, opts v1.UpdateOptions) (*v1alpha1.VaultKVSecret, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(vaultkvsecretsResource, "status", c.ns, vaultKVSecret), &v1alpha1.VaultKVSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultKVSecret), err
}

// Delete takes name of the vaultKVSecret and deletes it. Returns an error if one occurs.
func (c *FakeVaultKVSecrets) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(vaultkvsecretsResource, c.ns, name, opts), &v1alpha1.VaultKVSecret{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVaultKVSecrets) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(vaultkvsecretsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.VaultKVSecretList{})
	return err
}

// Patch applies the patch and returns the patched vaultKVSecret.
func (c *FakeVaultKVSecrets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VaultKVSecret, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(vaultkvsecretsResource, c.ns, name, pt, data, subresources...), &v1alpha1.VaultKVSecret{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultKVSecret), err
}
