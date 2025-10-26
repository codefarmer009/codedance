package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type ActionType string

const (
	ContinueAction ActionType = "continue"
	PauseAction    ActionType = "pause"
	RollbackAction ActionType = "rollback"
)

type Decision struct {
	Action ActionType
	Reason string
	Score  int
}

var canaryGVR = schema.GroupVersionResource{
	Group:    "deploy.codedance.io",
	Version:  "v1alpha1",
	Resource: "canarydeployments",
}

type CanaryController struct {
	clientset       *kubernetes.Clientset
	dynamicClient   dynamic.Interface
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

func (c *CanaryController) SetDynamicClient(client dynamic.Interface) {
	c.dynamicClient = client
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

	if totalSteps == 0 {
		return fmt.Errorf("no deployment steps defined")
	}

	if currentStep >= totalSteps-1 {
		return c.finalizeDeployment(ctx, canary)
	}

	nextStep := currentStep + 1
	if nextStep >= totalSteps {
		return fmt.Errorf("next step %d exceeds total steps %d", nextStep, totalSteps)
	}

	weight := canary.Spec.Strategy.Steps[nextStep].Weight

	if err := c.trafficManager.UpdateWeight(ctx, canary, weight); err != nil {
		return fmt.Errorf("update traffic weight: %w", err)
	}

	canary.Status.CurrentStep = nextStep
	canary.Status.CurrentWeight = weight
	canary.Status.LastUpdateTime = metav1.Now()

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
	if c.dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client not initialized")
	}

	list, err := c.dynamicClient.Resource(canaryGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list canary deployments: %w", err)
	}

	canaries := make([]*deployv1alpha1.CanaryDeployment, 0, len(list.Items))
	for _, item := range list.Items {
		canary := &deployv1alpha1.CanaryDeployment{}
		if err := convertUnstructuredToCanary(&item, canary); err != nil {
			fmt.Printf("failed to convert canary %s: %v\n", item.GetName(), err)
			continue
		}
		canaries = append(canaries, canary)
	}

	return canaries, nil
}

func (c *CanaryController) updateStatus(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	if c.dynamicClient == nil {
		return fmt.Errorf("dynamic client not initialized")
	}

	unstructuredCanary, err := convertCanaryToUnstructured(canary)
	if err != nil {
		return fmt.Errorf("failed to convert canary to unstructured: %w", err)
	}

	_, err = c.dynamicClient.Resource(canaryGVR).
		Namespace(canary.Namespace).
		UpdateStatus(ctx, unstructuredCanary, metav1.UpdateOptions{})

	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func convertUnstructuredToCanary(u *unstructured.Unstructured, canary *deployv1alpha1.CanaryDeployment) error {
	data, err := u.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, canary)
}

func convertCanaryToUnstructured(canary *deployv1alpha1.CanaryDeployment) (*unstructured.Unstructured, error) {
	data, err := json.Marshal(canary)
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{}
	if err := u.UnmarshalJSON(data); err != nil {
		return nil, err
	}

	return u, nil
}
