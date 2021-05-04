// +k8s:deepcopy-gen=package,register
// +kubebuilder:object:generate=true
// +groupName=gcpworkload.policies.nuxeo.io
// +versionName=v1alpha1
package v1alpha1

import (
	meta_policy_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/meta/v1alpha1"

	meta_k8s_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

type (
	TypeValue      = meta_policy_api.TypeValue
	KeyValue       = meta_policy_api.KeyValue
	ResourceKind   = meta_policy_api.ResourceKind
	AnnotationName = meta_policy_api.AnnotationName
)

const (
	ProfilesKey KeyValue                     = "gcpworkload.policies.nuxeo.io/profiles"
	TypeKey     KeyValue                     = "gcpworkload.policies.nuxeo.io/type"
	WatchKey    KeyValue                     = "gcpworkload.policies.nuxeo.io/watch"
	ProfileKind meta_policy_api.ResourceKind = "Profile"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "gcpworkload.policies.nuxeo.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	ProfilesResource = SchemeGroupVersion.WithResource("profiles")
)

const ()

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// gcpworkloadProfile is the schema for the gcpworkloadPolicy profile API
type Profile struct {
	meta_k8s_api.TypeMeta   `json:",inline"`
	meta_k8s_api.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProfileSpec   `json:"spec,omitempty"`
	Status ProfileStatus `json:"status,omitempty"`
}

type ProfileSpec struct {
	Selector meta_policy_api.ObjectSelector `json:"selector,omitempty"`

	ServiceAccount string `json:"serviceaccount"`
	Project        string `json:"project"`
}

// +kubebuilder:object:root=true

// ProfileList contains a list of gcpworkloadPolicyProfile
type ProfileList struct {
	meta_k8s_api.TypeMeta `json:",inline"`
	meta_k8s_api.ListMeta `json:"metadata,omitempty"`
	Items                 []Profile `json:"items"`
}

// ProfileStatus the status
type ProfileStatus struct {
}

func init() {
	SchemeBuilder.Register(&Profile{}, &ProfileList{})
}
