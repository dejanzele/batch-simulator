package simulator

import (
	"context"
	"github.com/dejanzele/batch-simulator/internal/kubernetes"
	"github.com/dejanzele/batch-simulator/internal/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/dynamic"
	k8sclient "k8s.io/client-go/kubernetes"
	"testing"
)

func TestIsKWOKInstalled_Integration(t *testing.T) {
	test.IntegrationTest(t)
	t.Parallel()

	output, installed := CheckIsKWOKInstalled(context.Background())
	assert.True(t, installed, "kwok is not installed")
	assert.Contains(t, string(output), "kwok version", "kwok version is not printed")
}

func TestIsKubectlInstalled_Integration(t *testing.T) {
	test.IntegrationTest(t)
	t.Parallel()

	output, installed := CheckIsKubectlInstalled(context.Background())
	assert.True(t, installed, "kubectl is not installed")
	assert.Contains(t, string(output), "Client Version", "Client Version is not printed")
}

func TestKWOKOperator_Integration(t *testing.T) {
	test.IntegrationTest(t)
	t.Parallel()

	client, err := kubernetes.NewClient(test.GetKubeconfig(), kubernetes.Config{})
	if err != nil {
		t.Fatalf("failed to create k8s client: %v", err)
	}

	dynamicClient, err := kubernetes.NewDynamicClient(test.GetKubeconfig(), kubernetes.Config{})
	if err != nil {
		t.Fatalf("failed to create dynamic client: %v", err)
	}

	ctx := context.Background()
	namespace := "kube-system"
	testInstallKWOKOperator(ctx, t, namespace)

	testCheckIsKWOKOperatorRunning(ctx, t, client, namespace)

	testCreateStages(ctx, t)

	testCheckAreStagesCreated(ctx, t, dynamicClient)

	testDeleteStages(ctx, t)

	testUninstallKWOKOperator(ctx, t, namespace)
}

func testInstallKWOKOperator(ctx context.Context, t *testing.T, namespace string) {
	t.Helper()

	output, err := InstallOperator(ctx, namespace)
	if err != nil {
		t.Fatalf("failed to install kwok operator: %v", err)
	}
	assert.NotEmpty(t, output)
}

func testCheckIsKWOKOperatorRunning(ctx context.Context, t *testing.T, client k8sclient.Interface, namespace string) { //nolint
	t.Helper()

	output, running, err := CheckIsOperatorRunning(ctx, client, namespace)
	assert.NoErrorf(t, err, "failed to wait for kwok operator to be installed & running: %v", err)
	assert.True(t, running, "kwok operator is not running")
	assert.NotEmpty(t, output)
}

func testCreateStages(ctx context.Context, t *testing.T) {
	t.Helper()

	output, err := CreateStages(ctx)
	if err != nil {
		t.Fatalf("failed to create stages: %v", err)
	}
	assert.NotEmpty(t, output)
}

func testCheckAreStagesCreated(ctx context.Context, t *testing.T, dynamicClient dynamic.Interface) {
	t.Helper()

	installed, missing, err := CheckAreStagesCreated(ctx, dynamicClient)
	if err != nil {
		t.Fatalf("failed to check if stages are created: %v", err)
	}
	assert.True(t, installed, "stages are not installed")
	assert.Empty(t, missing, "stages are missing")
}

func testDeleteStages(ctx context.Context, t *testing.T) {
	t.Helper()

	output, err := DeleteStages(ctx)
	if err != nil {
		t.Fatalf("failed to delete stages: %v", err)
	}
	assert.NotEmpty(t, output)
}

func testUninstallKWOKOperator(ctx context.Context, t *testing.T, namespace string) {
	t.Helper()

	output, err := UninstallOperator(ctx, namespace)
	if err != nil {
		t.Fatalf("failed to uninstall kwok operator: %v", err)
	}
	assert.NotEmpty(t, output)
}
