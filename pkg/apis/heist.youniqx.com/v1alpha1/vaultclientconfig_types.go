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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VaultClientConfigSpec defines the desired state of VaultClientConfig.
type VaultClientConfigSpec struct {
	Address                string                          `json:"address,omitempty"`
	Role                   string                          `json:"role,omitempty"`
	CACerts                []string                        `json:"caCerts,omitempty"`
	AuthMountPath          string                          `json:"authMountPath,omitempty"`
	CertificateAuthorities []*VaultCertificateAuthorityRef `json:"certificateAuthorities,omitempty"`
	KvSecrets              []*VaultKVSecretRef             `json:"kvSecrets,omitempty"`
	Certificates           []*VaultCertificateRef          `json:"certificates,omitempty"`
	TransitKeys            []*VaultTransitKeyRef           `json:"transitKeys,omitempty"`
	Templates              VaultBindingAgentConfig         `json:"templates,omitempty"`
}

type VaultCertificateAuthorityRef struct {
	Name         string                                       `json:"name,omitempty"`
	EnginePath   string                                       `json:"enginePath,omitempty"`
	KVSecrets    VaultCertificateAuthorityKVSecretRef         `json:"kvSecrets,omitempty"`
	Capabilities []VaultBindingCertificateAuthorityCapability `json:"capabilities,omitempty"`
}

type VaultCertificateAuthorityKVSecretRef struct {
	EnginePath        string `json:"enginePath,omitempty"`
	PublicSecretPath  string `json:"publicSecret,omitempty"`
	PrivateSecretPath string `json:"privateSecret,omitempty"`
}

type VaultKVSecretRef struct {
	Name         string                     `json:"name,omitempty"`
	EnginePath   string                     `json:"enginePath,omitempty"`
	SecretPath   string                     `json:"secretPath,omitempty"`
	Capabilities []VaultBindingKVCapability `json:"capabilities,omitempty"`
}

type VaultCertificateRef struct {
	Name         string                              `json:"name,omitempty"`
	EnginePath   string                              `json:"enginePath,omitempty"`
	RoleName     string                              `json:"roleName,omitempty"`
	Capabilities []VaultBindingCertificateCapability `json:"capabilities,omitempty"`
}

type VaultTransitKeyRef struct {
	Name         string                             `json:"name,omitempty"`
	EnginePath   string                             `json:"enginePath,omitempty"`
	KeyName      string                             `json:"keyName,omitempty"`
	Capabilities []VaultBindingTransitKeyCapability `json:"capabilities,omitempty"`
}

// VaultClientConfigStatus defines the observed state of VaultClientConfig.
type VaultClientConfigStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +genclient

// VaultClientConfig is the Schema for the vaultclientconfigs API.
type VaultClientConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultClientConfigSpec   `json:"spec,omitempty"`
	Status VaultClientConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VaultClientConfigList contains a list of VaultClientConfig.
type VaultClientConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultClientConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultClientConfig{}, &VaultClientConfigList{})
}
