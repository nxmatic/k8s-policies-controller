package node

import (
	node_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/node/v1alpha1"
	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/node/k8s"

	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/node/reviewer"
	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi"
	namespace_spi "github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/namespace"
	reviewer_spi "github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/reviewer"

	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	_name        string                                            = "node"
	_podResource schema.GroupVersionResource                       = node_api.PodsResource
	_podHook     reviewer_spi.Hook                                 = &podHook{}
	_plugin      spi.Plugin                                        = &plugin{}
	_hooks       map[schema.GroupVersionResource]reviewer_spi.Hook = map[schema.GroupVersionResource]reviewer_spi.Hook{
		_podResource: _podHook,
	}
)

func SupplyPlugin() spi.Plugin {
	return _plugin
}

type (
	plugin struct {
	}
	podHook struct {
	}
)

func (p *plugin) Name() string {
	return _name
}

func (p *plugin) Add(mgr manager.Manager) error {
	scheme := mgr.GetScheme()
	if err := node_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to setup scheme")
	}
	if err := core_api.SchemeBuilder.AddToScheme(scheme); err != nil {
		return errors.Wrap(err, "failed to load core scheme")
	}
	if k8s, err := k8s.NewInterface(mgr); err != nil {
		return errors.Wrap(err, "failed to acquire k8s interface")
	} else {
		namespace_spi.Add(_name, mgr, &k8s.Interface)
		reviewer_spi.Add(_name, mgr, &k8s.Interface, _hooks)
	}
	return nil
}

func (h *podHook) Review(s *reviewer_spi.GivenStage) *reviewer_spi.WhenStage {
	return reviewer.Given().
		The().RequestedObject(s).And().
		The().RequestedKind().IsAPod().End().And().
		The().RequestedProfile().Applies().End().
		End()
}
