package v1

import "k8s.io/apimachinery/pkg/runtime"

func (in *MetricRollup) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

func (in *MetricRollup) DeepCopy() *MetricRollup {
	if in == nil {
		return nil
	}
	out := new(MetricRollup)
	in.DeepCopyInto(out)
	return out
}

func (in *MetricRollup) DeepCopyInto(out *MetricRollup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

func (in *MetricRollupList) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}

func (in *MetricRollupList) DeepCopy() *MetricRollupList {
	if in == nil {
		return nil
	}
	out := new(MetricRollupList)
	in.DeepCopyInto(out)
	return out
}

func (in *MetricRollupList) DeepCopyInto(out *MetricRollupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]MetricRollup, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}
