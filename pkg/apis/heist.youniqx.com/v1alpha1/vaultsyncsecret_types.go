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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultSyncSecretTarget defines the desired state of VaultSyncSecret.
type VaultSyncSecretTarget struct {
	// Name is the name of the secret resource you want to create.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name,omitempty"`

	// Namespace is the namespace the secret should be created in.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=63
	Namespace string `json:"namespace,omitempty"`

	// Type is the type of secret which should be created.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:=Opaque;kubernetes.io/dockercfg;kubernetes.io/dockerconfigjson;kubernetes.io/basic-auth;kubernetes.io/ssh-auth;kubernetes.io/tls
	Type v1.SecretType `json:"type,omitempty"`
}

// VaultSyncCertificateAuthority configures syncing of values from a
// VaultCertificateAuthority.
type VaultSyncCertificateAuthority struct {
	// Name is the name of the VaultCertificateAuthority which should be
	// synced.
	// +required
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Fields is a list of fields which should be synced from the
	// VaultCertificateAuthority.
	// +required
	// +kubebuilder:validation:Required
	Fields []VaultSyncCertificateField `json:"fields,omitempty"`
}

// VaultSyncCertificate configures syncing of values from a VaultCertificateRole.
type VaultSyncCertificate struct {
	// Name is the name of the VaultCertificateAuthority which should be
	// synced.
	// +required
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Fields is a list of fields which should be synced from the
	// VaultCertificateAuthority.
	// +required
	// +kubebuilder:validation:Required
	Fields []VaultSyncCertificateField `json:"fields"`

	// CommonName is the CN (common name) of the issued certificate.
	// +required
	// +kubebuilder:validation:Required
	CommonName string `json:"commonName"`

	// AlternativeNames is a list of SANs (subject alternative names) requested
	// for this certificate. These will be set as an extension in the
	// certificate.
	// +optional
	// +kubebuilder:validation:Optional
	AlternativeNames []string `json:"alternativeNames,omitempty"`

	// ExcludeCNFromSans disables automatically adding the common name to the
	// SAN list.
	// +optional
	// +kubebuilder:validation:Optional
	ExcludeCNFromSans bool `json:"excludeCNFromSans,omitempty"`
}

// VaultSyncCertificateField configures syncing of values from a certificate.
type VaultSyncCertificateField struct {
	// Type is the name of the field which should be bound. Possible values are
	// defined in VaultCertificateFieldType.
	// +kubebuilder:validation:Enum:=certificate;private_key;cert_chain;full_cert_chain
	// +kubebuilder:default:=certificate
	Type VaultCertificateFieldType `json:"field"`

	// Key is the secret key used to store the value.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
}

// VaultSyncSecretStatus defines the observed state of VaultSyncSecret.
type VaultSyncSecretStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	AppliedSpec VaultSyncSecretSpec `json:"appliedSpec,omitempty"`
}

type VaultSyncKvSecret struct {
	// Name is the name of the VaultKVSecret.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Field is the name of the field in the VaultKVSecret.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Field string `json:"field,omitempty"`

	// Key is the secret key used to store the value.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key,omitempty"`
}

type VaultSyncCertificateAuthoritySource struct {
	// Name is the name of the VaultCertificateAuthority which should be synced.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Field is the field of the certificate authority which should be synced.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=certificate;private_key;cert_chain;full_cert_chain
	Field VaultCertificateFieldType `json:"field,omitempty"`
}

type VaultSyncCertificateSource struct {
	// Name is the name of the certificate template used to issue the
	// certificate which should be synced.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Field is the field of the certificate which should be synced.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=certificate;private_key;cert_chain;full_cert_chain
	Field VaultCertificateFieldType `json:"field,omitempty"`
}

type VaultSyncKVSecretSource struct {
	// Name is the name of the VaultKVSecret which should be synced.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name,omitempty"`

	// Field specifies a single field of the VaultKVSecret which should be synced.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Field string `json:"field,omitempty"`
}

type VaultSyncSecretSource struct {
	// CipherText represents a value which has been encrypted by Heists managed
	// Transit Engine.
	// +optional
	// +kubebuilder:validation:Optional
	CipherText EncryptedValue `json:"cipherText,omitempty"`

	// CertificateAuthority configures a VaultCertificateAuthority from which a
	// field should be synced.
	// +optional
	// +kubebuilder:validation:Optional
	CertificateAuthority *VaultSyncCertificateAuthoritySource `json:"certificateAuthority,omitempty"`

	// Certificate configures a VaultCertificateRole from which a field should be
	// synced.
	// +optional
	// +kubebuilder:validation:Optional
	Certificate *VaultSyncCertificateSource `json:"certificate,omitempty"`

	// KVSecret configures a VaultKVSecret from which a field should be synced
	// +optional
	// +kubebuilder:validation:Optional
	KVSecret *VaultSyncKVSecretSource `json:"kvSecret,omitempty"`
}

type VaultSyncSecretSpec struct {
	// Target configures the secret you want to sync values to.
	// +required
	// +kubebuilder:validation:Required
	Target VaultSyncSecretTarget `json:"target,omitempty"`

	// CertificateTemplates configures settings for certificates which may be
	// issued.
	// +optional
	// +kubebuilder:validation:Optional
	CertificateTemplates []VaultCertificateTemplate `json:"certificateTemplates,omitempty"`

	// Data is a map of values which should be synced to the Target Kubernetes
	// Secret.
	// +required
	// +kubebuilder:validation:Required
	Data map[string]VaultSyncSecretSource `json:"data,omitempty"`
}

type VaultCertificateTemplate struct {
	// Alias is the name of this certificate template.
	// +optional
	// +kubebuilder:validation:Optional
	Alias string `json:"alias,omitempty"`

	// CertificateRole is the name of the VaultCertificateRole to be used for issuing
	// this certificate.
	// +required
	// +kubebuilder:validation:Required
	CertificateRole string `json:"certificateRole"`

	// CommonName is the CN (common name) of the issued certificate.
	// +optional
	// +kubebuilder:validation:Optional
	CommonName string `json:"commonName,omitempty"`

	// DNSSans is a list of DNS subject alternative names requested for this
	// certificate.
	// +optional
	// +kubebuilder:validation:Optional
	DNSSans []string `json:"dnsSans,omitempty"`

	// OtherSans is a list of custom OID/UTF-8 subject alternative names
	// requested for this certificate.
	// Expected Format: `<oid>;<type>:<value>`
	// +optional
	// +kubebuilder:validation:Optional
	OtherSans []string `json:"otherSans,omitempty"`

	// IPSans is a list of IP subject alternative names requested for this
	// certificate.
	// +optional
	// +kubebuilder:validation:Optional
	IPSans []string `json:"ipSans,omitempty"`

	// AlternativeNames is a list of URI subject alternative names requested
	// for this certificate.
	// +optional
	// +kubebuilder:validation:Optional
	URISans []string `json:"uriSans,omitempty"`

	// TTL is the Time-To-Live requested for this certificate.
	// +optional
	// +kubebuilder:validation:Optional
	TTL metav1.Duration `json:"ttl,omitempty"`

	// ExcludeCNFromSans toggles if the common name should be excluded from the
	// subject alternative names of the certificate.
	// +optional
	// +kubebuilder:validation:Optional
	ExcludeCNFromSans bool `json:"excludeCNFromSans,omitempty"`
}

// +kubebuilder:resource:shortName=vss,categories=heist;youniqx
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provisioned",type="string",JSONPath=".status.conditions[?(@.type=='Provisioned')].status",description="The status of this VaultSyncSecret"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation Timestamp of the VaultSyncSecret"
// +genclient

// VaultSyncSecret is the Schema for the vaultsyncsecrets API.
type VaultSyncSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultSyncSecretSpec   `json:"spec,omitempty"`
	Status VaultSyncSecretStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VaultSyncSecretList contains a list of VaultSyncSecret.
type VaultSyncSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultSyncSecret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultSyncSecret{}, &VaultSyncSecretList{})
}
