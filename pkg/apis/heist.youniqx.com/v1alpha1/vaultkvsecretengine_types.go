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

// VaultKVSecretEngineSpec defines the desired state of VaultKVSecretEngine.
type VaultKVSecretEngineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// MaxVersions configures the maximum number of secret versions to keep
	MaxVersions int `json:"maxVersions"`

	// DeleteProtection configures that the secret engine should not be able to be deleted.
	// Defaults to false.
	// +optional
	DeleteProtection bool `json:"deleteProtection"`
}

// VaultKVSecretEngineStatus defines the observed state of VaultKVSecretEngine.
type VaultKVSecretEngineStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:resource:shortName=kvse,categories=heist;youniqx
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provisioned",type="string",JSONPath=".status.conditions[?(@.type=='Provisioned')].status",description="The status of this VaultKVSecretEngine"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation Timestamp of the engine"
// +genclient

// VaultKVSecretEngine is the Schema for the vaultkvsecretengines API.
type VaultKVSecretEngine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultKVSecretEngineSpec   `json:"spec,omitempty"`
	Status VaultKVSecretEngineStatus `json:"status,omitempty"`
}

// VaultKVSecretEngineList contains a list of VaultKVSecretEngine
// +kubebuilder:object:root=true
type VaultKVSecretEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultKVSecretEngine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultKVSecretEngine{}, &VaultKVSecretEngineList{})
}
