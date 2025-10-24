package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CanaryDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CanaryDeploymentSpec   `json:"spec"`
	Status CanaryDeploymentStatus `json:"status,omitempty"`
}

type CanaryDeploymentSpec struct {
	TargetDeployment string          `json:"targetDeployment"`
	CanaryVersion    string          `json:"canaryVersion"`
	Strategy         DeployStrategy  `json:"strategy"`
	Metrics          MetricsConfig   `json:"metrics"`
	AutoRollback     AutoRollbackConfig `json:"autoRollback"`
}

type DeployStrategy struct {
	Type  string         `json:"type"`
	Steps []DeployStep   `json:"steps"`
}

type DeployStep struct {
	Weight  int    `json:"weight"`
	Pause   string `json:"pause"`
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
	Phase          string      `json:"phase"`
	CurrentStep    int         `json:"currentStep"`
	CurrentWeight  int         `json:"currentWeight"`
	Reason         string      `json:"reason,omitempty"`
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
}

type CanaryDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CanaryDeployment `json:"items"`
}
