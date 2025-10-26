package metrics

import (
	"testing"

	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
)

func TestParseFloatFromResult_Vector(t *testing.T) {
	tests := []struct {
		name   string
		result model.Value
		want   float64
	}{
		{
			name:   "nil result returns 0",
			result: nil,
			want:   0,
		},
		{
			name: "vector with value",
			result: model.Vector{
				&model.Sample{Value: model.SampleValue(42.5)},
			},
			want: 42.5,
		},
		{
			name:   "empty vector returns 0",
			result: model.Vector{},
			want:   0,
		},
		{
			name: "scalar value",
			result: &model.Scalar{
				Value: model.SampleValue(99.9),
			},
			want: 99.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFloatFromResult(tt.result)
			if got != tt.want {
				t.Errorf("parseFloatFromResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPodReady(t *testing.T) {
	tests := []struct {
		name string
		pod  *corev1.Pod
		want bool
	}{
		{
			name: "pod with ready condition true",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "pod with ready condition false",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
						},
					},
				},
			},
			want: false,
		},
		{
			name: "pod without ready condition",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{},
				},
			},
			want: false,
		},
		{
			name: "pod with other conditions but not ready",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodScheduled,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPodReady(tt.pod)
			if got != tt.want {
				t.Errorf("isPodReady() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPrometheusAnalyzer(t *testing.T) {
	// Test with invalid URL should return error
	_, err := NewPrometheusAnalyzer("http://[::1]:namedport")
	if err == nil {
		t.Error("NewPrometheusAnalyzer() with invalid URL should return error")
	}

	// Test with valid URL
	analyzer, err := NewPrometheusAnalyzer("http://localhost:9090")
	if err != nil {
		t.Errorf("NewPrometheusAnalyzer() error = %v, want nil", err)
	}
	if analyzer == nil {
		t.Error("NewPrometheusAnalyzer() returned nil analyzer")
	}
	if analyzer.promClient == nil {
		t.Error("NewPrometheusAnalyzer() promClient is nil")
	}
}

func TestNewPrometheusAnalyzerWithClientset(t *testing.T) {
	// Test with valid URL
	analyzer, err := NewPrometheusAnalyzerWithClientset("http://localhost:9090", nil)
	if err != nil {
		t.Errorf("NewPrometheusAnalyzerWithClientset() error = %v, want nil", err)
	}
	if analyzer == nil {
		t.Error("NewPrometheusAnalyzerWithClientset() returned nil analyzer")
	}
	if analyzer.promClient == nil {
		t.Error("NewPrometheusAnalyzerWithClientset() promClient is nil")
	}
}
