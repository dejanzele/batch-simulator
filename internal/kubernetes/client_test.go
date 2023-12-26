package kubernetes

import (
	"context"
	"github.com/dejanzele/batch-simulator/internal/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestNewClient_Integration(t *testing.T) {
	test.IntegrationTest(t)
	t.Parallel()

	client, err := NewClient(test.GetKubeconfig(), Config{})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("failed to list pods: %v", err)
	}
}
