// +k8s:deepcopy-gen=package,register
// +kubebuilder:object:generate=true
// +groupName=node.policies.nuxeo.io
// +versionName=v1alpha1
package v1alpha1

import (
	meta_policy_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/meta/v1alpha1"
	k8s_core_api "k8s.io/api/core/v1"
	k8s_meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

type (
	KeyValue       = meta_policy_api.KeyValue
	ResourceKind   = meta_policy_api.ResourceKind
	AnnotationName = meta_policy_api.AnnotationName
)

const (
	ProfilesKey KeyValue     = "node.policies.nuxeo.io/profiles"
	TypeKey     KeyValue     = "node.policies.nuxeo.io/type"
	WatchKey    KeyValue     = "node.policies.nuxeo.io/watch"
	ProfileKind ResourceKind = "Profile"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "node.policies.nuxeo.io", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme

	NodepolicyprofilesResource = SchemeBuilder.GroupVersion.WithResource("profiles")
	PodsResource               = k8s_core_api.SchemeGroupVersion.WithResource("pods")
	NamespacesResource         = k8s_core_api.SchemeGroupVersion.WithResource("namespaces")
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProfileSpec defines the desired state of NodePolicyProfile
type ProfileSpec struct {
	Selector     meta_policy_api.ObjectSelector `json:"selector,omitempty"`
	Tolerations  []k8s_core_api.Toleration      `json:"tolerations,omitempty"`
	NodeAffinity k8s_core_api.NodeAffinity      `json:"nodeAffinity,omitempty"`
	NodeSelector map[string]string              `json:"nodeSelector,omitempty"`
}

// ProfileStatus defines the observed state of NodePolicyProfile
type ProfileStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// Profile is the Schema for the node.policyprofiles API
type Profile struct {
	k8s_meta_api.TypeMeta   `json:",inline"`
	k8s_meta_api.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProfileSpec   `json:"spec,omitempty"`
	Status ProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProfileList contains a list of NodePolicyProfile
type ProfileList struct {
	k8s_meta_api.TypeMeta `json:",inline"`
	k8s_meta_api.ListMeta `json:"metadata,omitempty"`
	Items                 []Profile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Profile{}, &ProfileList{})
}
