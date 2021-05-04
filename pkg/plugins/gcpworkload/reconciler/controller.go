package reconciler

import (
	gcp_iam_api "github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/iam/v1beta1"
	gcpworkload_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpworkload/v1alpha1"
	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/gcpworkload/k8s"
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

func Add(mgr manager.Manager, k8s *k8s.Interface) error {
	reconciler := &reconciler{
		k8s,
	}
	return add(mgr, reconciler)
}

// add adds a newReconcilierConfiguration Controller to mgr with r as the reconcile.Reconciler.
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a newReconcilierConfiguration controller
	c, err := controller.New("gcpworkloadprofiles", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch for changes to primary resource gcpworkloadPolicyProfile
	decorator := profileDecorator{handler: &handler.EnqueueRequestForObject{}}
	profileResource := &source.Kind{
		Type: &gcpworkload_api.Profile{
			TypeMeta: meta_api.TypeMeta{
				APIVersion: gcpworkload_api.SchemeGroupVersion.String(),
				Kind:       gcpworkload_api.ProfileKind.String(),
			},
		},
	}
	err = c.Watch(profileResource, &decorator)
	if err != nil {
		return errors.WithStack(err)
	}

	predicate, err := predicate.LabelSelectorPredicate(
		meta_api.LabelSelector{
			MatchLabels: map[string]string{
				gcpworkload_api.WatchKey.String(): "true",
			}})
	if err != nil {
		return errors.WithStack(err)
	}

	// Watch for changes to patched gcp IAMPolicyMember

	iamPolicyMemberResource := &source.Kind{
		Type: &gcp_iam_api.IAMPolicyMember{
			TypeMeta: meta_api.TypeMeta{
				APIVersion: core_api.SchemeGroupVersion.String(),
				Kind:       "Secrets",
			},
		},
	}

	err = c.Watch(iamPolicyMemberResource, &enqueueRequestForOwner{}, predicate)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
