package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var (
	GroupVersion  = schema.GroupVersion{Group: "promruo.io", Version: "v1"}
	SchemeBuilder = runtime.NewSchemeBuilder(addToScheme)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addToScheme(s *runtime.Scheme) error {
	s.AddKnownTypes(GroupVersion,
		&MetricRollup{},
		&MetricRollupList{},
	)
	utilruntime.Must(s.SetVersionPriority(GroupVersion))
	return nil
}
