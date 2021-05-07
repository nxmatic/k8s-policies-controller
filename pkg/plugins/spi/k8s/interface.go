package k8s

import (
	"context"
	"errors"
	"log"
	"regexp"
	"sync"

	meta_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/meta/v1alpha1"
	k8s_core_api "k8s.io/api/core/v1"
	k8s_meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s_labels "k8s.io/apimachinery/pkg/labels"
	k8s_selection "k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Interface for interacting with kube resources
type (
	Interface struct {
		manager.Manager
		*DefaultMetaSupplier
	}

	DefaultMetaSupplier struct {
		k8s   *Interface
		meta  k8s_meta_api.ObjectMeta
		init  sync.Once
		write sync.Mutex
	}
)

var (
	PodsResource            = k8s_core_api.SchemeGroupVersion.WithResource("pods")
	PodsKind                = k8s_core_api.SchemeGroupVersion.WithKind("Pod")
	NamespacesResource      = k8s_core_api.SchemeGroupVersion.WithResource("namespaces")
	NamespacesKind          = k8s_core_api.SchemeGroupVersion.WithKind("Namespace")
	ServiceaccountsResource = k8s_core_api.SchemeGroupVersion.WithResource("serviceaccounts")
	ServiceaccountsKind     = k8s_core_api.SchemeGroupVersion.WithKind("Serviceaccount")

	NoProfile = errors.New("no profile")
)

func NewInterface(mgr manager.Manager) *Interface {
	k8s := &Interface{
		Manager:             mgr,
		DefaultMetaSupplier: &DefaultMetaSupplier{},
	}
	k8s.DefaultMetaSupplier = &DefaultMetaSupplier{
		k8s: k8s,
	}
	return k8s
}

func (k8s *Interface) GetNamespace(name string) (*k8s_core_api.Namespace, error) {
	key := client.ObjectKey{
		Name: name,
	}
	ns := &k8s_core_api.Namespace{}
	err := k8s.GetClient().Get(context.Background(), key, ns)
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
	err := k8s.GetClient().Get(context.Background(), key, sa)
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
			if ok, err := regexp.MatchString(selector.Namespaces, meta.Namespace); err != nil {
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
			if !selector.Matches(k8s_labels.Set(meta.Labels)) {
				continue
			}
		}
		return profile, nil
	}
	return nil, NoProfile
}

func (k8s *Interface) ResolveProfiles(resource k8s_meta_api.ObjectMeta, collector *ProfileCollector) error {
	ns, err := k8s.GetNamespace(resource.Namespace)
	if err != nil {
		return err
	}

	collector.addNames(resource)
	collector.addNames(ns.ObjectMeta)
	collector.addNames(k8s.DefaultMetaSupplier.Get())

	return nil
}

func (supplier *DefaultMetaSupplier) Get() k8s_meta_api.ObjectMeta {
	supplier.init.Do(func() {
		list := &k8s_core_api.NamespaceList{}
		requirement, _ := k8s_labels.NewRequirement(meta_api.WatchKey.String(),
			k8s_selection.Equals,
			[]string{"true"})
		selector := k8s_labels.NewSelector().Add(*requirement)
		if err := supplier.k8s.GetClient().List(context.Background(), list, client.MatchingLabelsSelector{selector}); err != nil {
			log.Fatal("Cannot fetch default policies namespace: ", err)
		}
		if len(list.Items) == 0 {
			supplier.k8s.GetLogger().Info("no policies namespace matching", "selector", selector.String())
			return
		}
		if len(list.Items) > 1 {
			supplier.k8s.GetLogger().Info("multiple policies namespace not supported, peeking first in list", "namespace", list.Items[0].Name)
		}
		supplier.Write(&list.Items[0].ObjectMeta)
	})
	return supplier.meta
}

func (supplier *DefaultMetaSupplier) Write(meta *k8s_meta_api.ObjectMeta) {
	supplier.write.Lock()
	meta.DeepCopyInto(&supplier.meta)
	supplier.write.Unlock()
}
