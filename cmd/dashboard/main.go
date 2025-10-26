package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	deployv1alpha1 "github.com/codefarmer009/codedance/pkg/apis/deploy/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
	port       int
	clientset  *kubernetes.Clientset
	dynClient  dynamic.Interface
)

var canaryGVR = schema.GroupVersionResource{
	Group:    "deploy.codedance.io",
	Version:  "v1alpha1",
	Resource: "canarydeployments",
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	flag.IntVar(&port, "port", 8080, "Dashboard server port")
}

func main() {
	flag.Parse()

	config, err := buildConfig()
	if err != nil {
		log.Fatalf("Failed to build config: %v", err)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create clientset: %v", err)
	}

	dynClient, err = dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("Failed to create dynamic client: %v", err)
	}

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/canaries", handleCanaryList)
	http.HandleFunc("/api/canaries/", handleCanaryDetail)
	http.HandleFunc("/api/metrics/", handleMetrics)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting dashboard server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func buildConfig() (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/templates/index.html")
}

func handleCanaryList(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	list, err := dynClient.Resource(canaryGVR).Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list canaries: %v", err), http.StatusInternalServerError)
		return
	}

	canaries := make([]map[string]interface{}, 0)
	for _, item := range list.Items {
		canary := convertToCanaryInfo(&item)
		canaries = append(canaries, canary)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(canaries)
}

func handleCanaryDetail(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		http.Error(w, "namespace and name are required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	item, err := dynClient.Resource(canaryGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get canary: %v", err), http.StatusInternalServerError)
		return
	}

	canary := &deployv1alpha1.CanaryDeployment{}
	if err := convertUnstructuredToCanary(item, canary); err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert canary: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(canary)
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	if namespace == "" || name == "" {
		http.Error(w, "namespace and name are required", http.StatusBadRequest)
		return
	}

	metrics := map[string]interface{}{
		"timestamp":   time.Now().Unix(),
		"successRate": 99.5,
		"errorRate":   0.5,
		"latencyP50":  45.2,
		"latencyP90":  89.5,
		"latencyP99":  156.8,
		"requestRate": 1250.5,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func convertToCanaryInfo(u *unstructured.Unstructured) map[string]interface{} {
	spec, _, _ := unstructured.NestedMap(u.Object, "spec")
	status, _, _ := unstructured.NestedMap(u.Object, "status")

	return map[string]interface{}{
		"name":          u.GetName(),
		"namespace":     u.GetNamespace(),
		"phase":         status["phase"],
		"currentStep":   status["currentStep"],
		"currentWeight": status["currentWeight"],
		"strategy":      getStrategyType(spec),
		"createdAt":     u.GetCreationTimestamp().Format(time.RFC3339),
	}
}

func getStrategyType(spec map[string]interface{}) string {
	if strategy, ok := spec["strategy"].(map[string]interface{}); ok {
		if strategyType, ok := strategy["type"].(string); ok {
			return strategyType
		}
	}
	return "unknown"
}

func convertUnstructuredToCanary(u *unstructured.Unstructured, canary *deployv1alpha1.CanaryDeployment) error {
	data, err := u.MarshalJSON()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, canary)
}
