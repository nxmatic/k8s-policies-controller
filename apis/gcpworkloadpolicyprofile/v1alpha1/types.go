// +k8s:deepcopy-gen=package,register
// +kubebuilder:object:generate=true
// +groupName=gcpworkloadpolicy.nuxeo.io
// +versionName=v1alpha1
package v1alpha1

import (
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	ProfilesKey KeyValue = "gcpworkloadpolicy.nuxeo.io/profiles"
	TypeKey     KeyValue = "gcpworkloadpolicy.nuxeo.io/type"
	WatchKey    KeyValue = "gcpworkloadpolicy.nuxeo.io/watch"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "gcpworkloadpolicy.nuxeo.io", Version: "v1alpha1"}

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

// KeyValue typed gcpworkload annotation identifiers
type KeyValue string

func (name KeyValue) String() string {
	return string(name)
}

type TypeValue string

func (name TypeValue) String() string {
	return string(name)
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// gcpworkloadProfile is the schema for the gcpworkloadPolicy profile API
type Profile struct {
	meta_api.TypeMeta   `json:",inline"`
	meta_api.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProfileSpec   `json:"spec,omitempty"`
	Status ProfileStatus `json:"status,omitempty"`
}

type ProfileSpec struct {
	Namespaces     string                  `json:"namespaces,omitempty"`
	Selector       *meta_api.LabelSelector `json:"selector,omitempty"`
	ServiceAccount string                  `json:"serviceaccount"`
	Project        string                  `json:"project"`
}

// +kubebuilder:object:root=true

// ProfileList contains a list of gcpworkloadPolicyProfile
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
