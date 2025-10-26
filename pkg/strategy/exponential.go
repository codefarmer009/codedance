package strategy

import (
	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
)

type ExponentialStrategy struct{}

func NewExponentialStrategy() *ExponentialStrategy {
	return &ExponentialStrategy{}
}

func (s *ExponentialStrategy) GenerateSteps() []deployv1alpha1.DeployStep {
	return []deployv1alpha1.DeployStep{
		{Weight: 1, Pause: "5m"},
		{Weight: 5, Pause: "5m"},
		{Weight: 10, Pause: "10m"},
		{Weight: 25, Pause: "10m"},
		{Weight: 50, Pause: "15m"},
		{Weight: 100, Pause: "0"},
	}
}
