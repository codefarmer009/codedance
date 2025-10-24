package metrics

import (
	"context"
	"fmt"
	"time"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
	"github.com/codefarmer009/codedance/pkg/controller"
	promapi "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PrometheusAnalyzer struct {
	promClient v1.API
}

func NewPrometheusAnalyzer(promURL string) (*PrometheusAnalyzer, error) {
	client, err := promapi.NewClient(promapi.Config{
		Address: promURL,
	})
	if err != nil {
		return nil, err
	}

	return &PrometheusAnalyzer{
		promClient: v1.NewAPI(client),
	}, nil
}

func (m *PrometheusAnalyzer) Collect(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) (*controller.HealthMetrics, error) {
	metrics := &controller.HealthMetrics{}

	successRate, err := m.querySuccessRate(ctx, canary)
	if err != nil {
		return nil, err
	}
	metrics.SuccessRate = successRate

	latency, err := m.queryLatency(ctx, canary)
	if err != nil {
		return nil, err
	}
	metrics.Latency = latency

	errorRate, err := m.queryErrorRate(ctx, canary)
	if err != nil {
		return nil, err
	}
	metrics.ErrorRate = errorRate

	podHealth, err := m.queryPodHealth(ctx, canary)
	if err != nil {
		return nil, err
	}
	metrics.PodHealth = podHealth

	return metrics, nil
}

func (m *PrometheusAnalyzer) querySuccessRate(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) (float64, error) {
	query := canary.Spec.Metrics.SuccessRate.Query
	if query == "" {
		query = fmt.Sprintf(`
			sum(rate(http_requests_total{
				app="%s",
				version="%s",
				status=~"2.."
			}[5m])) /
			sum(rate(http_requests_total{
				app="%s",
				version="%s"
			}[5m])) * 100
		`, canary.Name, canary.Spec.CanaryVersion,
			canary.Name, canary.Spec.CanaryVersion)
	}

	result, _, err := m.promClient.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}

	return parseFloatFromResult(result), nil
}

func (m *PrometheusAnalyzer) queryLatency(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) (controller.LatencyMetrics, error) {
	latency := controller.LatencyMetrics{}

	query := canary.Spec.Metrics.Latency.Query
	if query == "" {
		query = fmt.Sprintf(`
			histogram_quantile(0.99,
				sum(rate(http_request_duration_seconds_bucket{app="%s"}[5m])) by (le))
		`, canary.Name)
	}

	result, _, err := m.promClient.Query(ctx, query, time.Now())
	if err != nil {
		return latency, err
	}

	p99Value := parseFloatFromResult(result)
	latency.P99 = time.Duration(p99Value * float64(time.Second))

	return latency, nil
}

func (m *PrometheusAnalyzer) queryErrorRate(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) (float64, error) {
	query := canary.Spec.Metrics.ErrorRate.Query
	if query == "" {
		query = fmt.Sprintf(`
			sum(rate(http_requests_total{
				app="%s",
				version="%s",
				status=~"5.."
			}[5m])) /
			sum(rate(http_requests_total{
				app="%s",
				version="%s"
			}[5m])) * 100
		`, canary.Name, canary.Spec.CanaryVersion,
			canary.Name, canary.Spec.CanaryVersion)
	}

	result, _, err := m.promClient.Query(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}

	return parseFloatFromResult(result), nil
}

func (m *PrometheusAnalyzer) queryPodHealth(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) (controller.PodHealthMetrics, error) {
	return controller.PodHealthMetrics{
		Ready:    3,
		NotReady: 0,
		Failed:   0,
	}, nil
}

func parseFloatFromResult(result model.Value) float64 {
	if result == nil {
		return 0
	}

	switch v := result.(type) {
	case model.Vector:
		if len(v) > 0 {
			return float64(v[0].Value)
		}
	case *model.Scalar:
		return float64(v.Value)
	}

	return 0
}
