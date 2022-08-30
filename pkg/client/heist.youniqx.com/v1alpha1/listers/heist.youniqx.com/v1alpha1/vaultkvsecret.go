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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// VaultKVSecretLister helps list VaultKVSecrets.
// All objects returned here must be treated as read-only.
type VaultKVSecretLister interface {
	// List lists all VaultKVSecrets in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.VaultKVSecret, err error)
	// VaultKVSecrets returns an object that can list and get VaultKVSecrets.
	VaultKVSecrets(namespace string) VaultKVSecretNamespaceLister
	VaultKVSecretListerExpansion
}

// vaultKVSecretLister implements the VaultKVSecretLister interface.
type vaultKVSecretLister struct {
	indexer cache.Indexer
}

// NewVaultKVSecretLister returns a new VaultKVSecretLister.
func NewVaultKVSecretLister(indexer cache.Indexer) VaultKVSecretLister {
	return &vaultKVSecretLister{indexer: indexer}
}

// List lists all VaultKVSecrets in the indexer.
func (s *vaultKVSecretLister) List(selector labels.Selector) (ret []*v1alpha1.VaultKVSecret, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.VaultKVSecret))
	})
	return ret, err
}

// VaultKVSecrets returns an object that can list and get VaultKVSecrets.
func (s *vaultKVSecretLister) VaultKVSecrets(namespace string) VaultKVSecretNamespaceLister {
	return vaultKVSecretNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// VaultKVSecretNamespaceLister helps list and get VaultKVSecrets.
// All objects returned here must be treated as read-only.
type VaultKVSecretNamespaceLister interface {
	// List lists all VaultKVSecrets in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.VaultKVSecret, err error)
	// Get retrieves the VaultKVSecret from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.VaultKVSecret, error)
	VaultKVSecretNamespaceListerExpansion
}

// vaultKVSecretNamespaceLister implements the VaultKVSecretNamespaceLister
// interface.
type vaultKVSecretNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all VaultKVSecrets in the indexer for a given namespace.
func (s vaultKVSecretNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.VaultKVSecret, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.VaultKVSecret))
	})
	return ret, err
}

// Get retrieves the VaultKVSecret from the indexer for a given namespace and name.
func (s vaultKVSecretNamespaceLister) Get(name string) (*v1alpha1.VaultKVSecret, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("vaultkvsecret"), name)
	}
	return obj.(*v1alpha1.VaultKVSecret), nil
}
