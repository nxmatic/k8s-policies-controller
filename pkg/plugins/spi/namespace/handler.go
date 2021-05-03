package namespace

import (
	"fmt"
	"reflect"

	core_api "k8s.io/api/core/v1"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type namespaceDecorator struct {
	handler handler.EventHandler
}

func (e *namespaceDecorator) Create(evt event.CreateEvent, q workqueue.RateLimitingInterface) {
	e.handler.Create(evt, q)
}

func (e *namespaceDecorator) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	if !reflect.DeepEqual(evt.ObjectOld.(*core_api.Namespace).ObjectMeta.Annotations, evt.ObjectNew.(*core_api.Namespace).ObjectMeta.Annotations) {
		log.Log.WithValues("", evt.ObjectNew.GetName()).Info(
			fmt.Sprintf("%T/%s has been updated", evt.ObjectNew, evt.ObjectNew.GetName()))
	}
	e.handler.Update(evt, q)
}

func (e *namespaceDecorator) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	e.handler.Delete(evt, q)
}

func (e *namespaceDecorator) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	e.handler.Generic(evt, q)
}
