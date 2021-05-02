// +k8s:deepcopy-gen=package,register
// +kubebuilder:object:generate=true
// +groupName=gcpauthpolicy.nuxeo.io
// +versionName=v1alpha1
package v1alpha1

import (
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	ProfilesKey KeyValue = "gcpauthpolicy.nuxeo.io/profiles"
	TypeKey     KeyValue = "gcpauthpolicy.nuxeo.io/type"
	WatchKey    KeyValue = "gcpauthpolicy.nuxeo.io/watch"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "gcpauthpolicy.nuxeo.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	ProfilesResource = SchemeGroupVersion.WithResource("profiles")
)

type ResourceKind string

func (name ResourceKind) String() string {
	return string(name)
}

const (
	ProfileKind ResourceKind = "Profile"
)

// KeyValue typed gcpauth annotation identifiers
type KeyValue string

func (name KeyValue) String() string {
	return string(name)
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// GCPAuthProfile is the schema for the GCPAuthPolicy profile API
type Profile struct {
	meta_api.TypeMeta   `json:",inline"`
	meta_api.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProfileSpec   `json:"spec,omitempty"`
	Status ProfileStatus `json:"status,omitempty"`
}

type ProfileSpec struct {
	Namespaces string                  `json:"namespaces,omitempty"`
	Selector   *meta_api.LabelSelector `json:"selector,omitempty"`
	Datasource SecretRef               `json:"datasource,omitempty"`
}

type SecretRef struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

// +kubebuilder:object:root=true

// ProfileList contains a list of GCPAuthPolicyProfile
type ProfileList struct {
	meta_api.TypeMeta `json:",inline"`
	meta_api.ListMeta `json:"metadata,omitempty"`
	Items             []Profile `json:"items"`
}

// ProfileStatus the status
type ProfileStatus struct {
}

func init() {
	SchemeBuilder.Register(&Profile{}, &ProfileList{})
}
