package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type MetricRollupSpec struct {
	Endpoint string `json:"endpoint"`
	Query    string `json:"query"`
	Interval string `json:"interval"`
}

type MetricRollupStatus struct {
	StorageReady bool   `json:"storageReady,omitempty"`
	CronJobName  string `json:"cronJobName,omitempty"`
}

type MetricRollup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetricRollupSpec   `json:"spec,omitempty"`
	Status MetricRollupStatus `json:"status,omitempty"`
}

type MetricRollupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MetricRollup `json:"items"`
}
