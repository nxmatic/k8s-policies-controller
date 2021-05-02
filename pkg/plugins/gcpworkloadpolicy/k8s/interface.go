package k8s

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/dynamic"

	cnrm_iam_api "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/iam/v1beta1"
	gcpworkload_api "github.com/nuxeo/k8s-policy-controller/apis/gcpworkloadpolicyprofile/v1alpha1"
	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8s_spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"
	"k8s.io/apimachinery/pkg/runtime"
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
	ProfileKey = gcpworkload_api.ProfileKey
)

var (
	IAMPolicyMembersResource = cnrm_iam_api.SchemeGroupVersion.WithResource("iampolicymembers")
	ProfilesResource         = gcpworkload_api.ProfilesResource
)

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

func (s *Interface) ResolveProfile(namespace *meta_api.ObjectMeta, resource *meta_api.ObjectMeta) (*gcpworkload_api.Profile, error) {
	names := s.MergeAnnotation(gcpworkload_api.ProfileKey.String(), resource, namespace, s.DefaultMeta)
	for _, name := range names {
		profile, err := s.GetProfile(name)
		if err != nil {
			return nil, errors.New("cannot retrieve profile " + name)
		}
		if profile.Spec.Selector != nil {
			selector, err := meta_api.LabelSelectorAsSelector(profile.Spec.Selector)
			if err != nil {
				return nil, err
			}
			if !selector.Matches(labels.Set(resource.Labels)) {
				continue
			}
		}
		return profile, nil
	}

	return nil, errors.New("Annotation not found")

}

func (s *Interface) GetProfile(name string) (*Profile, error) {
	resp, err := s.Interface.Resource(ProfilesResource).Get(context.TODO(), name, meta_api.GetOptions{})
	if err != nil {
		return nil, err
	}
	profile := &Profile{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(resp.UnstructuredContent(), profile)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s *Interface) CreateIAMPolicyMember(profile *Profile, sa *core_api.ServiceAccount) error {
	member := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": cnrm_iam_api.SchemeGroupVersion.String(),
			"kind":       "IAMPolicyMember",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("%s-%s-%s", profile.ObjectMeta.Name, sa.ObjectMeta.Namespace, sa.ObjectMeta.Name),
				"labels": map[string]interface{}{
					gcpworkload_api.ProfileKey.String(): profile.ObjectMeta.Name,
					gcpworkload_api.WatchKey.String():   "true",
				},
			},
			"spec": map[string]interface{}{
				"member": fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/%s]", profile.Spec.Project, sa.ObjectMeta.Namespace, sa.ObjectMeta.Name),
				"resourceRef": map[string]interface{}{
					"external":   fmt.Sprintf("projects/%s/serviceAccounts/%s@%s.iam.gserviceaccount.com", profile.Spec.Project, profile.Spec.ServiceAccount, profile.Spec.Project),
					"apiVersion": cnrm_iam_api.SchemeGroupVersion.String(),
					"kind":       "IAMServiceAccount",
				},
				"role": "roles/iam.workloadIdentityUser",
			},
		},
	}
	_, err :=
		s.Resource(IAMPolicyMembersResource).
			Namespace(sa.ObjectMeta.Namespace).
			Create(context.TODO(), &member, meta_api.CreateOptions{})
	if err != nil {
		return err
	}

	return nil

}

func (s *Interface) DeleteIAMPolicyMember(profile *Profile, sa *core_api.ServiceAccount) error {
	return s.Resource(IAMPolicyMembersResource).
		Namespace(sa.ObjectMeta.Namespace).
		DeleteCollection(context.TODO(),
			meta_api.DeleteOptions{},
			meta_api.ListOptions{
				LabelSelector: labels.Set{
					ProfileKey.String(): profile.ObjectMeta.Name,
				}.String(),
			})
}
