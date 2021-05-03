package k8s

import (
	"strings"

	policy_meta_api "github.com/nuxeo/k8s-policy-controller/pkg/apis/meta/v1alpha1"
	k8s_meta_api "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	ProfileGetter interface {
		GetName() string
		GetSelector() policy_meta_api.ObjectSelector
	}

	ProfileSupplier interface {
		Get(name string) (ProfileGetter, error)
		Key() string
	}

	ProfileCollector struct {
		accumulator accumulator_t
		supplier    ProfileSupplier
	}

	void          struct{}
	accumulator_t map[ProfileGetter]void
)

func NewProfileCollector(supplier ProfileSupplier) ProfileCollector {
	return ProfileCollector{
		accumulator: make(accumulator_t),
		supplier:    supplier,
	}
}

func (collector ProfileCollector) addNames(meta k8s_meta_api.ObjectMeta) {
	empty := void{}
	if value, ok := meta.Annotations[collector.supplier.Key()]; ok {
		for _, name := range strings.Split(value, ",") {
			name = strings.TrimSpace(name)
			if profile, err := collector.supplier.Get(name); err != nil {
				collector.accumulator[profile] = empty
			}
		}
	}
}

func (collector ProfileCollector) Profiles() []ProfileGetter {
	profiles := make([]ProfileGetter, 0, len(collector.accumulator))
	for profile, _ := range collector.accumulator {
		n := len(profiles)
		profiles = profiles[0 : n+1]
		profiles[n] = profile
	}
	return profiles
}
