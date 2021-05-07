// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:rbac:groups=node.policies.nuxeo.io,resources=profiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=node.policies.nuxeo.io,resources=profiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:webhook:name=node.policy,versions={v1,v1beta1},groups=node.policies.nuxeo.io,resources=pods,verbs="CREATE",path=/mutate-v1-pods,mutating=true,failurePolicy=Ignore
package v1alpha1
