package config

import (
	"time"
)

var (
	// QPS configures maximum queries per second to use while talking with Kubernetes API.
	QPS float32 = 2000
	// Burst configures maximum burst for throttle while talking with Kubernetes API.
	Burst = 2000
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
	// SimulatorNamespace is the namespace in which simulator pods should be created.
	SimulatorNamespace = "default"
	// PodCreatorFrequency is the frequency at which the pod creator should be invoked.
	PodCreatorFrequency = 1 * time.Second
	// PodCreatorRequests is the number of requests that should be made to the pod creator in each iteration.
	PodCreatorRequests = 5
	// PodCreatorLimit is the maximum number of pods that should be created.
	PodCreatorLimit int
	// NodeCreatorFrequency is the frequency at which the node creator should be invoked.
	NodeCreatorFrequency = 1 * time.Second
	// NodeCreatorRequests is the number of requests that should be made to the node creator in each iteration.
	NodeCreatorRequests = 2
	// NodeCreatorLimit is the maximum number of nodes that should be created.
	NodeCreatorLimit int
	// JobCreatorFrequency is the frequency at which the job creator should be invoked.
	JobCreatorFrequency = 1 * time.Second
	// JobCreatorRequests is the number of requests that should be made to the job creator in each iteration.
	JobCreatorRequests = 2
	// JobCreatorLimit is the maximum number of jobs that should be created.
	JobCreatorLimit int
	// DefaultPollInterval is the default interval at which the polling functions should be invoked.
	DefaultPollInterval = 2 * time.Second
	// DefaultPollTimeout is the default timeout for polling functions.
	DefaultPollTimeout = 150 * time.Second
	// Remote configures whether the simulator should be executed in a Kubernetes cluster.
	Remote bool
	// PodSpecSize is the size of the pod spec in bytes.
	PodSpecSize = 50 * 1024
	// SimulatorImage is the image used for the simulator.
	SimulatorImage = "dpejcev/batchsim"
	// SimulatorTag is the tag used for the simulator.
	SimulatorTag = "latest"
	// RandomEnvVars configures whether the simulator should use random envvars.
	RandomEnvVars = false
	// DefaultEnvVarsType is the default envvar type which are generated when creating fake pods.
	DefaultEnvVarsType = "medium"
)
