package controller

import (
	"context"
	"fmt"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DefaultRollbackManager struct {
	clientset      *kubernetes.Clientset
	trafficManager TrafficManager
}

func NewDefaultRollbackManager(clientset *kubernetes.Clientset, trafficManager TrafficManager) *DefaultRollbackManager {
	return &DefaultRollbackManager{
		clientset:      clientset,
		trafficManager: trafficManager,
	}
}

func (r *DefaultRollbackManager) Rollback(ctx context.Context, canary *deployv1alpha1.CanaryDeployment, reason string) error {
	fmt.Printf("Rolling back canary %s: %s\n", canary.Name, reason)

	if err := r.trafficManager.UpdateWeight(ctx, canary, 0); err != nil {
		return fmt.Errorf("revert traffic: %w", err)
	}

	if err := r.deleteCanaryPods(ctx, canary); err != nil {
		return fmt.Errorf("delete canary pods: %w", err)
	}

	canary.Status.Phase = "Failed"
	canary.Status.Reason = reason
	canary.Status.CurrentWeight = 0

	return nil
}

func (r *DefaultRollbackManager) deleteCanaryPods(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	deployment, err := r.clientset.AppsV1().
		Deployments(canary.Namespace).
		Get(ctx, canary.Spec.TargetDeployment+"-canary", metav1.GetOptions{})
	if err != nil {
		return err
	}

	replicas := int32(0)
	deployment.Spec.Replicas = &replicas

	_, err = r.clientset.AppsV1().
		Deployments(canary.Namespace).
		Update(ctx, deployment, metav1.UpdateOptions{})

	return err
}
