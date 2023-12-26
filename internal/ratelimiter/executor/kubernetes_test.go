package executor

import (
	"context"
	"errors"
	"github.com/dejanzele/batch-simulator/internal/ratelimiter"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	k8stesting "k8s.io/client-go/testing"
	"testing"
)

func TestPodCreator(t *testing.T) {
	t.Parallel()

	t.Run("pod creation succeeds", func(t *testing.T) {
		t.Parallel()

		fakeClient := fake.NewSimpleClientset()
		executor := NewPodCreator(fakeClient, "default")

		ctx := context.Background()
		testPod := corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test-container",
						Image: "test-image",
					},
				},
			},
		}
		if err := executor.Execute(ctx, &testPod); err != nil {
			t.Fatalf("failed to create pod: %v", err)
		}
		pods, err := fakeClient.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("failed to list pods: %v", err)
		}
		assert.Len(t, pods.Items, 1)
		assert.Equal(t, "test-pod", pods.Items[0].Name)
		assert.Equal(t, "test-container", pods.Items[0].Spec.Containers[0].Name)
		assert.Equal(t, "test-image", pods.Items[0].Spec.Containers[0].Image)
		assert.Equal(t, "kubernetes-pod-creator", executor.Identifier())
	})

	t.Run("pod creation returns error", func(t *testing.T) {
		t.Parallel()

		fakeClient := fake.NewSimpleClientset()
		fakeClient.
			CoreV1().(*fakecorev1.FakeCoreV1).
			PrependReactor("create", "pods", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.Pod{}, errors.New("error creating pod")
			})
		executor := NewPodCreator(fakeClient, "default")

		ctx := context.Background()
		testPod := corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "test-container",
						Image: "test-image",
					},
				},
			},
		}
		err := executor.Execute(ctx, &testPod)
		var createError *ratelimiter.CreateError
		assert.ErrorAs(t, err, &createError)
		assert.Equal(t, "v1", createError.APIGroup)
		assert.Equal(t, "Pod", createError.Kind)
		assert.Equal(t, "test-pod", createError.Resource.GetName())
		assert.Equal(t, "default", createError.Resource.GetNamespace())
		assert.Equal(t, "error creating pod", createError.Err.Error())
	})
}

func TestNodeCreator(t *testing.T) {
	t.Parallel()

	t.Run("node creation succeeds", func(t *testing.T) {
		t.Parallel()

		fakeClient := fake.NewSimpleClientset()
		executor := NewNodeCreator(fakeClient)

		ctx := context.Background()
		testNode := corev1.Node{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Node",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		}
		if err := executor.Execute(ctx, &testNode); err != nil {
			t.Fatalf("failed to create pod: %v", err)
		}
		nodes, err := fakeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("failed to list nodes: %v", err)
		}
		assert.Len(t, nodes.Items, 1)
		assert.Equal(t, "test-node", nodes.Items[0].Name)
		assert.Equal(t, "kubernetes-node-creator", executor.Identifier())
	})

	t.Run("node creation returns error", func(t *testing.T) {
		t.Parallel()

		fakeClient := fake.NewSimpleClientset()
		fakeClient.
			CoreV1().(*fakecorev1.FakeCoreV1).
			PrependReactor("create", "nodes", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &corev1.Node{}, errors.New("error creating node")
			})
		executor := NewNodeCreator(fakeClient)

		ctx := context.Background()
		testNode := corev1.Node{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Node",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-node",
			},
		}
		err := executor.Execute(ctx, &testNode)
		var createError *ratelimiter.CreateError
		assert.ErrorAs(t, err, &createError)
		assert.Equal(t, "v1", createError.APIGroup)
		assert.Equal(t, "Node", createError.Kind)
		assert.Equal(t, "test-node", createError.Resource.GetName())
		assert.Equal(t, "", createError.Resource.GetNamespace())
		assert.Equal(t, "error creating node", createError.Err.Error())
	})
}
