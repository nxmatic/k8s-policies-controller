package gcpworkload

import (
	gcpiam_api "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/iam/v1beta1"
	gcpworkload_api "github.com/nuxeo/k8s-policy-controller/pkg/apis/gcpworkload/v1alpha1"

	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpworkload/k8s"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpworkload/reconciler"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpworkload/reviewer"
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi"
	namespace_spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/namespace"
	reviewer_spi "github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/reviewer"

	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	_name                   string                                            = "gcpworkload"
	_serviceaccountResource schema.GroupVersionResource                       = core_api.SchemeGroupVersion.WithResource("serviceaccounts")
	_serviceaccountHook     reviewer_spi.Hook                                 = &serviceaccountHook{}
	_plugin                 spi.Plugin                                        = &plugin{}
	_hooks                  map[schema.GroupVersionResource]reviewer_spi.Hook = map[schema.GroupVersionResource]reviewer_spi.Hook{
		_serviceaccountResource: _serviceaccountHook,
	}
)

type (
	plugin struct {
	}
	serviceaccountHook struct{}
)

func SupplyPlugin() spi.Plugin {
	return _plugin
}

func (p *plugin) Name() string {
	return _name
}

func (p *plugin) Add(manager manager.Manager) error {
	scheme := manager.GetScheme()
	if err := gcpworkload_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to load gcpworkloadprofile scheme")
	}
	if err := gcpiam_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to load gcpworkloadprofile scheme")
	}
	if err := core_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to load core scheme")
	}
	k8s, err := k8s.NewInterface(manager)
	if err != nil {
		return errors.Wrap(err, "cannot acquire k8s interface")
	}
	reconciler.Add(manager, k8s)
	namespace_spi.Add(_name, manager, &k8s.Interface)
	reviewer_spi.Add(_name, manager, &k8s.Interface, _hooks)
	return nil
}

func (h *serviceaccountHook) Review(s *reviewer_spi.GivenStage) *reviewer_spi.WhenStage {
	return reviewer.Given().
		The().RequestedObject(s).And().
		The().RequestedKind().IsAServiceAccount().End().
		The().RequestedProfile().Applies().End().
		End()
}
