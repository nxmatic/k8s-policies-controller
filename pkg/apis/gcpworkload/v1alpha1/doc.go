// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:rbac:groups=gcpworkload.profile.nuuxeo.io,resources=profiles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gcpworkload.profile.nuuxeo.io,resources=profiles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="iam.cnrm.cloud.google.com",resources=iampolicymembers,verbs=*
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:webhook:versions={v1,v1beta1},groups=gcpworkload.profile.nuuxeo.io,resources=serviceaccounts,verbs="CREATE",name=gcpworkload.policy,path=/mutate-v1-serviceaccounts,mutating=true,failurePolicy=Ignore
package v1alpha1
