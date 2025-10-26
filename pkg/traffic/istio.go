package traffic

import (
	"context"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
	networkingv1beta1 "istio.io/api/networking/v1beta1"
	"istio.io/client-go/pkg/apis/networking/v1beta1"
	versionedclient "istio.io/client-go/pkg/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func NewIstioClient(config *rest.Config) (versionedclient.Interface, error) {
	return versionedclient.NewForConfig(config)
}

type IstioTrafficManager struct {
	istioClient versionedclient.Interface
}

func NewIstioTrafficManager(istioClient versionedclient.Interface) *IstioTrafficManager {
	return &IstioTrafficManager{
		istioClient: istioClient,
	}
}

func (m *IstioTrafficManager) UpdateWeight(ctx context.Context, canary *deployv1alpha1.CanaryDeployment, weight int) error {
	vs, err := m.istioClient.NetworkingV1beta1().
		VirtualServices(canary.Namespace).
		Get(ctx, canary.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	for i, route := range vs.Spec.Http {
		if len(route.Route) == 2 {
			vs.Spec.Http[i].Route[0].Weight = int32(100 - weight)
			vs.Spec.Http[i].Route[1].Weight = int32(weight)
		}
	}

	_, err = m.istioClient.NetworkingV1beta1().
		VirtualServices(canary.Namespace).
		Update(ctx, vs, metav1.UpdateOptions{})

	return err
}

func (m *IstioTrafficManager) CreateCanaryRoute(ctx context.Context, canary *deployv1alpha1.CanaryDeployment) error {
	vs := &v1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      canary.Name,
			Namespace: canary.Namespace,
		},
		Spec: networkingv1beta1.VirtualService{
			Hosts: []string{canary.Name + ".example.com"},
			Http: []*networkingv1beta1.HTTPRoute{
				{
					Route: []*networkingv1beta1.HTTPRouteDestination{
						{
							Destination: &networkingv1beta1.Destination{
								Host:   canary.Spec.TargetDeployment,
								Subset: "stable",
							},
							Weight: 100,
						},
						{
							Destination: &networkingv1beta1.Destination{
								Host:   canary.Spec.TargetDeployment,
								Subset: "canary",
							},
							Weight: 0,
						},
					},
				},
			},
		},
	}

	_, err := m.istioClient.NetworkingV1beta1().
		VirtualServices(canary.Namespace).
		Create(ctx, vs, metav1.CreateOptions{})

	return err
}
