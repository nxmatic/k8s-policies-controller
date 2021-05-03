package k8s

import (
	meta_policy_api "github.com/nuxeo/k8s-policy-controller/pkg/apis/meta/v1alpha1"
	node_policy_api "github.com/nuxeo/k8s-policy-controller/pkg/apis/node/v1alpha1"
	k8s_spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"
)

type ProfileSupplier struct {
	Interface
}

func (ProfileSupplier) Key() string {
	return node_policy_api.ProfilesKey.String()
}

func (supplier ProfileSupplier) Get(name string) (k8s_spi.ProfileGetter, error) {
	if profile, err := supplier.GetProfile(name); err != nil {
		return nil, err
	} else {
		adaptor := ProfileAdaptor{
			profile,
		}
		return adaptor, nil
	}
}

type ProfileAdaptor struct {
	*node_policy_api.Profile
}

func (profile ProfileAdaptor) GetSelector() meta_policy_api.ObjectSelector {
	return profile.Spec.Selector
}

func (profile ProfileAdaptor) GetName() string {
	return profile.Name
}
