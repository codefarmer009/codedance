package traffic

import (
	"context"
	"fmt"
	"strconv"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NginxTrafficManager struct {
	clientset *kubernetes.Clientset
}

func NewNginxTrafficManager(clientset *kubernetes.Clientset) *NginxTrafficManager {
	return &NginxTrafficManager{
		clientset: clientset,
	}
}

func (m *NginxTrafficManager) UpdateWeight(ctx context.Context, canary *deployv1alpha1.CanaryDeployment, weight int) error {
	ingress, err := m.clientset.NetworkingV1().
		Ingresses(canary.Namespace).
		Get(ctx, canary.Name+"-canary", metav1.GetOptions{})
	if err != nil {
		return err
	}

	if ingress.Annotations == nil {
		ingress.Annotations = make(map[string]string)
	}
	ingress.Annotations["nginx.ingress.kubernetes.io/canary-weight"] = strconv.Itoa(weight)

	_, err = m.clientset.NetworkingV1().
		Ingresses(canary.Namespace).
		Update(ctx, ingress, metav1.UpdateOptions{})

	return err
}

func (m *NginxTrafficManager) CreateCanaryRoute(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      canary.Name + "-canary",
			Namespace: canary.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/canary":        "true",
				"nginx.ingress.kubernetes.io/canary-weight": "0",
			},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: canary.Name + ".example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: fmt.Sprintf("%s-canary", canary.Spec.TargetDeployment),
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := m.clientset.NetworkingV1().
		Ingresses(canary.Namespace).
		Create(ctx, ingress, metav1.CreateOptions{})

	return err
}
