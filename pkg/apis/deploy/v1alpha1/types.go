package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type CanaryDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CanaryDeploymentSpec   `json:"spec"`
	Status CanaryDeploymentStatus `json:"status,omitempty"`
}

type CanaryDeploymentSpec struct {
	TargetDeployment string             `json:"targetDeployment"`
	CanaryVersion    string             `json:"canaryVersion"`
	Strategy         DeployStrategy     `json:"strategy"`
	Metrics          MetricsConfig      `json:"metrics"`
	AutoRollback     AutoRollbackConfig `json:"autoRollback"`
}

type DeployStrategy struct {
	Type  string       `json:"type"`
	Steps []DeployStep `json:"steps"`
}

type DeployStep struct {
	Weight  int           `json:"weight"`
	Pause   string        `json:"pause"`
	Metrics []MetricCheck `json:"metrics,omitempty"`
}

type MetricCheck struct {
	Name      string  `json:"name"`
	Threshold float64 `json:"threshold"`
}

type MetricsConfig struct {
	SuccessRate MetricThreshold `json:"successRate"`
	Latency     LatencyConfig   `json:"latency"`
	ErrorRate   MetricThreshold `json:"errorRate"`
}

type MetricThreshold struct {
	Threshold float64 `json:"threshold"`
	Query     string  `json:"query"`
}

type LatencyConfig struct {
	P99   string `json:"p99"`
	Query string `json:"query"`
}

type AutoRollbackConfig struct {
	Enabled       bool `json:"enabled"`
	OnMetricsFail bool `json:"onMetricsFail"`
	OnPodCrash    bool `json:"onPodCrash"`
}

type CanaryDeploymentStatus struct {
	Phase          string             `json:"phase"`
	CurrentStep    int                `json:"currentStep"`
	CurrentWeight  int                `json:"currentWeight"`
	Reason         string             `json:"reason,omitempty"`
	LastUpdateTime metav1.Time        `json:"lastUpdateTime,omitempty"`
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
}

type CanaryDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CanaryDeployment `json:"items"`
}

func (in *CanaryDeployment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *CanaryDeployment) DeepCopy() *CanaryDeployment {
	if in == nil {
		return nil
	}
	out := new(CanaryDeployment)
	in.DeepCopyInto(out)
	return out
}

func (in *CanaryDeployment) DeepCopyInto(out *CanaryDeployment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *CanaryDeploymentSpec) DeepCopyInto(out *CanaryDeploymentSpec) {
	*out = *in
	in.Strategy.DeepCopyInto(&out.Strategy)
	out.Metrics = in.Metrics
	out.AutoRollback = in.AutoRollback
}

func (in *DeployStrategy) DeepCopyInto(out *DeployStrategy) {
	*out = *in
	if in.Steps != nil {
		in, out := &in.Steps, &out.Steps
		*out = make([]DeployStep, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *DeployStep) DeepCopyInto(out *DeployStep) {
	*out = *in
	if in.Metrics != nil {
		in, out := &in.Metrics, &out.Metrics
		*out = make([]MetricCheck, len(*in))
		copy(*out, *in)
	}
}

func (in *CanaryDeploymentStatus) DeepCopyInto(out *CanaryDeploymentStatus) {
	*out = *in
	in.LastUpdateTime.DeepCopyInto(&out.LastUpdateTime)
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *CanaryDeploymentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *CanaryDeploymentList) DeepCopy() *CanaryDeploymentList {
	if in == nil {
		return nil
	}
	out := new(CanaryDeploymentList)
	in.DeepCopyInto(out)
	return out
}

func (in *CanaryDeploymentList) DeepCopyInto(out *CanaryDeploymentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CanaryDeployment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
