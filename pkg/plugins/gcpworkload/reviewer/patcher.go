package reviewer

import (
	"fmt"

	gcpworkload_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpworkload/v1alpha1"
	meta_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/meta/v1alpha1"

	core_api "k8s.io/api/core/v1"

	spi "github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/reviewer"
)

type (
	ServiceaccountPatcher struct {
		*core_api.ServiceAccount
		*gcpworkload_api.Profile
		Patch []spi.PatchOperation
	}
)

func NewServiceAccountPatcher(sa *core_api.ServiceAccount, profile *gcpworkload_api.Profile) *ServiceaccountPatcher {
	return &ServiceaccountPatcher{
		ServiceAccount: sa,
		Profile:        profile,
	}
}

func (p *ServiceaccountPatcher) Create() ([]spi.PatchOperation, error) {
	p.Patch = make([]spi.PatchOperation, 0, 2)
	p.Patch = append(p.Patch, p.addLabelsPatch())
	p.Patch = append(p.Patch, p.addAnnotationsPatch())
	return p.Patch, nil
}

func (p *ServiceaccountPatcher) addLabelsPatch() spi.PatchOperation {
	if p.ServiceAccount.Labels == nil {
		return spi.PatchOperation{
			Op:   "add",
			Path: "/metadata/labels",
			Value: map[string]string{
				gcpworkload_api.ProfilesKey.String(): p.Profile.Name,
				gcpworkload_api.WatchKey.String():    "true",
			},
		}
	}
	return spi.PatchOperation{
		Op:    "add",
		Path:  "/metadata/labels/" + gcpworkload_api.WatchKey.Encoded(),
		Value: p.Profile.Name,
	}
}

func (p *ServiceaccountPatcher) addAnnotationsPatch() spi.PatchOperation {
	var saKey meta_api.KeyValue = "iam.gke.io/gcp-service-account"

	value := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", p.Spec.ServiceAccount, p.Spec.Project)
	if p.ServiceAccount.ObjectMeta.Annotations == nil {
		return spi.PatchOperation{
			Op:   "add",
			Path: "/metadata/annotations",
			Value: map[string]string{
				saKey.String(): value,
			},
		}
	}
	return spi.PatchOperation{
		Op:    "add",
		Path:  "/metadata/annotations/" + saKey.Encoded(),
		Value: value,
	}
}
