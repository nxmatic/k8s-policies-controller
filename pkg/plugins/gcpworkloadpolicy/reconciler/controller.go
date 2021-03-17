package reconciler

import (
	gcpworkloadpolicy_api "github.com/nuxeo/k8s-policy-controller/apis/gcpworkloadpolicyprofile/v1alpha1"

	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/gcpworkloadpolicy/k8s"

	"github.com/pkg/errors"
	meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager, client dynamic.Interface) error {
	reconciler := &reconciler{
		k8s.NewInterface(client),
	}
	return add(mgr, reconciler)
}

// add adds a newReconcilierConfiguration Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a newReconcilierConfiguration controller
	c, err := controller.New("gcpworkloadpolicyprofiles", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch for changes to primary resource gcpworkloadPolicyProfile
	decorator := profileDecorator{handler: &handler.EnqueueRequestForObject{}}
	profileResource := &source.Kind{
		Type: &gcpworkloadpolicy_api.Profile{
			TypeMeta: meta_api.TypeMeta{
				APIVersion: gcpworkloadpolicy_api.SchemeGroupVersion.String(),
				Kind:       gcpworkloadpolicy_api.ProfileKind.String(),
			},
		},
	}
	err = c.Watch(profileResource, &decorator)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
