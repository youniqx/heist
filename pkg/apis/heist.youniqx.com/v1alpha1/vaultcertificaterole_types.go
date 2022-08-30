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

// VaultCertificateRoleSpec defines the desired state of VaultCertificateRole.
type VaultCertificateRoleSpec struct {
	// Issuer specifies the certificate authority used to issue the certificate.
	Issuer string `json:"issuer,omitempty"`

	// Subject configures the subject fields of the Certificate.
	Subject VaultCertificateRoleSubject `json:"subject,omitempty"`

	// Settings configures the settings of the certificate.
	Settings VaultCertificateRoleSettings `json:"settings,omitempty"`
}

type VaultCertificateRoleSubject struct {
	// Organization sets the organization (O) field in the certificate subject.
	// +optional
	Organization []string `json:"organization,omitempty"`

	// OrganizationalUnit sets the organizational unit (OU) field in the certificate subject.
	// +optional
	OrganizationalUnit []string `json:"ou,omitempty"`

	// Country sets the country field (C) in the certificate subject.
	// +optional
	Country []string `json:"country,omitempty"`

	// Locality sets the locality field (L) in the certificate subject.
	// +optional
	Locality []string `json:"locality,omitempty"`

	// Province sets the state or province field (ST) in the certificate subject.
	// +optional
	Province []string `json:"province,omitempty"`

	// StreetAddress sets the street address field in the certificate subject.
	// +optional
	StreetAddress []string `json:"streetAddress,omitempty"`

	// PostalCode sets the postal code field in the certificate subject.
	// +optional
	PostalCode []string `json:"postalCode,omitempty"`
}

type VaultCertificateRoleSettings struct {
	// TTL configures the validity of the certificate.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#ttl.
	// +required
	// +kubebuilder:validation:Required
	TTL metav1.Duration `json:"ttl,omitempty"`

	// MaxTTL configures the maximum validity of the certificate.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#max_ttl.
	// +optional
	MaxTTL metav1.Duration `json:"maxTTL,omitempty"`

	// AllowLocalhost configures if the certificate is valid for localhost.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allow_localhost.
	// +optional
	AllowLocalhost bool `json:"allowLocalhost,omitempty"`

	// AllowedDomains configures a list of domains for which this certificate can be issued.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allowed_domains.
	// +optional
	AllowedDomains []string `json:"allowedDomains,omitempty"`

	// AllowedDomainsTemplate configures if the list of allowed domains can make used of templates.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allowed_domains_template.
	// +optional
	AllowedDomainsTemplate bool `json:"allowedDomainsTemplate,omitempty"`

	// AllowBareDomains configures if certificates can be issued for bare domains.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allow_bare_domains.
	// +optional
	AllowBareDomains bool `json:"allowBareDomains,omitempty"`

	// AllowSubdomains configures if certificates can be issued for subdomains.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allow_subdomains.
	// +optional
	AllowSubdomains bool `json:"allowSubdomains,omitempty"`

	// AllowGlobDomains configures if certificates can be issued for wildcard domains.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allow_glob_domains.
	// +optional
	AllowGlobDomains bool `json:"allowGlobDomains,omitempty"`

	// AllowAnyName configures if certificates can be issued for any common name.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allow_any_name.
	// +optional
	AllowAnyName bool `json:"allowAnyName,omitempty"`

	// EnforceHostNames configures if host names should be enforced.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#enforce_hostnames.
	// +optional
	EnforceHostNames bool `json:"enforceHostNames,omitempty"`

	// AllowIPSans configures if certificates with IP subject alternative names
	// can be issued.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allow_ip_sans.
	// +optional
	AllowIPSans bool `json:"allowIPSans,omitempty"`

	// AllowedURISans configures an allow list of URI subject alternative names
	// for which certificates can be issued.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allowed_uri_sans.
	//	+optional
	AllowedURISans []string `json:"allowedURISans,omitempty"`

	// AllowedOtherSans configures an allow list of other subject alternative
	// names for which certificates can be issued.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#allowed_other_sans.
	// +optional
	AllowedOtherSans []string `json:"allowedOtherSans,omitempty"`

	// ServerFlag configures if issued certificates should have the server flag
	// set.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#server_flag.
	// +optional
	ServerFlag bool `json:"serverFlag,omitempty"`

	// ClientFlag configures if issued certificates should have the client flag
	// set.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#client_flag.
	// +optional
	ClientFlag bool `json:"clientFlag,omitempty"`

	// CodeSigningFlag configures if issued certificates should have the code
	// signing flag set.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#code_signing_flag.
	// +optional
	CodeSigningFlag bool `json:"codeSigningFlag,omitempty"`

	// EmailProtectionFlag configures if issued certificates should have the
	// email protection flag set.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#email_protection_flag.
	// +optional
	EmailProtectionFlag bool `json:"emailProtectionFlag,omitempty"`

	// KeyType sets the key algorithm of the CA certificate.
	// Can be either rsa, ec or any if either type and any bit size should be
	// supported. ED25519 is not supported yet.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#key_type-3.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:=rsa;ec;any
	// +kubebuilder:default:=any
	KeyType pki.KeyType `json:"keyType"`

	// KeyBits sets the size of the key of the certificate authority.
	// Ignored in signing operations when KeyType is `any`.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#key_bits-3.
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum:=224;256;384;521;2048;3072;4096
	KeyBits pki.KeyBits `json:"keyBits,omitempty"`

	// KeyUsage configures a list of usages issued certificate should allow.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#key_usage-1.
	// +optional
	KeyUsage []pki.KeyUsage `json:"keyUsage,omitempty"`

	// ExtendedKeyUsage configures a list of extended key usages issued
	// certificate should allow.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#ext_key_usage-1.
	// +optional
	ExtendedKeyUsage []pki.ExtendedKeyUsage `json:"extKeyUsage,omitempty"`

	// ExtendedKeyUsageOIDS configures a list of key usage OIDs issued
	// certificate should allow.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#ext_key_usage_oids-1.
	// +optional
	ExtendedKeyUsageOIDS []string `json:"extKeyUsageOIDS,omitempty"`

	// UseCSRCommonName configures if the common name from a CSR should be set
	// in issued certificate.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#use_csr_common_name.
	// +optional
	UseCSRCommonName bool `json:"useCSRCommonName,omitempty"`

	// UseCSRSans configures if the subject alternative names from a CSR should
	// be included in issued certificates.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#use_csr_sans.
	// +optional
	UseCSRSans bool `json:"useCSRSans,omitempty"`

	// RequireCommonName configures if setting a common name is required when
	// issuing certificates.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#require_cn.
	// +optional
	RequireCommonName bool `json:"requireCN,omitempty"`

	// PolicyIdentifiers configures a list of policy OIDs which should be set
	// on issued certificates.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#policy_identifiers.
	// +optional
	PolicyIdentifiers []string `json:"policyIdentifiers,omitempty"`

	// BasicConstraintsValidForNonCA configures if basic constraints should be
	// valid when issuing non-ca certificates.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#basic_constraints_valid_for_non_ca.
	// +optional
	BasicConstraintsValidForNonCA bool `json:"basicConstraintsValidForNonCA,omitempty"`

	// NotBeforeDuration configures a delay which has to elapse for any issued
	// certificate to become valid.
	// Additional information: https://www.vaultproject.io/api-docs/secret/pki#not_before_duration-2.
	// +optional
	NotBeforeDuration metav1.Duration `json:"notBeforeDuration,omitempty"`
}

// VaultCertificateRoleStatus defines the observed state of VaultCertificateRole.
type VaultCertificateRoleStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=vc,categories=heist;youniqx
// +kubebuilder:printcolumn:name="Provisioned",type="string",JSONPath=".status.conditions[?(@.type=='Provisioned')].status",description="The status of this CertificateRole"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation Timestamp of the certificate"
// +genclient

// VaultCertificateRole is the Schema for the VaultCertificateRole API.
type VaultCertificateRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultCertificateRoleSpec   `json:"spec,omitempty"`
	Status VaultCertificateRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VaultCertificateRoleList contains a list of VaultCertificateRole.
type VaultCertificateRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultCertificateRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultCertificateRole{}, &VaultCertificateRoleList{})
}
