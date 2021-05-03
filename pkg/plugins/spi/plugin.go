package spi

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type (
	Supplier func() Plugin

	Plugin interface {
		Name() string
		Add(manager manager.Manager) error
	}
)
