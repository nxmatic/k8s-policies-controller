package namespace

import (
	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"
	"github.com/pkg/errors"
	core_api "k8s.io/api/core/v1"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(name string, mgr manager.Manager, k8s *k8s.Interface) error {
	reconciler := &reconciler{
		k8s,
	}
	// Create a newReconcilierConfiguration controller
	ctrl, err := controller.New(name+"/namespace", mgr, controller.Options{Reconciler: reconciler})
	if err != nil {
		return errors.WithStack(err)
	}
	return add(ctrl, reconciler)
}

// add adds a newReconcilierConfiguration Controller to mgr with r as the reconcile.Reconciler.
func add(ctrl controller.Controller, r reconcile.Reconciler) error {

	// Watch for changes to primary resource GCPAuthPolicyProfile
	decorator := namespaceDecorator{handler: &handler.EnqueueRequestForObject{}}
	resource := &source.Kind{
		Type: &core_api.Namespace{
			TypeMeta: meta_api.TypeMeta{
				APIVersion: core_api.SchemeGroupVersion.String(),
				Kind:       "Namespace",
			},
		},
	}
	predicate, err := predicate.LabelSelectorPredicate(
		meta_api.LabelSelector{
			MatchLabels: map[string]string{
				"policies.nuxeo.io/watch": "true",
			}})
	if err != nil {
		return errors.WithStack(err)
	}
	err = ctrl.Watch(resource, &decorator, predicate)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
