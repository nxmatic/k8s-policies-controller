package reviewer

import (
	"fmt"

	gcpworkload_api "github.com/nuxeo/k8s-policy-controller/pkg/apis/gcpworkload/v1alpha1"

	core_api "k8s.io/api/core/v1"

	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/reviewer"
)

type serviceaccountPatcher struct {
	*core_api.ServiceAccount
	*gcpworkload_api.Profile
	Patch []reviewer.PatchOperation
}

func (p *serviceaccountPatcher) Create() ([]reviewer.PatchOperation, error) {
	p.Patch = make([]reviewer.PatchOperation, 0, 2)
	p.Patch = append(p.Patch, p.addLabelsPatch())
	p.Patch = append(p.Patch, p.addAnnotationsPatch())
	return p.Patch, nil
}

func (p *serviceaccountPatcher) addLabelsPatch() reviewer.PatchOperation {
	if p.ServiceAccount.Labels == nil {
		return reviewer.PatchOperation{
			Op:   "add",
			Path: "/metadata/labels",
			Value: map[string]string{
				"gcpworkload.nuxeo.io/profile": p.Profile.Name,
			},
		}
	}
	return reviewer.PatchOperation{
		Op:    "add",
		Path:  "/metadata/labels/gcpworkload.nuxeo.io~1profile",
		Value: p.Profile.Name,
	}
}

func (p *serviceaccountPatcher) addAnnotationsPatch() reviewer.PatchOperation {
	value := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", p.Spec.ServiceAccount, p.Spec.Project)
	if p.ServiceAccount.ObjectMeta.Annotations == nil {
		return reviewer.PatchOperation{
			Op:   "add",
			Path: "/metadata/annotations",
			Value: map[string]string{
				"iam.gke.io/gcp-service-account": value,
			},
		}
	}
	return reviewer.PatchOperation{
		Op:    "add",
		Path:  "/metadata/annotations/iam.gke.io~1gcp-service-account",
		Value: value,
	}
}
