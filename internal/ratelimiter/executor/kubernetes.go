package executor

import (
	"context"

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

func (p *PodCreator) Identifier() string {
	return "kubernetes-pod-creator"
}

func (p *PodCreator) Execute(ctx context.Context, item *corev1.Pod) error {
	_, err := p.client.CoreV1().Pods(p.namespace).Create(ctx, item, metav1.CreateOptions{})
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

func (n *NodeCreator) Identifier() string {
	return "kubernetes-node-creator"
}

func (n *NodeCreator) Execute(ctx context.Context, item *corev1.Node) error {
	_, err := n.client.CoreV1().Nodes().Create(ctx, item, metav1.CreateOptions{})
	if err != nil {
		return ratelimiter.NewCreateError(err, "v1", "Node", item)
	}
	return nil
}

var _ ratelimiter.Executor[*corev1.Node] = &NodeCreator{}
