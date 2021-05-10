package reviewer

import (
	"encoding/json"

	gcpworkload_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpworkload/v1alpha1"
	admission_api "k8s.io/api/admission/v1"
	core_api "k8s.io/api/core/v1"

	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/gcpworkload/k8s"
	spi "github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/reviewer"
)

type (
	RequestedServiceAccountStage struct {
		*k8s.Interface
		*spi.GivenStage
		core_api.ServiceAccount
		gcpworkload_api.Profile
	}
	RequestedKindStage struct {
		*RequestedServiceAccountStage
	}
	RequestedProfileStage struct {
		*RequestedServiceAccountStage
		*core_api.Namespace
	}
)

func Given() *RequestedServiceAccountStage {
	return &RequestedServiceAccountStage{}
}

func (stage *RequestedServiceAccountStage) RequestedObject(o *spi.GivenStage) *RequestedServiceAccountStage {
	stage.GivenStage = o
	stage.Interface = o.Outer.(*k8s.Interface)
	return stage
}

func (stage *RequestedServiceAccountStage) The() *RequestedServiceAccountStage {
	return stage
}

func (stage *RequestedServiceAccountStage) And() *RequestedServiceAccountStage {
	return stage
}

func (r *RequestedServiceAccountStage) RequestedKind() *RequestedKindStage {
	return &RequestedKindStage{r}
}

func (stage *RequestedKindStage) Or() *RequestedKindStage {
	return stage
}

func (stage *RequestedKindStage) IsAServiceAccount() *RequestedKindStage {
	if stage.Object.Raw == nil {
		return stage
	}
	err := json.Unmarshal(stage.AdmissionRequest.Object.Raw, &stage.ServiceAccount)
	if err != nil {
		stage.Allow(nil)
		return stage
	}
	return stage
}

func (stage *RequestedKindStage) End() *RequestedServiceAccountStage {
	return stage.RequestedServiceAccountStage
}

func (stage *RequestedServiceAccountStage) RequestedProfile() *RequestedProfileStage {
	return &RequestedProfileStage{stage, nil}
}

func (stage *RequestedProfileStage) Applies() *RequestedProfileStage {
	if !stage.CanContinue() {
		return stage
	}

	if _, ok := stage.ServiceAccount.ObjectMeta.Annotations["iam.gke.io/gcp-service-account"]; ok {
		stage.Allow(nil)
		return stage
	}

	profile, err := stage.Interface.ResolveProfile(stage.Namespace.ObjectMeta)
	if err != nil {
		stage.Allow(err)
		return stage
	}
	stage.Profile = *profile

	switch stage.Operation {
	case admission_api.Create:
		stage.Error = stage.Interface.CreateIAMPolicyMember(&stage.Profile, &stage.ServiceAccount)
		if stage.Error != nil {
			stage.Fail(stage.Error)
			return stage
		}
	case admission_api.Delete:
		stage.Error = stage.Interface.DeleteIAMPolicyMember(&stage.Profile, &stage.ServiceAccount)
		stage.Allow(stage.Error)
	}

	stage.Logger = stage.Logger.WithValues("profile", stage.Profile.ObjectMeta.Name)

	return stage
}

func (stage *RequestedProfileStage) And() *RequestedProfileStage {
	return stage
}

func (stage *RequestedProfileStage) The() *RequestedProfileStage {
	return stage
}

func (stage *RequestedProfileStage) End() *RequestedServiceAccountStage {
	return stage.RequestedServiceAccountStage
}

func (stage *RequestedServiceAccountStage) End() *spi.WhenStage {
	patcher := NewServiceAccountPatcher(&stage.ServiceAccount, &stage.Profile)
	return &spi.WhenStage{
		GivenStage: stage.GivenStage,
		Patcher:    patcher,
	}
}
