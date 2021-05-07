package namespace

import (
	meta_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/meta/v1alpha1"
	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/spi/k8s"

	"github.com/pkg/errors"
	k8s_core_api "k8s.io/api/core/v1"
	k8s_meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	// Watch for changes to primary resource Namespace
	decorator := namespaceDecorator{handler: &handler.EnqueueRequestForObject{}}
	resource := &source.Kind{
		Type: &k8s_core_api.Namespace{
			TypeMeta: k8s_meta_api.TypeMeta{
				APIVersion: k8s_core_api.SchemeGroupVersion.String(),
				Kind:       "Namespace",
			},
		},
	}
	predicate, err := predicate.LabelSelectorPredicate(
		k8s_meta_api.LabelSelector{
			MatchLabels: map[string]string{
				meta_api.WatchKey.String(): "true",
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
