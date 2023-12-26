package config

import "time"

var (
	// QPS configures maximum queries per second to use while talking with Kubernetes API.
	QPS float32 = 500
	// Burst configures maximum burst for throttle while talking with Kubernetes API.
	Burst = 500
	// Kubeconfig is the path to the kubeconfig file. Defaults to $HOME/.kube/config.
	Kubeconfig string
	// NoGUI disables printing graphical elements like spinners, progress bars...
	NoGUI bool
	// Verbose configures verbose output.
	Verbose bool
	// Silent disables all internal logs.
	Silent bool
	// Debug enables debug output.
	Debug bool
	// KWOKNamespace is the namespace in which kwok-operator is expected or installed.
	KWOKNamespace = "kube-system"
	// Namespace is the namespace in which pods should be created.
	Namespace = "default"
	// PodCreatorFrequency is the frequency at which the pod creator should be invoked.
	PodCreatorFrequency = 1 * time.Second
	// PodCreatorRequests is the number of requests that should be made to the pod creator in each iteration.
	PodCreatorRequests int32 = 5
	// PodCreatorLimit is the maximum number of pods that should be created.
	PodCreatorLimit int32
	// NodeCreatorFrequency is the frequency at which the node creator should be invoked.
	NodeCreatorFrequency = 1 * time.Second
	// NodeCreatorRequests is the number of requests that should be made to the node creator in each iteration.
	NodeCreatorRequests int32 = 2
	// NodeCreatorLimit is the maximum number of nodes that should be created.
	NodeCreatorLimit int32
	// JobCreatorFrequency is the frequency at which the job creator should be invoked.
	JobCreatorFrequency = 1 * time.Second
	// JobCreatorRequests is the number of requests that should be made to the job creator in each iteration.
	JobCreatorRequests int32 = 2
	// JobCreatorLimit is the maximum number of jobs that should be created.
	JobCreatorLimit int32
	// DefaultPollInterval is the default interval at which the polling functions should be invoked.
	DefaultPollInterval = 2 * time.Second
	// DefaultPollTimeout is the default timeout for polling functions.
	DefaultPollTimeout = 30 * time.Second
)
