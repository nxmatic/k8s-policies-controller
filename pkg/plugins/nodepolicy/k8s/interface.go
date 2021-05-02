package k8s

import (
	"context"
	"errors"

	nodepolicy_api "github.com/nuxeo/k8s-policy-controller/apis/nodepolicyprofile/v1alpha1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"

	k8s_spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	NodepolicyprofilesResources = nodepolicy_api.NodepolicyprofilesResource
)

type Interface struct {
	k8s_spi.Interface
}

func NewInterface(client dynamic.Interface) (*Interface, error) {
	spi, err := k8s_spi.NewInterface(client)
	if err != nil {
		return nil, err
	}
	k8s := &Interface{
		Interface: *spi,
	}
	k8s.SetConcreteRef(k8s)
	return k8s, nil
}

func (s *Interface) ResolveProfile(namespace *meta_api.ObjectMeta, resource *meta_api.ObjectMeta) (*nodepolicy_api.Profile, error) {
	names := s.MergeAnnotation(nodepolicy_api.AnnotationPolicyProfile.String(), resource, namespace, s.DefaultMeta)
	for _, name := range names {
		profile, err := s.GetProfile(name)
		if err != nil {
			return nil, errors.New("cannot retrieve profile " + name)
		}
		if profile.Spec.PodSelector != nil {
			selector, err := meta_api.LabelSelectorAsSelector(profile.Spec.PodSelector)
			if err != nil {
				return nil, err
			}
			if !selector.Matches(labels.Set(resource.Labels)) {
				continue
			}
		}
		return profile, nil
	}
	return nil, errors.New("no profile")
}

func (s *Interface) GetProfile(name string) (*nodepolicy_api.Profile, error) {
	resp, err := s.Interface.Resource(NodepolicyprofilesResources).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	profile := &nodepolicy_api.Profile{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}
