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

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	scheme "github.com/youniqx/heist/pkg/client/heist.youniqx.com/v1alpha1/clientset/heist/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// VaultClientConfigsGetter has a method to return a VaultClientConfigInterface.
// A group's client should implement this interface.
type VaultClientConfigsGetter interface {
	VaultClientConfigs(namespace string) VaultClientConfigInterface
}

// VaultClientConfigInterface has methods to work with VaultClientConfig resources.
type VaultClientConfigInterface interface {
	Create(ctx context.Context, vaultClientConfig *v1alpha1.VaultClientConfig, opts v1.CreateOptions) (*v1alpha1.VaultClientConfig, error)
	Update(ctx context.Context, vaultClientConfig *v1alpha1.VaultClientConfig, opts v1.UpdateOptions) (*v1alpha1.VaultClientConfig, error)
	UpdateStatus(ctx context.Context, vaultClientConfig *v1alpha1.VaultClientConfig, opts v1.UpdateOptions) (*v1alpha1.VaultClientConfig, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.VaultClientConfig, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.VaultClientConfigList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VaultClientConfig, err error)
	VaultClientConfigExpansion
}

// vaultClientConfigs implements VaultClientConfigInterface
type vaultClientConfigs struct {
	client rest.Interface
	ns     string
}

// newVaultClientConfigs returns a VaultClientConfigs
func newVaultClientConfigs(c *HeistV1alpha1Client, namespace string) *vaultClientConfigs {
	return &vaultClientConfigs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the vaultClientConfig, and returns the corresponding vaultClientConfig object, and an error if there is any.
func (c *vaultClientConfigs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.VaultClientConfig, err error) {
	result = &v1alpha1.VaultClientConfig{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VaultClientConfigs that match those selectors.
func (c *vaultClientConfigs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.VaultClientConfigList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.VaultClientConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested vaultClientConfigs.
func (c *vaultClientConfigs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a vaultClientConfig and creates it.  Returns the server's representation of the vaultClientConfig, and an error, if there is any.
func (c *vaultClientConfigs) Create(ctx context.Context, vaultClientConfig *v1alpha1.VaultClientConfig, opts v1.CreateOptions) (result *v1alpha1.VaultClientConfig, err error) {
	result = &v1alpha1.VaultClientConfig{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(vaultClientConfig).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a vaultClientConfig and updates it. Returns the server's representation of the vaultClientConfig, and an error, if there is any.
func (c *vaultClientConfigs) Update(ctx context.Context, vaultClientConfig *v1alpha1.VaultClientConfig, opts v1.UpdateOptions) (result *v1alpha1.VaultClientConfig, err error) {
	result = &v1alpha1.VaultClientConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		Name(vaultClientConfig.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(vaultClientConfig).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *vaultClientConfigs) UpdateStatus(ctx context.Context, vaultClientConfig *v1alpha1.VaultClientConfig, opts v1.UpdateOptions) (result *v1alpha1.VaultClientConfig, err error) {
	result = &v1alpha1.VaultClientConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		Name(vaultClientConfig.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(vaultClientConfig).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the vaultClientConfig and deletes it. Returns an error if one occurs.
func (c *vaultClientConfigs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *vaultClientConfigs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched vaultClientConfig.
func (c *vaultClientConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.VaultClientConfig, err error) {
	result = &v1alpha1.VaultClientConfig{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("vaultclientconfigs").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}