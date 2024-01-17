package k8s

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	t.Parallel()

	t.Run("manager stops when context times out", func(t *testing.T) {
		fakeClient := fake.NewSimpleClientset()
		config := ManagerConfig{
			NodeRateLimiterConfig: RateLimiterConfig{
				Frequency: 10 * time.Millisecond,
				Requests:  1,
				Limit:     10,
			},
			PodRateLimiterConfig: RateLimiterConfig{
				Frequency: 10 * time.Millisecond,
				Requests:  1,
				Limit:     10,
			},
		}
		manager := NewManager(fakeClient, &config)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Millisecond)
		defer cancel()

		var err error
		go func() {
			err = manager.Start(ctx)
		}()

		assert.Eventually(t, func() bool {
			return errors.Is(err, context.DeadlineExceeded)
		},
			30*time.Millisecond,
			10*time.Millisecond,
		)
	})

	t.Run("manager stops when context is cancelled", func(t *testing.T) {
		fakeClient := fake.NewSimpleClientset()
		config := ManagerConfig{
			NodeRateLimiterConfig: RateLimiterConfig{
				Frequency: 10 * time.Millisecond,
				Requests:  1,
			},
			PodRateLimiterConfig: RateLimiterConfig{
				Frequency: 10 * time.Millisecond,
				Requests:  1,
			},
		}
		manager := NewManager(fakeClient, &config)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var err error
		go func() {
			err = manager.Start(ctx)
		}()

		time.Sleep(10 * time.Millisecond)
		cancel()

		assert.Eventually(t, func() bool {
			return errors.Is(err, context.Canceled)
		},
			30*time.Millisecond,
			10*time.Millisecond,
		)
	})
}

func TestManager_Start(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	config := ManagerConfig{
		NodeRateLimiterConfig: RateLimiterConfig{
			Frequency: 10 * time.Millisecond,
			Requests:  2,
			Limit:     3,
		},
		PodRateLimiterConfig: RateLimiterConfig{
			Frequency: 10 * time.Millisecond,
			Requests:  2,
			Limit:     3,
		},
	}
	manager := NewManager(fakeClient, &config)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	go manager.Start(ctx)

	assert.Eventually(t, func() bool {
		nodeList, _ := fakeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		podList, _ := fakeClient.CoreV1().Pods("default").List(ctx, metav1.ListOptions{})
		return len(nodeList.Items) == 3 && len(podList.Items) == 3
	},
		100*time.Millisecond,
		20*time.Millisecond,
	)
}
