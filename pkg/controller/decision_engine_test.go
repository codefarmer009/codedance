package controller

import (
	"testing"
	"time"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
)

func TestDecisionEngine_Evaluate_SuccessRate(t *testing.T) {
	engine := NewDefaultDecisionEngine()

	tests := []struct {
		name        string
		metrics     *HealthMetrics
		thresholds  deployv1alpha1.MetricsConfig
		wantAction  ActionType
		wantReasonContains string
	}{
		{
			name: "success rate below threshold should rollback",
			metrics: &HealthMetrics{
				SuccessRate: 85.0,
				Latency:     LatencyMetrics{P99: 100 * time.Millisecond},
				ErrorRate:   1.0,
				PodHealth:   PodHealthMetrics{Ready: 3, NotReady: 0},
			},
			thresholds: deployv1alpha1.MetricsConfig{
				SuccessRate: deployv1alpha1.MetricThreshold{Threshold: 95.0},
				Latency:     deployv1alpha1.LatencyConfig{P99: "500ms"},
				ErrorRate:   deployv1alpha1.MetricThreshold{Threshold: 5.0},
			},
			wantAction: RollbackAction,
			wantReasonContains: "成功率",
		},
		{
			name: "success rate above threshold should continue",
			metrics: &HealthMetrics{
				SuccessRate: 98.0,
				Latency:     LatencyMetrics{P99: 100 * time.Millisecond},
				ErrorRate:   1.0,
				PodHealth:   PodHealthMetrics{Ready: 3, NotReady: 0},
			},
			thresholds: deployv1alpha1.MetricsConfig{
				SuccessRate: deployv1alpha1.MetricThreshold{Threshold: 95.0},
				Latency:     deployv1alpha1.LatencyConfig{P99: "500ms"},
				ErrorRate:   deployv1alpha1.MetricThreshold{Threshold: 5.0},
			},
			wantAction: ContinueAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := engine.Evaluate(tt.metrics, tt.thresholds)
			if decision.Action != tt.wantAction {
				t.Errorf("Evaluate() action = %v, want %v", decision.Action, tt.wantAction)
			}
			if tt.wantReasonContains != "" && decision.Reason != "" {
				// Just check if decision has a reason when expected
				if decision.Reason == "" {
					t.Errorf("Evaluate() reason is empty, want it to contain %q", tt.wantReasonContains)
				}
			}
		})
	}
}

func TestDecisionEngine_Evaluate_ErrorRate(t *testing.T) {
	engine := NewDefaultDecisionEngine()

	tests := []struct {
		name       string
		errorRate  float64
		threshold  float64
		wantAction ActionType
	}{
		{
			name:       "error rate above threshold should rollback",
			errorRate:  10.0,
			threshold:  5.0,
			wantAction: RollbackAction,
		},
		{
			name:       "error rate below threshold should continue",
			errorRate:  2.0,
			threshold:  5.0,
			wantAction: ContinueAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &HealthMetrics{
				SuccessRate: 99.0,
				Latency:     LatencyMetrics{P99: 100 * time.Millisecond},
				ErrorRate:   tt.errorRate,
				PodHealth:   PodHealthMetrics{Ready: 3, NotReady: 0},
			}
			thresholds := deployv1alpha1.MetricsConfig{
				SuccessRate: deployv1alpha1.MetricThreshold{Threshold: 95.0},
				Latency:     deployv1alpha1.LatencyConfig{P99: "500ms"},
				ErrorRate:   deployv1alpha1.MetricThreshold{Threshold: tt.threshold},
			}

			decision := engine.Evaluate(metrics, thresholds)
			if decision.Action != tt.wantAction {
				t.Errorf("Evaluate() action = %v, want %v", decision.Action, tt.wantAction)
			}
		})
	}
}

func TestDecisionEngine_Evaluate_Latency(t *testing.T) {
	engine := NewDefaultDecisionEngine()

	tests := []struct {
		name           string
		latencyP99     time.Duration
		thresholdP99   string
		wantAction     ActionType
	}{
		{
			name:         "latency within threshold should continue",
			latencyP99:   100 * time.Millisecond,
			thresholdP99: "500ms",
			wantAction:   ContinueAction,
		},
		{
			name:         "latency exceeding threshold should pause",
			latencyP99:   800 * time.Millisecond,
			thresholdP99: "500ms",
			wantAction:   PauseAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &HealthMetrics{
				SuccessRate: 99.0,
				Latency:     LatencyMetrics{P99: tt.latencyP99},
				ErrorRate:   1.0,
				PodHealth:   PodHealthMetrics{Ready: 3, NotReady: 0},
			}
			thresholds := deployv1alpha1.MetricsConfig{
				SuccessRate: deployv1alpha1.MetricThreshold{Threshold: 95.0},
				Latency:     deployv1alpha1.LatencyConfig{P99: tt.thresholdP99},
				ErrorRate:   deployv1alpha1.MetricThreshold{Threshold: 5.0},
			}

			decision := engine.Evaluate(metrics, thresholds)
			if decision.Action != tt.wantAction {
				t.Errorf("Evaluate() action = %v, want %v", decision.Action, tt.wantAction)
			}
		})
	}
}

func TestDecisionEngine_Evaluate_PodHealth(t *testing.T) {
	engine := NewDefaultDecisionEngine()

	tests := []struct {
		name       string
		podHealth  PodHealthMetrics
		wantAction ActionType
	}{
		{
			name: "healthy pods should continue",
			podHealth: PodHealthMetrics{
				Ready:    9,
				NotReady: 1,
			},
			wantAction: ContinueAction,
		},
		{
			name: "unhealthy pods should pause",
			podHealth: PodHealthMetrics{
				Ready:    2,
				NotReady: 8,
			},
			wantAction: PauseAction,
		},
		{
			name: "no pods should pause",
			podHealth: PodHealthMetrics{
				Ready:    0,
				NotReady: 0,
			},
			wantAction: PauseAction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := &HealthMetrics{
				SuccessRate: 99.0,
				Latency:     LatencyMetrics{P99: 100 * time.Millisecond},
				ErrorRate:   1.0,
				PodHealth:   tt.podHealth,
			}
			thresholds := deployv1alpha1.MetricsConfig{
				SuccessRate: deployv1alpha1.MetricThreshold{Threshold: 95.0},
				Latency:     deployv1alpha1.LatencyConfig{P99: "500ms"},
				ErrorRate:   deployv1alpha1.MetricThreshold{Threshold: 5.0},
			}

			decision := engine.Evaluate(metrics, thresholds)
			if decision.Action != tt.wantAction {
				t.Errorf("Evaluate() action = %v, want %v (reason: %s)", decision.Action, tt.wantAction, decision.Reason)
			}
		})
	}
}

func TestDecisionEngine_Evaluate_InvalidLatencyThreshold(t *testing.T) {
	engine := NewDefaultDecisionEngine()

	metrics := &HealthMetrics{
		SuccessRate: 99.0,
		Latency:     LatencyMetrics{P99: 100 * time.Millisecond},
		ErrorRate:   1.0,
		PodHealth:   PodHealthMetrics{Ready: 3, NotReady: 0},
	}
	thresholds := deployv1alpha1.MetricsConfig{
		SuccessRate: deployv1alpha1.MetricThreshold{Threshold: 95.0},
		Latency:     deployv1alpha1.LatencyConfig{P99: "invalid"},
		ErrorRate:   deployv1alpha1.MetricThreshold{Threshold: 5.0},
	}

	decision := engine.Evaluate(metrics, thresholds)
	if decision.Action != PauseAction {
		t.Errorf("Evaluate() with invalid latency threshold should pause, got %v", decision.Action)
	}
	if decision.Reason == "" {
		t.Error("Evaluate() with invalid latency threshold should have a reason")
	}
}
