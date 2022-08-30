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
	"github.com/youniqx/heist/pkg/vault/pki"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VaultCertificateAuthoritySpec defines the desired state of VaultCertificateAuthority.
type VaultCertificateAuthoritySpec struct {
	// Plugin configures the plugin backend used for this engine. Defaults to pki.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Default:=pki
	Plugin string `json:"plugin,omitempty"`

	// Issuer implicitly defines whether the CA is an intermediate or a root CA.
	// If left empty the CA is assumed to be a root CA and will be self-signed.
	// Otherwise, the configured name is a reference to the parent CAs Kubernetes
	// object.
	// +optional
	Issuer string `json:"issuer,omitempty"`

	// Import can be used to import an already existing certificate.
	// +optional
	Import *VaultCertificateAuthorityImport `json:"import,omitempty"`

	// Subject configures the subject fields of the Certificate Authority
	// It is recommended to set a least one field im the Subject section
	// +optional
	Subject VaultCertificateAuthoritySubject `json:"subject,omitempty"`

	// Tuning can be used to tune the PKI Secret Engine in Vault
	// +optional
	Tuning VaultCertificateAuthorityTuning `json:"tuning,omitempty"`

	// Settings configures the key pair of the Certificate Authority
	Settings VaultCertificateAuthoritySettings `json:"settings,omitempty"`

	// DeleteProtection configures that the secret should not be able to be deleted.
	// Defaults to false.
	// +optional
	DeleteProtection bool `json:"deleteProtection"`
}

type VaultCertificateAuthorityImport struct {
	// Certificate contains the certificate matching the private key that should
	// be imported. Can be either encrypted, or plain text.
	Certificate string `json:"certificate,omitempty"`

	// PrivateKey is the private key that should be imported. The private key
	// must be encrypted with the default Heist transit engine to ensure no
	// secrets are stored in plaintext as a Kubernetes object.
	PrivateKey string `json:"privateKey,omitempty"`
}

type VaultCertificateAuthoritySubject struct {
	// CommonName sets the CN (common name) field in the certificate subject
	// +optional
	CommonName string `json:"commonName,omitempty"`

	// Organization sets the organization (O) field in the certificate's subject.
	// +optional
	Organization []string `json:"organization,omitempty"`

	// OrganizationalUnit sets the OU (organizational unit) field in the
	// certificate's subject.
	// +optional
	OrganizationalUnit []string `json:"ou,omitempty"`

	// Country sets the C (country) field in the certificate's subject.
	// +optional
	Country []string `json:"country,omitempty"`

	// Locality sets the L (locality) field in the certificate's subject.
	// +optional
	Locality []string `json:"locality,omitempty"`

	// Province sets the ST (province) field in the certificate's subject.
	// +optional
	Province []string `json:"province,omitempty"`

	// StreetAddress sets the street address field in the certificate's subject.
	// +optional
	StreetAddress []string `json:"streetAddress,omitempty"`

	// PostalCode sets the postal code field in the certificate's subject.
	// +optional
	PostalCode []string `json:"postalCode,omitempty"`
}

type VaultCertificateAuthorityTuning struct {
	// DefaultLeaseTTL sets the default validity of certificates issued by the
	// PKI secret engine.
	// +optional
	DefaultLeaseTTL metav1.Duration `json:"defaultLeaseTTL,omitempty"`

	// MaxLeaseTTL sets the maximum validity of any certificate issued by the
	// PKI secret engine.
	// +optional
	MaxLeaseTTL metav1.Duration `json:"maxLeaseTTL,omitempty"`

	// Description sets the description of the PKI secret engine in Vault.
	// +optional
	Description string `json:"description,omitempty"`
}

type VaultCertificateAuthoritySettings struct {
	// SubjectAlternativeNames sets subject alternative names extensions for
	// the certificate.
	// +optional
	SubjectAlternativeNames []string `json:"subjectAlternativeNames,omitempty"`

	// IPSans sets the IP subject alternative names extension for the
	// certificate.
	// +optional
	IPSans []string `json:"ipSans,omitempty"`

	// URISans sets URI subject alternative names extension for the
	// certificate.
	// +optional
	URISans []string `json:"uriSans,omitempty"`

	// OtherSans sets subject alternative names extension that do not fall into
	// the other categories for the certificate.
	// +optional
	OtherSans []string `json:"otherSans,omitempty"`

	// TTL sets the validity period of the CA certificate.
	// +required
	// +kubebuilder:validation:Required
	TTL metav1.Duration `json:"ttl,omitempty"`

	// KeyType sets the key algorithm of the CA certificate. Can be either rsa
	// or ec.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=rsa;ec
	// +kubebuilder:default:=rsa
	KeyType pki.KeyType `json:"keyType"`

	// KeyBits sets the size of the key of the certificate authority. The
	// KeyBits value provided must be a valid value for the configured KeyType.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=224;256;384;521;2048;3072;4096
	// +kubebuilder:default:=2048
	KeyBits pki.KeyBits `json:"keyBits"`

	// ExcludeCNFromSans configures if the common name set in the subject should
	// be excluded from the subject alternative names extension.
	// +optional
	ExcludeCNFromSans bool `json:"excludeCNFromSans,omitempty"`

	// PermittedDNSDomains configures an allow list of domains for which
	// certificates can be issued using the certificate authority.
	// +optional
	PermittedDNSDomains []string `json:"permittedDNSDomains,omitempty"`

	// Exported configures if the CA should be generated in exported mode.
	// If this is set to true then the private key of the CA can be bound to
	// and accessed by applications. If it is set to false then the private key
	// will be inaccessible. Defaults to false. This setting can not be changed
	// after the PKI is created.
	// +optional
	Exported bool `json:"exported,omitempty"`
}

// VaultCertificateAuthorityStatus defines the observed state of VaultCertificateAuthority.
type VaultCertificateAuthorityStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=vca,categories=heist;youniqx
// +kubebuilder:printcolumn:name="Provisioned",type="string",JSONPath=".status.conditions[?(@.type=='Provisioned')].status",description="The status of this Certificate Authority"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation Timestamp of the Certificate Authority"
// +genclient

// VaultCertificateAuthority is the Schema for the VaultCertificateAuthorities API.
type VaultCertificateAuthority struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultCertificateAuthoritySpec   `json:"spec,omitempty"`
	Status VaultCertificateAuthorityStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VaultCertificateAuthorityList contains a list of VaultCertificateAuthority.
type VaultCertificateAuthorityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultCertificateAuthority `json:"items,omitempty"`
}

func init() {
	SchemeBuilder.Register(&VaultCertificateAuthority{}, &VaultCertificateAuthorityList{})
}
