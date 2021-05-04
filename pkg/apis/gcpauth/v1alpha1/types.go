// +k8s:deepcopy-gen=package,register
// +kubebuilder:object:generate=true
// +groupName=gcpauth.policies.nuxeo.io
// +versionName=v1alpha1
package v1alpha1

import (
	meta_policy_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/meta/v1alpha1"
	meta_k8s_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

type (
	KeyValue       = meta_policy_api.KeyValue
	ResourceKind   = meta_policy_api.ResourceKind
	AnnotationName = meta_policy_api.AnnotationName
)

const (
	ProfilesKey KeyValue     = "gcpauth.policies.nuxeo.io/profiles"
	TypeKey     KeyValue     = "gcpauth.policies.nuxeo.io/type"
	WatchKey    KeyValue     = "gcpauth.policies.nuxeo.io/watch"
	ProfileKind ResourceKind = "Profile"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "gcpauth.policies.nuxeo.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	ProfilesResource = SchemeGroupVersion.WithResource("profiles")
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// GCPAuthProfile is the schema for the gcpauth.policy profile API
type Profile struct {
	meta_k8s_api.TypeMeta   `json:",inline"`
	meta_k8s_api.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProfileSpec   `json:"spec,omitempty"`
	Status ProfileStatus `json:"status,omitempty"`
}

type ProfileSpec struct {
	Selector meta_policy_api.ObjectSelector `json:"selector,omitempty"`

	Datasource SecretRef `json:"datasource,omitempty"`
}

type SecretRef struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

// +kubebuilder:object:root=true

// ProfileList contains a list of gcpauth.policyProfile
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
