package k8s

import (
	"context"

	node_policy_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/node/v1alpha1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	k8s_spi "github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/k8s"
)

var (
	NodepolicyprofilesResources = node_policy_api.NodepolicyprofilesResource
)

type (
	Interface struct {
		k8s_spi.Interface
	}

	Profile = node_policy_api.Profile
	Spec    = node_policy_api.ProfileSpec
	Status  = node_policy_api.ProfileStatus

	KeyValue = node_policy_api.KeyValue
)

func NewInterface(mgr manager.Manager) (*Interface, error) {
	spi, err := k8s_spi.NewInterface(mgr)
	if err != nil {
		return nil, err
	}
	return &Interface{
		Interface: *spi,
	}, nil
}

func (s *Interface) ResolveProfile(meta meta_api.ObjectMeta) (*node_policy_api.Profile, error) {
	if resolved, err := s.Interface.ResolveProfile(meta, s.newProfileCollector()); err != nil {
		return nil, err
	} else {
		profile := resolved.(ProfileAdaptor)
		return profile.Profile, nil
	}
}

func (s *Interface) GetProfile(name string) (*node_policy_api.Profile, error) {
	key := client.ObjectKey{
		Name: name,
	}
	profile := &node_policy_api.Profile{}
	err := s.Manager.GetClient().Get(context.Background(), key, profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (k8s Interface) newProfileCollector() k8s_spi.ProfileCollector {
	return k8s_spi.NewProfileCollector(
		k8s.newProfileSupplier())
}

func (k8s Interface) newProfileSupplier() k8s_spi.ProfileSupplier {
	return ProfileSupplier{
		k8s,
	}
}
