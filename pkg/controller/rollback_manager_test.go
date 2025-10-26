package controller

import (
	"context"
	"testing"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type mockTrafficManager struct {
	updateWeightCalled bool
	lastWeight         int
	shouldError        bool
}

func (m *mockTrafficManager) UpdateWeight(ctx context.Context, canary *deployv1alpha1.CanaryDeployment, weight int) error {
	m.updateWeightCalled = true
	m.lastWeight = weight
	if m.shouldError {
		return context.DeadlineExceeded
	}
	return nil
}

func (m *mockTrafficManager) CreateCanaryRoute(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	return nil
}

func TestNewDefaultRollbackManager(t *testing.T) {
	mockTM := &mockTrafficManager{}
	manager := NewDefaultRollbackManager(nil, mockTM)

	if manager == nil {
		t.Fatal("NewDefaultRollbackManager() returned nil")
	}
	if manager.trafficManager == nil {
		t.Error("trafficManager not set")
	}
}

func TestRollbackManager_SetController(t *testing.T) {
	mockTM := &mockTrafficManager{}
	manager := NewDefaultRollbackManager(nil, mockTM)
	
	var controller *CanaryController
	
	manager.SetController(controller)
	
	if manager.controller != controller {
		t.Error("SetController() did not set controller")
	}
}

func TestRollbackManager_Rollback_UpdatesTraffic(t *testing.T) {
	mockTM := &mockTrafficManager{}
	
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app-canary",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { r := int32(3); return &r }(),
		},
	}
	
	fakeClient := fake.NewSimpleClientset(deployment)
	manager := NewDefaultRollbackManager(kubernetes.Interface(fakeClient), mockTM)

	canary := &deployv1alpha1.CanaryDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-canary",
			Namespace: "default",
		},
		Spec: deployv1alpha1.CanaryDeploymentSpec{
			TargetDeployment: "test-app",
		},
	}

	ctx := context.Background()
	err := manager.Rollback(ctx, canary, "test rollback")
	if err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}

	if !mockTM.updateWeightCalled {
		t.Error("Rollback() did not call UpdateWeight")
	}
	if mockTM.lastWeight != 0 {
		t.Errorf("Rollback() weight = %d, want 0", mockTM.lastWeight)
	}
}

func TestRollbackManager_Rollback_UpdatesStatus(t *testing.T) {
	mockTM := &mockTrafficManager{}
	
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app-canary",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { r := int32(3); return &r }(),
		},
	}
	
	fakeClient := fake.NewSimpleClientset(deployment)
	manager := NewDefaultRollbackManager(kubernetes.Interface(fakeClient), mockTM)

	canary := &deployv1alpha1.CanaryDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-canary",
			Namespace: "default",
		},
		Spec: deployv1alpha1.CanaryDeploymentSpec{
			TargetDeployment: "test-app",
		},
	}

	ctx := context.Background()
	reason := "test rollback reason"
	err := manager.Rollback(ctx, canary, reason)
	if err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}

	if canary.Status.Phase != "Failed" {
		t.Errorf("Rollback() Phase = %s, want Failed", canary.Status.Phase)
	}
	if canary.Status.Reason != reason {
		t.Errorf("Rollback() Reason = %s, want %s", canary.Status.Reason, reason)
	}
	if canary.Status.CurrentWeight != 0 {
		t.Errorf("Rollback() CurrentWeight = %d, want 0", canary.Status.CurrentWeight)
	}
}

func TestRollbackManager_Rollback_ScalesDownCanaryPods(t *testing.T) {
	mockTM := &mockTrafficManager{}
	
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app-canary",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { r := int32(3); return &r }(),
		},
	}
	
	fakeClient := fake.NewSimpleClientset(deployment)
	manager := NewDefaultRollbackManager(kubernetes.Interface(fakeClient), mockTM)

	canary := &deployv1alpha1.CanaryDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-canary",
			Namespace: "default",
		},
		Spec: deployv1alpha1.CanaryDeploymentSpec{
			TargetDeployment: "test-app",
		},
	}

	ctx := context.Background()
	err := manager.Rollback(ctx, canary, "test rollback")
	if err != nil {
		t.Errorf("Rollback() error = %v, want nil", err)
	}

	// Verify the deployment was scaled down
	updatedDeployment, err := fakeClient.AppsV1().Deployments("default").Get(ctx, "test-app-canary", metav1.GetOptions{})
	if err != nil {
		t.Errorf("Failed to get updated deployment: %v", err)
	}
	
	if updatedDeployment.Spec.Replicas == nil || *updatedDeployment.Spec.Replicas != 0 {
		t.Errorf("Deployment replicas = %v, want 0", updatedDeployment.Spec.Replicas)
	}
}
