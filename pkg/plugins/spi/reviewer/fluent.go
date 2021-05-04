package reviewer

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	admission_api "k8s.io/api/admission/v1"

	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/k8s"
)

const (
	proceed   ResponseType = iota
	allowed   ResponseType = iota
	failure   ResponseType = iota
	jsonpatch ResponseType = iota
)

type (
	ResponseType int8

	BaseStage struct {
		Error error
		ResponseType
		logr.Logger
		*k8s.Interface
	}

	GivenStage struct {
		BaseStage
		*admission_api.AdmissionRequest
	}

	RequestedObjectStage struct {
		*GivenStage
	}

	PatcherStage struct {
		*GivenStage
	}

	WhenStage struct {
		*GivenStage
		Patcher
		Patch []PatchOperation
	}

	ThenStage struct {
		*WhenStage
	}

	EndStage struct {
		*ThenStage
	}
)

/**
 * Base
 */

func (s *BaseStage) setResponseType(cause error, value ResponseType) *BaseStage {
	if cause != nil {
		s.Error = cause
	}
	s.ResponseType = value
	return s
}

func (s *BaseStage) CanContinue() bool {
	return s.ResponseType == proceed
}

func (s *BaseStage) Allow(cause error) *BaseStage {
	return s.setResponseType(cause, allowed)
}

func (s *BaseStage) Fail(cause error) *BaseStage {
	return s.setResponseType(cause, failure)
}

func (s *BaseStage) JsonPatch() *BaseStage {
	return s.setResponseType(nil, jsonpatch)
}

/**
 * Given
 */

func Given(logger logr.Logger, k8s *k8s.Interface) *GivenStage {
	given := GivenStage{}
	given.Logger = logger
	given.Interface = k8s
	return &given
}

func (g *GivenStage) Request(request *admission_api.AdmissionRequest) *GivenStage {
	g.AdmissionRequest = request
	g.Logger = g.Logger.WithValues("kind", g.RequestKind, "namespace", g.AdmissionRequest.Namespace, "name", g.AdmissionRequest.Name)
	return g
}

func (g *GivenStage) An() *GivenStage {
	return g
}

func (g *GivenStage) A() *GivenStage {
	return g
}

func (g *GivenStage) The() *GivenStage {
	return g
}

func (g *GivenStage) Or() *GivenStage {
	return g
}

func (g *GivenStage) And() *GivenStage {
	return g
}

func (g *GivenStage) Group() *GivenStage {
	return g
}

func (g *GivenStage) End() *GivenStage {
	return g
}

func (g *GivenStage) RequestedObject() *RequestedObjectStage {
	return &RequestedObjectStage{g}
}

func (s *RequestedObjectStage) NamespaceIsNot(name string) *RequestedObjectStage {
	s.Logger = s.Logger.WithValues("namespace", s.AdmissionRequest.Namespace)
	if s.AdmissionRequest.Namespace == name {
		s.Allow(errors.Errorf("namespace is %s", name))
	}
	return s
}

func (s *RequestedObjectStage) IsValid() *RequestedObjectStage {
	if s.Operation == admission_api.Delete {
		return s
	}
	if s.AdmissionRequest.Object.Raw == nil {
		s.Allow(errors.New("Request object raw is nil"))
	}

	return s
}

func (s *RequestedObjectStage) The() *RequestedObjectStage {
	return s
}

func (s *RequestedObjectStage) And() *RequestedObjectStage {
	return s
}

func (s *RequestedObjectStage) Or() *RequestedObjectStage {
	return s
}

func (s *RequestedObjectStage) End() *GivenStage {
	return s.GivenStage
}

/**
 * When stage
 */

func (g *GivenStage) When(hook Hook) *WhenStage {
	return hook.Review(g)
}

func (w *WhenStage) I() *WhenStage {
	if !w.CanContinue() {
		return w
	}
	return w
}

func (w *WhenStage) PatchTheRequest() *WhenStage {
	if !w.CanContinue() {
		return w
	}
	w.Patch, w.Error = w.Patcher.Create()
	if w.Error != nil {
		w.Fail(errors.WithMessage(w.Error, "Can't patch object"))
		return w
	}
	w.JsonPatch()
	return w
}

func (w *WhenStage) Then() *ThenStage {
	return &ThenStage{
		WhenStage: w,
	}
}

/**
 * Then stage
 */

func (t *ThenStage) I() *ThenStage {
	return t
}

func (t *ThenStage) Can() *ThenStage {
	return t
}

func (t *ThenStage) ReturnThePatch() *ThenStage {
	return t
}

func (t *ThenStage) OrElse() *ThenStage {
	return t
}

func (t *ThenStage) ReturnTheStatus() *ThenStage {
	return t
}

/**
 * End
 */
func (t *ThenStage) End() *EndStage {
	return &EndStage{t}
}

func (e *EndStage) Response() *admission_api.AdmissionResponse {
	switch e.ResponseType {
	case allowed:
		e.Logger.Info("allowing")
		return newAllowedResponse(e.AdmissionRequest, e.Error)
	case failure:
		e.Logger.Error(e.Error, "denying")
		return newFailureResponse(e.AdmissionRequest, e.Error)
	case jsonpatch:
		e.Logger.WithValues("patch", e.Patch).Info("patching")
		return newJSONPatchResponse(e.AdmissionRequest, e.Patch)
	}
	return newAllowedResponse(e.AdmissionRequest, errors.New("should never reach this code"))
}
