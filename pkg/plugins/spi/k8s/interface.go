package k8s

import (
	"context"
	"strings"

	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

// Interface for interacting with kube resources
type (
	Interface struct {
		dynamic.Interface
		DefaultMeta *meta_api.ObjectMeta
		ConcreteRef interface{}
	}
)

var (
	PodsResource            = core_api.SchemeGroupVersion.WithResource("pods")
	PodsKind                = core_api.SchemeGroupVersion.WithKind("Pod")
	NamespacesResource      = core_api.SchemeGroupVersion.WithResource("namespaces")
	NamespacesKind          = core_api.SchemeGroupVersion.WithKind("Namespace")
	ServiceaccountsResource = core_api.SchemeGroupVersion.WithResource("serviceaccounts")
	ServiceaccountsKind     = core_api.SchemeGroupVersion.WithKind("Serviceaccount")
)

func NewInterface(base dynamic.Interface) (*Interface, error) {
	k8s := &Interface{
		Interface: base,
	}
	defaultNS, err := k8s.GetNamespace("default")
	if err != nil {
		return nil, err
	}
	k8s.DefaultMeta = &defaultNS.ObjectMeta
	return k8s, nil
}

func (k8s *Interface) SetConcreteRef(ref interface{}) {
	k8s.ConcreteRef = ref
}

func (k8s *Interface) NewReplicator() *Replicator {
	return &Replicator{k8s.Interface}
}

func (k8s *Interface) GetNamespace(name string) (*core_api.Namespace, error) {
	resp, err := k8s.Interface.Resource(NamespacesResource).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	namespace := &core_api.Namespace{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), namespace)
	if err != nil {
		return nil, err
	}
	return namespace, nil
}

func (k8s *Interface) GetServiceAccount(name string, namespace string) (*core_api.ServiceAccount, error) {
	resp, err := k8s.Interface.Resource(ServiceaccountsResource).Namespace(namespace).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	serviceaccount := &core_api.ServiceAccount{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), serviceaccount)
	if err != nil {
		return nil, err
	}
	return serviceaccount, nil
}

func (k8s *Interface) MergeAnnotation(key string, metas ...*meta_api.ObjectMeta) []string {
	exists := struct{}{}
	set := make(map[string]struct{})
	for _, meta := range metas {
		if meta == nil {
			continue
		}
		if value, ok := meta.Annotations[key]; ok {
			for _, name := range strings.Split(value, ",") {
				name = strings.TrimSpace(name)
				set[name] = exists
			}
		}
	}
	names := make([]string, 0, len(set))
	for name, _ := range set {
		n := len(names)
		names = names[0 : n+1]
		names[n] = name
	}
	return names
}
