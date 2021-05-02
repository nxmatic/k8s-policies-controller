package namespace

import (
	"context"
	"time"

	"github.com/nuxeo/k8s-policy-controller/pkg/plugins/spi/k8s"

	core_api "k8s.io/api/core/v1"
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
	requeueOnError reconcile.Result = reconcile.Result{RequeueAfter: 10 * time.Second}
)

func (r *reconciler) Reconcile(ctx context.Context, o reconcile.Request) (reconcile.Result, error) {
	namespace, err := r.Interface.GetNamespace(o.Name)
	if err != nil {
		if !errors_api.IsNotFound(err) {
			return requeueOnError, err
		}
		return end, nil
	}
	return r.updateHandler(namespace)
}

func (r *reconciler) deleteHandler(name string) (reconcile.Result, error) {
	r.Interface.DefaultMeta = nil
	return end, nil
}

func (r *reconciler) updateHandler(namespace *core_api.Namespace) (reconcile.Result, error) {
	r.Interface.DefaultMeta = &namespace.ObjectMeta
	return end, nil
}
