package reconciler

import (
	"context"
	"time"

	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/gcpworkload/k8s"

	errors_api "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type (
	reconciler struct {
		*k8s.Interface
	}
)

var (
	end            reconcile.Result = reconcile.Result{}
	requeueOnError reconcile.Result = reconcile.Result{RequeueAfter: 5 * time.Minute}
)

func (r *reconciler) Reconcile(ctx context.Context, o reconcile.Request) (reconcile.Result, error) {
	profile, err := r.Interface.GetProfile(o.Name)
	if err != nil {
		if !errors_api.IsNotFound(err) {
			return requeueOnError, err
		}
		return end, nil
	}
	return r.updateHandler(profile)
}

func (r *reconciler) deleteHandler(name string) (reconcile.Result, error) {
	return end, nil
}

func (r *reconciler) updateHandler(profile *k8s.Profile) (reconcile.Result, error) {
	return end, nil
}
