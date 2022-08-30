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

// VaultTransitEngineSpec defines the desired state of VaultTransitEngine.
type VaultTransitEngineSpec struct {
	// Plugin configures the plugin backend used for this engine. Defaults to transit.
	// https://www.vaultproject.io/docs/upgrading/plugins#overriding-built-in-plugins
	// +optional
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Default:=transit
	Plugin string `json:"plugin,omitempty"`
}

// VaultTransitEngineStatus defines the observed state of VaultTransitEngine.
type VaultTransitEngineStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:resource:shortName=vte,categories=heist;youniqx
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Synced",type="string",JSONPath=".status.conditions[?(@.type=='Provisioned')].status",description="The status of this VaultTransitEngine"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Creation Timestamp of the VaultTransitEngine"
// +genclient

// VaultTransitEngine is the Schema for the vaulttransitengines API.
type VaultTransitEngine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VaultTransitEngineSpec   `json:"spec,omitempty"`
	Status VaultTransitEngineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VaultTransitEngineList contains a list of VaultTransitEngine.
type VaultTransitEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VaultTransitEngine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VaultTransitEngine{}, &VaultTransitEngineList{})
}
