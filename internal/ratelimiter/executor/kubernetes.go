package executor

import (
	"context"
	"fmt"

	"github.com/dejanzele/batch-simulator/internal/simulator/resources"
	"github.com/dejanzele/batch-simulator/internal/util"

	batchv1 "k8s.io/api/batch/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/dejanzele/batch-simulator/internal/ratelimiter"
)

// kubernetesExecutor is defines base fields for kubernetes executors.
type kubernetesExecutor struct {
	client    kubernetes.Interface
	namespace string
}

// PodCreator is used to create Pods.
type PodCreator struct {
	kubernetesExecutor
}

func NewPodCreator(client kubernetes.Interface, namespace string) *PodCreator {
	return &PodCreator{
		kubernetesExecutor: kubernetesExecutor{
			client:    client,
			namespace: namespace,
		},
	}
}

// Identifier returns the executor identifier.
func (c *PodCreator) Identifier() string {
	return "kubernetes-pod-creator"
}

// Execute creates a Pod.
func (c *PodCreator) Execute(ctx context.Context) error {
	name := fmt.Sprintf("fake-pod-%s", util.RandomRFC1123Name(16))
	item := resources.NewFakePod(name, c.namespace)
	_, err := c.client.CoreV1().Pods(c.namespace).Create(ctx, item, metav1.CreateOptions{})
	if err != nil {
		return ratelimiter.NewCreateError(err, "v1", "Pod", item)
	}
	return nil
}

var _ ratelimiter.Executor[*corev1.Pod] = &PodCreator{}

type NodeCreator struct {
	kubernetesExecutor
}

func NewNodeCreator(client kubernetes.Interface) *NodeCreator {
	return &NodeCreator{
		kubernetesExecutor: kubernetesExecutor{
			client: client,
		},
	}
}

// Identifier returns the executor identifier.
func (c *NodeCreator) Identifier() string {
	return "kubernetes-node-creator"
}

// Execute creates a Node.
func (c *NodeCreator) Execute(ctx context.Context) error {
	name := fmt.Sprintf("fake-node-%s", util.RandomRFC1123Name(16))
	item := resources.NewFakeNode(name)
	_, err := c.client.CoreV1().Nodes().Create(ctx, item, metav1.CreateOptions{})
	if err != nil {
		return ratelimiter.NewCreateError(err, "v1", "Node", item)
	}
	return nil
}

var _ ratelimiter.Executor[*corev1.Node] = &NodeCreator{}

type JobCreator struct {
	kubernetesExecutor
}

func NewJobCreator(client kubernetes.Interface, namespace string) *JobCreator {
	return &JobCreator{
		kubernetesExecutor: kubernetesExecutor{
			client:    client,
			namespace: namespace,
		},
	}
}

// Identifier returns the executor identifier.
func (c *JobCreator) Identifier() string {
	return "kubernetes-job-creator"
}

// Execute creates a Node.
func (c *JobCreator) Execute(ctx context.Context) error {
	name := fmt.Sprintf("fake-job-%s", util.RandomRFC1123Name(16))
	item := resources.NewFakeJob(name, c.namespace)
	_, err := c.client.BatchV1().Jobs(c.namespace).Create(ctx, item, metav1.CreateOptions{})
	if err != nil {
		return ratelimiter.NewCreateError(err, "batch/v1", "Job", item)
	}
	return nil
}

var _ ratelimiter.Executor[*batchv1.Job] = &JobCreator{}
