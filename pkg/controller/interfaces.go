package controller

import (
	"context"
	"time"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
)

type TrafficManager interface {
	UpdateWeight(ctx context.Context, canary *deployv1alpha1.CanaryDeployment, weight int) error
	CreateCanaryRoute(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error
}

type MetricsAnalyzer interface {
	Collect(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) (*HealthMetrics, error)
}

type DecisionEngine interface {
	Evaluate(metrics *HealthMetrics, thresholds deployv1alpha1.MetricsConfig) Decision
}

type RollbackManager interface {
	Rollback(ctx context.Context, canary *deployv1alpha1.CanaryDeployment, reason string) error
}

type HealthMetrics struct {
	SuccessRate float64
	Latency     LatencyMetrics
	ErrorRate   float64
	PodHealth   PodHealthMetrics
	Resources   ResourceMetrics
}

type LatencyMetrics struct {
	P50 time.Duration
	P90 time.Duration
	P99 time.Duration
}

type PodHealthMetrics struct {
	Ready    int
	NotReady int
	Failed   int
}

type ResourceMetrics struct {
	CPUUsage    float64
	MemoryUsage float64
}
