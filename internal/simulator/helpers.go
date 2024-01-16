package simulator

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os/exec"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
)

var (
	ErrOperatorNotInstalled = errors.New("kwok-operator is not installed")
	//go:embed "data/stages.yaml"
	kwokStages                  string
	stagesSchema                = schema.GroupVersionResource{Group: "kwok.x-k8s.io", Version: "v1alpha1", Resource: "stages"}
	stageNodeHeartbeatWithLease = "node-heartbeat-with-lease"
	stageNodeInitialize         = "node-initialize"
	stagePodComplete            = "pod-complete"
	stagePodDelete              = "pod-delete"
	stagePodReady               = "pod-ready"
	kwokRepository              = "kubernetes-sigs/kwok"
	kwokVersion                 = "v0.4.0"
	kwokOperatorManifest        = fmt.Sprintf("https://github.com/%s/releases/download/%s/kwok.yaml", kwokRepository, kwokVersion)
)

// InstallOperator installs kwok-operator in the cluster.
func InstallOperator(ctx context.Context, namespace string) (output []byte, err error) {
	output, err = exec.CommandContext(ctx, "kubectl", "apply", "--filename", kwokOperatorManifest, "--namespace", namespace).CombinedOutput()
	return
}

// UninstallOperator uninstalls kwok-operator from the cluster.
func UninstallOperator(ctx context.Context, namespace string) (output []byte, err error) {
	output, err = exec.CommandContext(ctx, "kubectl", "delete", "--filename", kwokOperatorManifest, "--namespace", namespace).CombinedOutput()
	return
}

// CreateStages creates the kwok stages in the cluster required for node and pod lifecycle.
func CreateStages(ctx context.Context) (output []byte, err error) {
	// Create the kubectl apply command
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	// Create a buffer with the kwokStages string and use it as stdin
	cmd.Stdin = bytes.NewBufferString(kwokStages)
	// Run the command and capture the combined output (stdout and stderr)
	return cmd.CombinedOutput()

}

// DeleteStages deletes the kwok stages from the cluster.
func DeleteStages(ctx context.Context) (output []byte, err error) {
	// Create the kubectl delete command
	cmd := exec.CommandContext(ctx, "kubectl", "delete", "-f", "-")
	// Create a buffer with the kwokStages string and use it as stdin
	cmd.Stdin = bytes.NewBufferString(kwokStages)
	// Run the command and capture the combined output (stdout and stderr)
	return cmd.CombinedOutput()
}

func CheckAreStagesCreated(ctx context.Context, client dynamic.Interface) (found bool, missing []string, err error) {
	_, err = client.Resource(stagesSchema).Get(ctx, stageNodeHeartbeatWithLease, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			missing = append(missing, stageNodeHeartbeatWithLease)
		} else {
			return false, nil, err
		}
	}
	_, err = client.Resource(stagesSchema).Get(ctx, stageNodeInitialize, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			missing = append(missing, stageNodeInitialize)
		} else {
			return false, nil, err
		}
	}
	_, err = client.Resource(stagesSchema).Get(ctx, stagePodComplete, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			missing = append(missing, stagePodComplete)
		} else {
			return false, nil, err
		}
	}
	_, err = client.Resource(stagesSchema).Get(ctx, stagePodDelete, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			missing = append(missing, stagePodDelete)
		} else {
			return false, nil, err
		}
	}
	_, err = client.Resource(stagesSchema).Get(ctx, stagePodReady, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			missing = append(missing, stagePodReady)
		} else {
			return false, nil, err
		}
	}
	return len(missing) == 0, missing, nil
}

// CheckIsKWOKInstalled checks if the kwok binary is installed in the system.
func CheckIsKWOKInstalled(ctx context.Context) (output []byte, installed bool) {
	output, err := exec.CommandContext(ctx, "kwok", "--version").CombinedOutput()
	return output, err == nil
}

// CheckIsKubectlInstalled checks if the kubectl binary is installed in the system.
func CheckIsKubectlInstalled(ctx context.Context) (output []byte, installed bool) {
	output, err := exec.CommandContext(ctx, "kubectl", "version").CombinedOutput()
	return output, err == nil
}

// CheckIsOperatorRunning checks if kwok-operator is installed & running with at least 1 replica in the cluster.
func CheckIsOperatorRunning(ctx context.Context, client kubernetes.Interface, namespace string) (output []byte, running bool, err error) {
	err = wait.PollUntilContextTimeout(
		ctx,
		config.DefaultPollInterval,
		2*config.DefaultPollTimeout,
		false,
		func(ctx context.Context) (done bool, err error) {
			deployment, err := client.AppsV1().Deployments(namespace).Get(ctx, "kwok-controller", metav1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return false, ErrOperatorNotInstalled
				}
				return false, err
			}
			if deployment.Status.AvailableReplicas == 0 {
				return false, nil
			}
			return true, nil
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrOperatorNotInstalled):
			return []byte("kwok-operator is not installed"), false, nil
		case errors.Is(err, context.DeadlineExceeded):
			return []byte("timed out waiting for kwok-operator start "), false, nil
		default:
			return nil, false, fmt.Errorf("failed to check if kwok-operator is running: %w", err)
		}
	}
	return []byte("kwok-operator is running"), true, nil
}
