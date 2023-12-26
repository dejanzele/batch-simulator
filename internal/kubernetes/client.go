package kubernetes

import (
	"errors"
	"log/slog"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	QPS   float32
	Burst int
}

// NewClient creates a new Kubernetes client and automatically detects in-cluster and out-of-cluster config.
func NewClient(kubeconfig *string, config Config) (kubernetes.Interface, error) {
	restConfig, err := loadRESTConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	restConfig.QPS = config.QPS
	restConfig.Burst = config.Burst

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// NewDynamicClient creates a new dynamic Kubernetes client and automatically detects in-cluster and out-of-cluster config.
func NewDynamicClient(kubeconfig *string, config Config) (dynamic.Interface, error) {
	restConfig, err := loadRESTConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	restConfig.QPS = config.QPS
	restConfig.Burst = config.Burst

	return dynamic.NewForConfig(restConfig)
}

// loadRESTConfig first tries to create an in-cluster config, and if it errors with ErrNotInCluster,
// it tries to create an out-of-cluster config based on provided kubeconfig path.
func loadRESTConfig(kubeconfig *string) (config *rest.Config, err error) {
	config, err = rest.InClusterConfig()
	if errors.Is(err, rest.ErrNotInCluster) {
		slog.Info("creating k8s config for out of cluster client")
		return clientcmd.BuildConfigFromFlags("", *kubeconfig)
	}
	slog.Info("running with in cluster client configuration")
	return config, err
}
