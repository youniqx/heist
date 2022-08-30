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
	"github.com/youniqx/heist/pkg/vault/transit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VaultTransitKeySpec defines the desired state of VaultTransitKey.
type VaultTransitKeySpec struct {
	// Engine configures the used transit engine.
	// +required
	// +kubebuilder:validation:Required
	Engine string `json:"engine"`

	// Type configures the transit key type. Must be a vault supported key type.
	// Additional information: https://www.vaultproject.io/api/secret/transit#type.
	// +required
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum:=aes128-gcm96;aes256-gcm96;chacha20-poly1305;ed25519;ecdsa-p256;ecdsa-p384;ecdsa-p521;rsa-2048;rsa-3072;rsa-4096
	Type transit.KeyType `json:"type"`

	// MinimumDecryptionVersion specifies the minimum version of the key that can be used to decrypt the ciphertext.
	// Adjusting this as part of a key rotation policy can prevent old copies of ciphertext from being
	// decrypted, should they fall into the wrong hands. For signatures, this value controls the minimum
	// version of signature that can be verified against. For HMACs, this controls the minimum version
	// of a key allowed to be used as the key for verification.
	// +optional
	// +kubebuilder:validation:Optional
	MinimumDecryptionVersion int `json:"minimumDecryptionVersion,omitempty"`

	// MinimumEncryptionVersion Specifies the minimum version of the key that can be used to encrypt
	// plaintext, sign payloads, or generate HMACs. Must be 0 (which will use the latest version) or
	// a value greater or equal to min_decryption_version.
	// +optional
	// +kubebuilder:validation:Optional
	MinimumEncryptionVersion int `json:"minimumEncryptionVersion,omitempty"`

	// Exportable enables keys to be exportable. This allows for all the valid keys in the key
	// ring to be exported. Once set, this cannot be disabled.
	// +optional
	// +kubebuilder:validation:Optional
	Exportable bool `json:"exportable,omitempty"`

	// AllowPlaintextBackup enables taking backups of named key in the
	// plaintext format. Once set, this cannot be disabled.
	// +optional
	// +kubebuilder:validation:Optional
	AllowPlaintextBackup bool `json:"allowPlaintextBackup,omitempty"`

	// DeleteProtection configures that the secret should not be able to be deleted.
	// Defaults to false.
	// +optional
	// +kubebuilder:validation:Optional
	DeleteProtection bool `json:"deleteProtection,omitempty"`
}

// VaultTransitKeyStatus defines the observed state of VaultTransitKey.
type VaultTransitKeyStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// AppliedSpec contains more information about the current state of the
	// VaultTransitKey object.
	// +optional
	AppliedSpec VaultTransitKeySpec `json:"appliedSpec,omitempty"`
}

// +kubebuilder:resource:shortName=vtk,categories=heist;youniqx
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Synced",type="string",JSONPath=".status.conditions[?(@.type=='Provisioned')].status",description="The status of this VaultTransitKey"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation Timestamp of the VaultTransitKey"
// +genclient

// VaultTransitKey is the Schema for the vaulttransitengines API.
type VaultTransitKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultTransitKeySpec   `json:"spec,omitempty"`
	Status VaultTransitKeyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VaultTransitKeyList contains a list of VaultTransitKey.
type VaultTransitKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultTransitKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultTransitKey{}, &VaultTransitKeyList{})
}
