package reviewer

import (
	"encoding/json"

	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpworkloadpolicy/k8s"

	gcpworkload_api "github.com/nuxeo/k8s-policy-controller/apis/gcpworkloadpolicyprofile/v1alpha1"
	admission_api "k8s.io/api/admission/v1"
	core_api "k8s.io/api/core/v1"

	spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/reviewer"
)

type (
	RequestedServiceAccountStage struct {
		k8s.Interface
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

func (s *RequestedServiceAccountStage) RequestedObject(o *spi.GivenStage) *RequestedServiceAccountStage {
	s.GivenStage = o
	s.Interface = k8s.Interface{Interface: o.Interface}
	return s
}

func (s *RequestedServiceAccountStage) The() *RequestedServiceAccountStage {
	return s
}

func (s *RequestedServiceAccountStage) And() *RequestedServiceAccountStage {
	return s
}

func (r *RequestedServiceAccountStage) RequestedKind() *RequestedKindStage {
	return &RequestedKindStage{r}
}

func (s *RequestedKindStage) Or() *RequestedKindStage {
	return s
}

func (s *RequestedKindStage) IsAServiceAccount() *RequestedKindStage {
	err := json.Unmarshal(s.AdmissionRequest.Object.Raw, &s.ServiceAccount)
	if err != nil {
		s.Allow(nil)
		return s
	}
	s.Logger = s.Logger.WithValues("name", s.ServiceAccount.ObjectMeta.Name)

	return s
}

func (s *RequestedKindStage) End() *RequestedServiceAccountStage {
	return s.RequestedServiceAccountStage
}

func (s *RequestedServiceAccountStage) RequestedProfile() *RequestedProfileStage {
	return &RequestedProfileStage{s, nil}
}

func (s *RequestedProfileStage) Applies() *RequestedProfileStage {
	if !s.CanContinue() {
		return s
	}

	s.Namespace, s.Error = s.Interface.GetNamespace(s.AdmissionRequest.Namespace)
	if s.Error != nil {
		s.Allow(s.Error)
		return s
	}

	profile, err := s.Interface.ResolveProfile(&s.Namespace.ObjectMeta, &s.ServiceAccount.ObjectMeta)
	if err != nil {
		s.Allow(err)
		return s
	}
	s.Profile = *profile

	switch s.Operation {
	case admission_api.Create:
		s.Error = s.Interface.CreateIAMPolicyMember(&s.Profile, &s.ServiceAccount)
		if s.Error != nil {
			s.Fail(s.Error)
			return s
		}
	case admission_api.Delete:
		s.Error = s.Interface.DeleteIAMPolicyMember(&s.Profile, &s.ServiceAccount)
		s.Allow(s.Error)
	}

	s.Logger = s.Logger.WithValues("profile", s.Profile.ObjectMeta.Name)

	return s
}

func (s *RequestedProfileStage) And() *RequestedProfileStage {
	return s
}

func (s *RequestedProfileStage) The() *RequestedProfileStage {
	return s
}

func (s *RequestedProfileStage) End() *RequestedServiceAccountStage {
	return s.RequestedServiceAccountStage
}

func (s *RequestedServiceAccountStage) End() *spi.WhenStage {
	return &spi.WhenStage{
		GivenStage: s.GivenStage,
		Patcher: &serviceaccountPatcher{
			ServiceAccount: &s.ServiceAccount,
			Profile:        &s.Profile,
		}}
}
