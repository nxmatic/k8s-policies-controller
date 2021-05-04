package k8s

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cnrm_iam_api "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/iam/v1beta1"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	gcpworkload_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpworkload/v1alpha1"
	gcpworkload_policy_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpworkload/v1alpha1"
	k8s_spi "github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/k8s"
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	Interface struct {
		k8s_spi.Interface
	}

	Profile = gcpworkload_api.Profile
	Spec    = gcpworkload_api.ProfileSpec
	Status  = gcpworkload_api.ProfileStatus

	TypeValue = gcpworkload_api.TypeValue
	KeyValue  = gcpworkload_api.KeyValue
)

const (
	ProfileKey = gcpworkload_api.ProfilesKey
)

var (
	IAMPolicyMembersResource = cnrm_iam_api.SchemeGroupVersion.WithResource("iampolicymembers")
	ProfilesResource         = gcpworkload_api.ProfilesResource
)

func NewInterface(mgr manager.Manager) (*Interface, error) {
	if spi, err := k8s_spi.NewInterface(mgr); err != nil {
		return nil, err
	} else {
		return &Interface{
			Interface: *spi,
		}, nil
	}
}

func (s *Interface) ResolveProfile(meta meta_api.ObjectMeta) (*gcpworkload_policy_api.Profile, error) {
	if resolved, err := s.Interface.ResolveProfile(meta, s.newProfileCollector()); err != nil {
		return nil, err
	} else {
		profile := resolved.(ProfileAdaptor)
		return profile.Profile, nil
	}
}

func (s *Interface) GetProfile(name string) (*Profile, error) {
	key := client.ObjectKey{
		Name: name,
	}
	profile := &Profile{}
	if err := s.GetClient().Get(context.Background(), key, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (s *Interface) CreateIAMPolicyMember(profile *Profile, sa *core_api.ServiceAccount) error {
	memberOf := fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/%s]", profile.Spec.Project, sa.ObjectMeta.Namespace, sa.ObjectMeta.Name)
	member := &cnrm_iam_api.IAMPolicyMember{
		ObjectMeta: meta_api.ObjectMeta{
			Name: fmt.Sprintf("%s-%s-%s", profile.ObjectMeta.Name, sa.ObjectMeta.Namespace, sa.ObjectMeta.Name),
			Labels: map[string]string{
				gcpworkload_api.ProfilesKey.String(): profile.ObjectMeta.Name,
				gcpworkload_api.WatchKey.String():    "true",
			},
		},
		Spec: cnrm_iam_api.IAMPolicyMemberSpec{
			Member: &memberOf,
			ResourceRef: v1alpha1.ResourceRef{
				External: fmt.Sprintf("projects/%s/serviceAccounts/%s@%s.iam.gserviceaccount.com", profile.Spec.Project, profile.Spec.ServiceAccount, profile.Spec.Project),
			},
			Role: "roles/iam.workloadIdentityUser",
		},
	}

	if err := s.GetClient().Create(context.Background(), member); err != nil {
		return err
	}

	return nil
}

func (s *Interface) DeleteIAMPolicyMember(profile *Profile, sa *core_api.ServiceAccount) error {
	return s.GetClient().DeleteAllOf(context.Background(),
		&cnrm_iam_api.IAMPolicyMember{},
		client.InNamespace(sa.Namespace),
		&client.MatchingLabels{
			ProfileKey.String(): profile.ObjectMeta.Name,
		})

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
