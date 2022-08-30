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

type VaultCertificateFieldType string

const (
	// VaultBindingCertificateFieldTypeCertChain is the field type for binding
	// the cert chain of a certificate.
	VaultBindingCertificateFieldTypeCertChain VaultCertificateFieldType = "cert_chain"

	// VaultBindingCertificateFieldTypeFullCertChain is the field type for
	// binding the full cert chain (including root) of a certificate.
	VaultBindingCertificateFieldTypeFullCertChain VaultCertificateFieldType = "full_cert_chain"

	// VaultBindingCertificateFieldTypePrivateKey is the field type for binding
	// the private key of a certificate.
	VaultBindingCertificateFieldTypePrivateKey VaultCertificateFieldType = "private_key"

	// VaultBindingCertificateFieldTypeCertificate is the field type for
	// binding the public part a certificate.
	VaultBindingCertificateFieldTypeCertificate VaultCertificateFieldType = "certificate"
)

// VaultBindingSubject defines the desired service account for the VaultBinding.
type VaultBindingSubject struct {
	// Name is the name of the service account you want to grant access to the
	// referenced secrets.
	// +required
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// VaultBindingStatus defines the observed state of VaultBinding.
type VaultBindingStatus struct {
	Conditions  []metav1.Condition `json:"conditions"`
	AppliedSpec VaultBindingSpec   `json:"appliedSpec,omitempty"`
}

type VaultBindingValueTemplate struct {
	// Path is the desired output path for this value. Relative paths are
	// interpreted to be relative to the default Heist secret directory
	// /heist/secrets. The path must be in a shared directory, where the Heist
	// Agent and application container have access.
	// +required
	// +kubebuilder:validation:Required
	Path string `json:"path,omitempty"`

	// Mode is the desired file mode of the output file. 0640 is the default.
	// The agent is the owner of the file, owner permissions will always be
	// read + write and cannot be modified this way. Must be specified as octal.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern:=`^[0][0-7]{3}$`
	Mode string `json:"mode,omitempty"`

	// Template is the template for this value.
	// The template supports [sprig](https://masterminds.github.io/sprig/)
	// template functions and can access all bound secrets and associated
	// capabilities with additional template functions:
	//   - `kvSecret "<name>" "<field>"`: retrieves the value of field "<field>"
	//     from a KV secret with name "<name>".
	//   - `caField "<name>" "<field>"`: retrieves the value of field "<field>"
	//     from CA "<name>". Supported values for "<field>" are defined in
	//     VaultCertificateFieldType.
	//   - `certField "<name>" "<field>"`: retrieves the value of field "<field>"
	//     from certificate template "<name>". Supported values for "<field>"
	//     are defined in VaultCertificateFieldType.
	// +required
	// +kubebuilder:validation:Required
	Template string `json:"template,omitempty"`
}

type VaultBindingAgentConfig struct {
	// CertificateTemplates is a list of certificate templates to be used when issuing
	// certificates in the agent.
	// +optional
	// +kubebuilder:validation:Optional
	CertificateTemplates []VaultCertificateTemplate `json:"certificateTemplates,omitempty"`

	// Templates is a list of files to be populated in relevant pods by the
	// Heist agent.
	// +optional
	// +kubebuilder:validation:Optional
	Templates []VaultBindingValueTemplate `json:"templates,omitempty"`
}

// VaultBindingKVCapability represents capabilities for VaultKVSecret objects
// which can be granted to a subject.
// The "read" capability is granted by default.
// +kubebuilder:validation:Enum:=read
type VaultBindingKVCapability string

const (
	VaultBindingKVCapabilityRead VaultBindingKVCapability = "read"
)

type VaultBindingKV struct {
	// Name is the name of the VaultKVSecret.
	// +required
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Capabilities is a list of granted capabilities for the specified KV
	// secret in Vault.
	// If not otherwise set then the "read" capability is granted by default
	// https://www.vaultproject.io/docs/concepts/policies#capabilities, however,
	// currently Heist only supports "read".
	// +optional
	// +kubebuilder:validation:Optional
	Capabilities []VaultBindingKVCapability `json:"capabilities,omitempty"`
}

// VaultBindingCertificateAuthorityCapability represents Vault capabilities for
// VaultCertificateAuthority objects which can be granted to a subject.
// The "read_public" capability is granted by default
// +kubebuilder:validation:Enum:=read_public;read_private
type VaultBindingCertificateAuthorityCapability string

const (
	VaultBindingCertificateAuthorityCapabilityReadPublic  VaultBindingCertificateAuthorityCapability = "read_public"
	VaultBindingCertificateAuthorityCapabilityReadPrivate VaultBindingCertificateAuthorityCapability = "read_private"
)

type VaultBindingCertificateAuthority struct {
	// Name is the name of the VaultCertificateAuthority Kubernetes object. It
	// is expected to be in the same namespace as the binding.
	// +required
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Capabilities is a list of Vault capabilities for which access is granted.
	// If not otherwise set then the "read_public" capability will be granted
	// by default.
	// +optional
	// +kubebuilder:validation:Optional
	Capabilities []VaultBindingCertificateAuthorityCapability `json:"capabilities,omitempty"`
}

// VaultBindingCertificateCapability represents capabilities for
// VaultCertificateRole objects which can be granted to a subject.
// The "issue" capability is granted by default
// +kubebuilder:validation:Enum:=issue;sign_csr;sign_verbatim
type VaultBindingCertificateCapability string

const (
	// VaultBindingCertificateCapabilityIssue allows the bound ServiceAccount to
	// issue a new certificate based on the provided configuration. This
	// capability is the minimum requirement when issuing a certificate with a
	// VaultBinding. If no Capability is configured, the
	// VaultBindingCertificateCapabilityIssue will be added automatically.
	VaultBindingCertificateCapabilityIssue VaultBindingCertificateCapability = "issue"
	// VaultBindingCertificateCapabilitySignCSR allows the bound ServiceAccount
	// to be able to sign user provided CSRs, using the fields as configured in
	// the VaultCertificateAuthority.
	VaultBindingCertificateCapabilitySignCSR VaultBindingCertificateCapability = "sign_csr"
	// VaultBindingCertificateCapabilitySignVerbatim allows the bound
	// ServiceAccount to be able to sign user provided CSRs, using the
	// fields provided by the CSRs. Generally speaking it is safer to use the
	// capability VaultBindingCertificateCapabilitySignCSR, since it performs
	// validation before issuing the certificate.
	VaultBindingCertificateCapabilitySignVerbatim VaultBindingCertificateCapability = "sign_verbatim"
)

type VaultBindingCertificate struct {
	// Name is the name of the VaultCertificateRole.
	// +required
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Capabilities is a list of Vault capabilities for which access is granted.
	// If not otherwise set then the "issue" capability will be granted by
	// default.
	// +optional
	// +kubebuilder:validation:Optional
	Capabilities []VaultBindingCertificateCapability `json:"capabilities,omitempty"`
}

// VaultBindingHeistCapability represents general capabilities which can be
// granted to a subject.
// +kubebuilder:validation:Enum:=encrypt
type VaultBindingHeistCapability string

const (
	// VaultBindingHeistCapabilityEncrypt allows the service account to use the
	// default managed transit engine `managed.encrypt` to encrypt values.
	VaultBindingHeistCapabilityEncrypt VaultBindingHeistCapability = "encrypt"
)

// VaultBindingTransitKeyCapability represents capabilities for VaultTransitKey
// objects which can be granted to a subject.
// +kubebuilder:validation:Enum:=encrypt;decrypt;datakey;rewrap;sign;hmac;verify;read
type VaultBindingTransitKeyCapability string

const (
	// VaultBindingTransitKeyCapabilityEncrypt allows the service account to use
	// the transit engine to encrypt data.
	VaultBindingTransitKeyCapabilityEncrypt VaultBindingTransitKeyCapability = "encrypt"
	// VaultBindingTransitKeyCapabilityDecrypt allows the service account to use
	// the transit engine to decrypt data.
	VaultBindingTransitKeyCapabilityDecrypt VaultBindingTransitKeyCapability = "decrypt"
	// VaultBindingTransitKeyCapabilityDatakey allows the service account to use
	// the transit engine to use a datakey that can be used for offline de- and
	// encryption. The datakey is NOT the transit key used when encrypting or
	// decrypting values with the API. Vault provides an example
	// [Use Case](https://learn.hashicorp.com/tutorials/vault/eaas-transit#generate-data-key)
	// with a tutorial on how to use datakeys.
	VaultBindingTransitKeyCapabilityDatakey VaultBindingTransitKeyCapability = "datakey"
	// VaultBindingTransitKeyCapabilityRewrap allows the service account to use
	// the transit engine to rewrap an already encrypted secret with the latest
	// version of the encryption key.
	VaultBindingTransitKeyCapabilityRewrap VaultBindingTransitKeyCapability = "rewrap"
	// VaultBindingTransitKeyCapabilitySign allows the service account to use
	// the transit engine to sign data.
	VaultBindingTransitKeyCapabilitySign VaultBindingTransitKeyCapability = "sign"
	// VaultBindingTransitKeyCapabilityHmac allows the service account to use
	// the transit engine to generate a digest of the provided data and key.
	VaultBindingTransitKeyCapabilityHmac VaultBindingTransitKeyCapability = "hmac"
	// VaultBindingTransitKeyCapabilityVerify allows the service account to use
	// the transit engine to verify signed data.
	VaultBindingTransitKeyCapabilityVerify VaultBindingTransitKeyCapability = "verify"
	// VaultBindingTransitKeyCapabilityRead allows the service account to use
	// the transit engine to retrieve information about the transit key. The
	// transit key itself is not exposed via the API.
	VaultBindingTransitKeyCapabilityRead VaultBindingTransitKeyCapability = "read"
)

type VaultBindingTransitKey struct {
	// Name is the name of the VaultTransitKey.
	// +required
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Capabilities is a list of Vault capabilities for which access is granted.
	// +optional
	// +kubebuilder:validation:Optional
	Capabilities []VaultBindingTransitKeyCapability `json:"capabilities,omitempty"`
}

type VaultBindingSpec struct {
	// Subject configures the service account to which access is granted.
	// +required
	// +kubebuilder:validation:Required
	Subject VaultBindingSubject `json:"subject,omitempty"`

	// Capabilities configures general Vault capabilities for which access is
	// granted.
	// +optional
	// +kubebuilder:validation:Optional
	Capabilities []VaultBindingHeistCapability `json:"capabilities,omitempty"`

	// KVSecrets is a list of kv secrets to which access is granted.
	// +optional
	// +kubebuilder:validation:Optional
	KVSecrets []VaultBindingKV `json:"kvSecrets,omitempty"`

	// CertificateAuthorities is a list of certificate authorities to which
	// access is granted.
	// +optional
	// +kubebuilder:validation:Optional
	CertificateAuthorities []VaultBindingCertificateAuthority `json:"certificateAuthorities,omitempty"`

	// CertificateRoles is a list of certificate roles for which access
	// is granted.
	// +optional
	// +kubebuilder:validation:Optional
	CertificateRoles []VaultBindingCertificate `json:"certificateRoles,omitempty"`

	// TransitKeys is a list of transit keys and capabilities for which access
	// is granted.
	// +optional
	// +kubebuilder:validation:Optional
	TransitKeys []VaultBindingTransitKey `json:"transitKeys,omitempty"`

	// Agent can be used to configure the Heist agent sidecar.
	// +optional
	// +kubebuilder:validation:Optional
	Agent VaultBindingAgentConfig `json:"agent,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=vb,categories=heist;youniqx
// +kubebuilder:printcolumn:name="Provisioned",type="string",JSONPath=".status.conditions[?(@.type=='Provisioned')].status",description="The status of this VaultBinding"
// +kubebuilder:printcolumn:name="Active",type="string",JSONPath=".status.conditions[?(@.type=='Active')].status",description="The status of this VaultBinding"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation Timestamp of the binding"
// +genclient

// VaultBinding is the Schema for the VaultBindings API.
type VaultBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultBindingSpec   `json:"spec,omitempty"`
	Status VaultBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VaultBindingList contains a list of VaultBinding.
type VaultBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultBinding{}, &VaultBindingList{})
}
