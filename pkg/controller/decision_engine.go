package controller

import (
	"fmt"
	"time"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
)

type DefaultDecisionEngine struct{}

func NewDefaultDecisionEngine() *DefaultDecisionEngine {
	return &DefaultDecisionEngine{}
}

func (d *DefaultDecisionEngine) Evaluate(metrics *HealthMetrics, thresholds deployv1alpha1.MetricsConfig) Decision {
	score := 0

	if metrics.SuccessRate >= thresholds.SuccessRate.Threshold {
		score += 40
	} else {
		return Decision{
			Action: RollbackAction,
			Reason: fmt.Sprintf("成功率 %.2f%% 低于阈值 %.2f%%",
				metrics.SuccessRate, thresholds.SuccessRate.Threshold),
		}
	}

	maxLatencyP99, _ := time.ParseDuration(thresholds.Latency.P99)
	if metrics.Latency.P99 <= maxLatencyP99 {
		score += 30
	} else if metrics.Latency.P99 <= time.Duration(float64(maxLatencyP99)*1.2) {
		score += 15
	} else {
		return Decision{
			Action: PauseAction,
			Reason: fmt.Sprintf("P99延迟 %v 超过阈值 %v",
				metrics.Latency.P99, maxLatencyP99),
		}
	}

	if metrics.ErrorRate <= thresholds.ErrorRate.Threshold {
		score += 30
	} else {
		return Decision{
			Action: RollbackAction,
			Reason: fmt.Sprintf("错误率 %.2f%% 超过阈值 %.2f%%",
				metrics.ErrorRate, thresholds.ErrorRate.Threshold),
		}
	}

	totalPods := metrics.PodHealth.Ready + metrics.PodHealth.NotReady
	if totalPods == 0 {
		return Decision{
			Action: PauseAction,
			Reason: "没有可用的 Pod",
		}
	}

	podHealthRate := float64(metrics.PodHealth.Ready) / float64(totalPods)
	if podHealthRate < 0.8 {
		return Decision{
			Action: PauseAction,
			Reason: fmt.Sprintf("Pod 健康率 %.2f%% 过低", podHealthRate*100),
		}
	}

	if score >= 90 {
		return Decision{Action: ContinueAction, Score: score}
	} else if score >= 70 {
		return Decision{
			Action: PauseAction,
			Score:  score,
			Reason: "指标轻微异常，建议暂停观察",
		}
	} else {
		return Decision{
			Action: RollbackAction,
			Score:  score,
			Reason: "多项指标异常，建议回滚",
		}
	}
}
