package strategy

import (
	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
)

type LinearStrategy struct{}

func NewLinearStrategy() *LinearStrategy {
	return &LinearStrategy{}
}

func (s *LinearStrategy) GenerateSteps() []deployv1alpha1.DeployStep {
	return []deployv1alpha1.DeployStep{
		{Weight: 5, Pause: "5m"},
		{Weight: 10, Pause: "5m"},
		{Weight: 25, Pause: "10m"},
		{Weight: 50, Pause: "10m"},
		{Weight: 75, Pause: "10m"},
		{Weight: 100, Pause: "0"},
	}
}
