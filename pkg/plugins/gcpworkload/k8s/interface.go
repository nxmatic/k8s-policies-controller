package k8s

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	cnrm_iam_api "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/iam/v1beta1"
	"github.com/go-logr/logr"
	gcpworkload_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpworkload/v1alpha1"
	k8s_spi "github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/k8s"
	k8s_core_api "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	k8s_meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8s_labels "k8s.io/apimachinery/pkg/labels"
	k8s_selection "k8s.io/apimachinery/pkg/selection"
	k8s_types "k8s.io/apimachinery/pkg/types"
)

type (
	Interface struct {
		k8s_spi.Interface
	}

	Patch func(sa *k8s_core_api.ServiceAccount, profile *Profile) ([]byte, error)

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

func NewInterface(mgr manager.Manager) *Interface {
	spi := k8s_spi.NewInterface(mgr)
	k8s := &Interface{
		Interface: *spi,
	}
	k8s.Interface.Outer = k8s
	return k8s
}

func (k8s *Interface) ResolveProfile(meta k8s_meta_api.ObjectMeta) (*gcpworkload_api.Profile, error) {
	if resolved, err := k8s.Interface.ResolveProfile(meta, k8s.newProfileCollector()); err != nil {
		return nil, err
	} else {
		profile := resolved.(ProfileAdaptor)
		return profile.Profile, nil
	}
}

func (k8s *Interface) GetProfile(name string) (*Profile, error) {
	key := client.ObjectKey{
		Name: name,
	}
	profile := &Profile{}
	if err := k8s.GetClient().Get(context.Background(), key, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

func (k8s *Interface) SynchronizeServiceAccounts(patch Patch, mainLogger logr.Logger) error {
	list := &k8s_core_api.ServiceAccountList{}
	requirement, _ := k8s_labels.NewRequirement(gcpworkload_api.WatchKey.String(),
		k8s_selection.NotEquals,
		[]string{"true"})
	selector := k8s_labels.Everything().Add(*requirement)
	if err := k8s.GetClient().List(context.Background(), list, client.MatchingLabelsSelector{selector}); err != nil {
		return err
	}

	for _, item := range list.Items {
		if item.Namespace == "kube-system" || item.Namespace == k8s.DefaultMetaSupplier.Get().Namespace {
			continue
		}
		logger := mainLogger.WithValues("namespace", item.Namespace, "service account", item.Name)
		if profile, err := k8s.ResolveProfile(item.ObjectMeta); err == nil {
			logger = logger.WithValues("profile", profile.Name)
			logger.Info("patching")
			if err := k8s.CreateIAMPolicyMember(profile, &item); err != nil {
				logger.Error(err, "cannot create iam policy member")
			}
			if bytes, err := patch(&item, profile); err == nil {
				if err := k8s.GetClient().Patch(context.Background(), &item, client.RawPatch(k8s_types.JSONPatchType, bytes)); err != nil {
					logger.Error(err, "cannot patch")
				} else {

				}
			}
		} else {
			if err != k8s_spi.NoProfile {
				logger.Error(err, "cannot resolve profile for")
			}
			logger.Info("no profile matching")
		}
	}

	return nil
}

func (k8s *Interface) CreateIAMPolicyMember(profile *Profile, sa *k8s_core_api.ServiceAccount) error {
	memberOf := fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/%s]",
		profile.Spec.Project,
		sa.ObjectMeta.Namespace, sa.ObjectMeta.Name)
	external := fmt.Sprintf("projects/%s/serviceAccounts/%s@%s.iam.gserviceaccount.com",
		profile.Spec.Project, sa.Name, profile.Spec.Project)
	member := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "iam.cnrm.cloud.google.com/v1beta1",
			"kind":       "IAMPolicyMember",
			"metadata": map[string]interface{}{
				"name":      fmt.Sprintf("%s-%s-%s", profile.Name, sa.Namespace, sa.Name),
				"namespace": sa.Namespace,
				"labels": map[string]interface{}{
					gcpworkload_api.WatchKey.String(): "true",
				},
				"annotations": map[string]interface{}{
					gcpworkload_api.ProfilesKey.String(): profile.Name,
				},
			},
			"spec": map[string]interface{}{
				"member": memberOf,
				"resourceRef": map[string]interface{}{
					"apiVersion": "iam.cnrm.cloud.google.com/v1beta1",
					"external":   external,
					"kind":       "IAMServiceAccount",
				},
				"role": "roles/iam.workloadIdentityUser",
			},
		},
	}

	if err := k8s.GetClient().Create(context.Background(), member); err != nil {
		if !k8s_errors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}

func (k8s *Interface) DeleteIAMPolicyMember(profile *Profile, sa *k8s_core_api.ServiceAccount) error {
	return k8s.GetClient().DeleteAllOf(context.Background(),
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
