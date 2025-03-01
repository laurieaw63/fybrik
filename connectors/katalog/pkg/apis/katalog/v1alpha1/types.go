// Copyright 2021 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"fybrik.io/fybrik/pkg/model/datacatalog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Namespaced
// Asset defines an asset in the catalog
type Asset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AssetSpec `json:"spec,omitempty"`
}

type AssetSpec struct {
	// Asset details
	Details datacatalog.ResourceDetails `json:"details"`
	// Asset metadata
	Metadata datacatalog.ResourceMetadata `json:"metadata"`
	// Reference to a Secret resource holding credentials for this asset
	SecretRef SecretRef `json:"secretRef"`
}

// SecretRef is a reference to a local Kubernetes secret.
type SecretRef struct {
	// Name of the Secret resource (must exist in the same namespace)
	Name string `json:"name"`
}

// +kubebuilder:object:root=true
// AssetList contains a list of Asset resources
type AssetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Asset `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Asset{}, &AssetList{})
}
