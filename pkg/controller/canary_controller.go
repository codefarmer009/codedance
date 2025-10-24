package controller

import (
	"context"
	"fmt"
	"time"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ActionType string

const (
	ContinueAction  ActionType = "continue"
	PauseAction     ActionType = "pause"
	RollbackAction  ActionType = "rollback"
)

type Decision struct {
	Action ActionType
	Reason string
	Score  int
}

type CanaryController struct {
	clientset       *kubernetes.Clientset
	trafficManager  TrafficManager
	metricsAnalyzer MetricsAnalyzer
	decisionEngine  DecisionEngine
	rollbackManager RollbackManager
}

func NewCanaryController(
	clientset *kubernetes.Clientset,
	trafficManager TrafficManager,
	metricsAnalyzer MetricsAnalyzer,
	decisionEngine DecisionEngine,
	rollbackManager RollbackManager,
) *CanaryController {
	return &CanaryController{
		clientset:       clientset,
		trafficManager:  trafficManager,
		metricsAnalyzer: metricsAnalyzer,
		decisionEngine:  decisionEngine,
		rollbackManager: rollbackManager,
	}
}

func (c *CanaryController) Run(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := c.reconcile(ctx); err != nil {
				fmt.Printf("reconcile failed: %v\n", err)
			}
		}
	}
}

func (c *CanaryController) reconcile(ctx context.Context) error {
	canaries, err := c.listCanaryDeployments(ctx)
	if err != nil {
		return err
	}

	for _, canary := range canaries {
		if err := c.processCanary(ctx, canary); err != nil {
			fmt.Printf("process canary %s failed: %v\n", canary.Name, err)
			continue
		}
	}

	return nil
}

func (c *CanaryController) processCanary(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	metrics, err := c.metricsAnalyzer.Collect(ctx, canary)
	if err != nil {
		return fmt.Errorf("collect metrics: %w", err)
	}

	decision := c.decisionEngine.Evaluate(metrics, canary.Spec.Metrics)

	switch decision.Action {
	case ContinueAction:
		return c.progressToNextStep(ctx, canary)
	case PauseAction:
		return c.pauseDeployment(ctx, canary, decision.Reason)
	case RollbackAction:
		return c.rollbackManager.Rollback(ctx, canary, decision.Reason)
	default:
		return fmt.Errorf("unknown action: %v", decision.Action)
	}
}

func (c *CanaryController) progressToNextStep(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	currentStep := canary.Status.CurrentStep
	totalSteps := len(canary.Spec.Strategy.Steps)

	if currentStep >= totalSteps-1 {
		return c.finalizeDeployment(ctx, canary)
	}

	nextStep := currentStep + 1
	weight := canary.Spec.Strategy.Steps[nextStep].Weight

	if err := c.trafficManager.UpdateWeight(ctx, canary, weight); err != nil {
		return fmt.Errorf("update traffic weight: %w", err)
	}

	canary.Status.CurrentStep = nextStep
	canary.Status.CurrentWeight = weight
	canary.Status.LastUpdateTime = v1.Now()

	return c.updateStatus(ctx, canary)
}

func (c *CanaryController) pauseDeployment(ctx context.Context, canary *deployv1alpha1.CanaryDeployment, reason string) error {
	canary.Status.Phase = "Paused"
	canary.Status.Reason = reason
	return c.updateStatus(ctx, canary)
}

func (c *CanaryController) finalizeDeployment(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	canary.Status.Phase = "Completed"
	canary.Status.CurrentWeight = 100
	return c.updateStatus(ctx, canary)
}

func (c *CanaryController) listCanaryDeployments(ctx context.Context) ([]*deployv1alpha1.CanaryDeployment, error) {
	return nil, nil
}

func (c *CanaryController) updateStatus(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	return nil
}
