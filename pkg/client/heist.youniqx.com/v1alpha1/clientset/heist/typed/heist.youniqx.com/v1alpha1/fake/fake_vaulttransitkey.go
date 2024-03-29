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

// FakeVaultTransitKeys implements VaultTransitKeyInterface
type FakeVaultTransitKeys struct {
	Fake *FakeHeistV1alpha1
	ns   string
}

var vaulttransitkeysResource = v1alpha1.SchemeGroupVersion.WithResource("vaulttransitkeys")

var vaulttransitkeysKind = v1alpha1.SchemeGroupVersion.WithKind("VaultTransitKey")

// Get takes name of the vaultTransitKey, and returns the corresponding vaultTransitKey object, and an error if there is any.
func (c *FakeVaultTransitKeys) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VaultTransitKey, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(vaulttransitkeysResource, c.ns, name), &v1alpha1.VaultTransitKey{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultTransitKey), err
}

// List takes label and field selectors, and returns the list of VaultTransitKeys that match those selectors.
func (c *FakeVaultTransitKeys) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VaultTransitKeyList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(vaulttransitkeysResource, vaulttransitkeysKind, c.ns, opts), &v1alpha1.VaultTransitKeyList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VaultTransitKeyList{ListMeta: obj.(*v1alpha1.VaultTransitKeyList).ListMeta}
	for _, item := range obj.(*v1alpha1.VaultTransitKeyList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested vaultTransitKeys.
func (c *FakeVaultTransitKeys) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(vaulttransitkeysResource, c.ns, opts))

}

// Create takes the representation of a vaultTransitKey and creates it.  Returns the server's representation of the vaultTransitKey, and an error, if there is any.
func (c *FakeVaultTransitKeys) Create(ctx context.Context, vaultTransitKey *v1alpha1.VaultTransitKey, opts v1.CreateOptions) (result *v1alpha1.VaultTransitKey, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(vaulttransitkeysResource, c.ns, vaultTransitKey), &v1alpha1.VaultTransitKey{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultTransitKey), err
}

// Update takes the representation of a vaultTransitKey and updates it. Returns the server's representation of the vaultTransitKey, and an error, if there is any.
func (c *FakeVaultTransitKeys) Update(ctx context.Context, vaultTransitKey *v1alpha1.VaultTransitKey, opts v1.UpdateOptions) (result *v1alpha1.VaultTransitKey, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(vaulttransitkeysResource, c.ns, vaultTransitKey), &v1alpha1.VaultTransitKey{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultTransitKey), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeVaultTransitKeys) UpdateStatus(ctx context.Context, vaultTransitKey *v1alpha1.VaultTransitKey, opts v1.UpdateOptions) (*v1alpha1.VaultTransitKey, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(vaulttransitkeysResource, "status", c.ns, vaultTransitKey), &v1alpha1.VaultTransitKey{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultTransitKey), err
}

// Delete takes name of the vaultTransitKey and deletes it. Returns an error if one occurs.
func (c *FakeVaultTransitKeys) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(vaulttransitkeysResource, c.ns, name, opts), &v1alpha1.VaultTransitKey{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVaultTransitKeys) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(vaulttransitkeysResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.VaultTransitKeyList{})
	return err
}

// Patch applies the patch and returns the patched vaultTransitKey.
func (c *FakeVaultTransitKeys) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VaultTransitKey, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(vaulttransitkeysResource, c.ns, name, pt, data, subresources...), &v1alpha1.VaultTransitKey{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultTransitKey), err
}
