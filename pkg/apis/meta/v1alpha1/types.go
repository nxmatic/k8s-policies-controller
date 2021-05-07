// +k8s:deepcopy-gen=package,register
// +kubebuilder:object:generate=true
// +versionName=v1alpha1
package v1alpha1

import (
	"strings"

	meta_k8s_api "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ResourceKind typed identifiers

type ResourceKind string

func (name ResourceKind) String() string {
	return string(name)
}

// KeyValue typed identifiers

type KeyValue string

func (name KeyValue) String() string {
	return string(name)
}

func (name KeyValue) Encoded() string {
	return strings.ReplaceAll(name.String(), "/", "~1")
}

// TypeValye typed identifiers

type TypeValue string

func (name TypeValue) String() string {
	return string(name)
}

// AnnotationName typed identifiers
type AnnotationName string

func (name AnnotationName) String() string {
	return string(name)
}

const (
	WatchKey KeyValue = "policies.nuxeo.io/watch"
)

// ObjectSelector defines the criterias for selecting the targeted objects
type ObjectSelector struct {
	Namespaces string                      `json:"namespaces,omitempty"`
	Objects    *meta_k8s_api.LabelSelector `json:"objects,omitempty"`
}
