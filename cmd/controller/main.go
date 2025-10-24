package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codefarmer009/codedance/pkg/controller"
	"github.com/codefarmer009/codedance/pkg/metrics"
	"github.com/codefarmer009/codedance/pkg/traffic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig  string
	prometheusURL string
	useIstio    bool
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	flag.StringVar(&prometheusURL, "prometheus-url", "http://prometheus:9090", "Prometheus server URL")
	flag.BoolVar(&useIstio, "use-istio", true, "Use Istio for traffic management")
}

func main() {
	flag.Parse()

	config, err := buildConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build config: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create clientset: %v\n", err)
		os.Exit(1)
	}

	metricsAnalyzer, err := metrics.NewPrometheusAnalyzer(prometheusURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create metrics analyzer: %v\n", err)
		os.Exit(1)
	}

	var trafficManager controller.TrafficManager
	if useIstio {
		trafficManager = traffic.NewNginxTrafficManager(clientset)
	} else {
		trafficManager = traffic.NewNginxTrafficManager(clientset)
	}

	decisionEngine := controller.NewDefaultDecisionEngine()
	rollbackManager := controller.NewDefaultRollbackManager(clientset, trafficManager)

	canaryController := controller.NewCanaryController(
		clientset,
		trafficManager,
		metricsAnalyzer,
		decisionEngine,
		rollbackManager,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("Received shutdown signal")
		cancel()
	}()

	fmt.Println("Starting Canary Controller...")
	if err := canaryController.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Controller error: %v\n", err)
		os.Exit(1)
	}
}

func buildConfig() (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
