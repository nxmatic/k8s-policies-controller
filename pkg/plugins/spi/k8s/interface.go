package k8s

import (
	"context"
	"errors"
	"regexp"

	k8s_core_api "k8s.io/api/core/v1"
	k8s_meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Interface for interacting with kube resources
type (
	Interface struct {
		manager.Manager
		DefaultMeta k8s_meta_api.ObjectMeta
	}
)

var (
	PodsResource            = k8s_core_api.SchemeGroupVersion.WithResource("pods")
	PodsKind                = k8s_core_api.SchemeGroupVersion.WithKind("Pod")
	NamespacesResource      = k8s_core_api.SchemeGroupVersion.WithResource("namespaces")
	NamespacesKind          = k8s_core_api.SchemeGroupVersion.WithKind("Namespace")
	ServiceaccountsResource = k8s_core_api.SchemeGroupVersion.WithResource("serviceaccounts")
	ServiceaccountsKind     = k8s_core_api.SchemeGroupVersion.WithKind("Serviceaccount")
)

func NewInterface(mgr manager.Manager) (*Interface, error) {
	return &Interface{
		Manager: mgr,
	}, nil
}

func (k8s *Interface) GetNamespace(name string) (*k8s_core_api.Namespace, error) {
	key := client.ObjectKey{
		Name: name,
	}
	ns := &k8s_core_api.Namespace{}
	err := k8s.Manager.GetClient().Get(context.Background(), key, ns)
	if err != nil {
		return nil, err
	}
	return ns, nil
}

func (k8s *Interface) GetServiceAccount(name string, namespace string) (*k8s_core_api.ServiceAccount, error) {
	key := client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}
	sa := &k8s_core_api.ServiceAccount{}
	err := k8s.Manager.GetClient().Get(context.Background(), key, sa)
	if err != nil {
		return nil, err
	}
	return sa, nil
}

func (s *Interface) ResolveProfile(meta k8s_meta_api.ObjectMeta, collector ProfileCollector) (ProfileGetter, error) {
	if err := s.ResolveProfiles(meta, &collector); err != nil {
		return nil, err
	}
	for _, profile := range collector.Profiles() {
		selector := profile.GetSelector()
		if selector.Namespaces != "" {
			if ok, err := regexp.MatchString(selector.Namespaces, meta.Name); err != nil {
				return nil, errors.New("Cannot evaluate " + selector.Namespaces + " on profile " + profile.GetName())
			} else if !ok {
				continue
			}
		}
		if objectsSelector := selector.Objects; objectsSelector != nil {
			selector, err := k8s_meta_api.LabelSelectorAsSelector(objectsSelector)
			if err != nil {
				return nil, err
			}
			if !selector.Matches(labels.Set(meta.Labels)) {
				continue
			}
		}
		return profile, nil
	}
	return nil, errors.New("profile not found")
}

func (k8s *Interface) ResolveProfiles(resource k8s_meta_api.ObjectMeta, collector *ProfileCollector) error {
	defaultNS, err := k8s.GetNamespace(resource.Namespace)
	if err != nil {
		return err
	}

	collector.addNames(resource)
	collector.addNames(defaultNS.ObjectMeta)
	collector.addNames(k8s.DefaultMeta)

	return nil
}
