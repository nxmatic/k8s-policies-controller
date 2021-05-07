package reconciler

import (
	"context"
	"encoding/json"

	"github.com/go-logr/logr"
	gcpworkload_api "github.com/nuxeo/k8s-policies-controller/pkg/apis/gcpworkload/v1alpha1"
	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/gcpworkload/k8s"
	"github.com/nuxeo/k8s-policies-controller/pkg/plugins/gcpworkload/reviewer"
	k8s_core_api "k8s.io/api/core/v1"
)

type synchronizer struct {
	*k8s.Interface
	logr.Logger
}

func newSynchronizer(k8s *k8s.Interface) *synchronizer {
	return &synchronizer{
		Interface: k8s,
		Logger:    k8s.GetLogger().WithName("synchronizer"),
	}
}

func (synch *synchronizer) Start(ctx context.Context) error {
	synch.Logger.Info("starting")
	go synch.synchronize(ctx)
	return nil
}

func (synch *synchronizer) synchronize(ctx context.Context) {
	patch := func(sa *k8s_core_api.ServiceAccount, profile *gcpworkload_api.Profile) ([]byte, error) {
		patcher := reviewer.NewServiceAccountPatcher(sa, profile)
		patch, err := patcher.Create()
		if err != nil {
			return nil, err
		}
		bytes, err := json.Marshal(patch)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	}
	synch.SynchronizeServiceAccounts(patch, synch.Logger)
}
