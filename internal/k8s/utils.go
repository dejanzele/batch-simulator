package k8s

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WaitForJobPodsReady waits for all pods associated with the given job to become ready.
func WaitForJobPodsReady(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace string,
	jobName string,
	timeout time.Duration,
) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	watch, err := clientset.BatchV1().Jobs(namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", jobName),
	})
	if err != nil {
		return fmt.Errorf("failed to watch job: %v", err)
	}

	for {
		select {
		case event := <-watch.ResultChan():
			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				continue
			}

			if job.Status.Ready != nil && *job.Status.Ready > 0 {
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for job pod to become ready: %w", ctx.Err())
		}
	}
}

// WatchJobPodLogs streams logs from all pods associated with the given job.
func WatchJobPodLogs(ctx context.Context, clientset kubernetes.Interface, namespace, jobName string, out io.Writer) error {
	// First, get all pods for the job
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return fmt.Errorf("failed to list pods for job %s: %v", jobName, err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(pods.Items))

	for i := range pods.Items {
		go func(pod *corev1.Pod) {
			if err := streamLogs(ctx, clientset, namespace, pod, out, &wg); err != nil {
				fmt.Printf("failed to stream logs from pod %s: %v\n", pod.Name, err)
			}
		}(&pods.Items[i])
	}
	wg.Wait()

	return nil
}

func streamLogs(
	ctx context.Context,
	clientset kubernetes.Interface,
	namespace string,
	pod *corev1.Pod,
	out io.Writer,
	wg *sync.WaitGroup,
) error {
	defer wg.Done()
	fmt.Printf("Streaming logs from pod: %s\n", pod.Name)

	req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Follow: true,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("failed to stream logs from pod %s: %v", pod.Name, err)
	}
	defer func() {
		_ = stream.Close()
	}()

	// Stream logs to stdout
	_, err = io.Copy(out, stream)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to copy logs to stdout for pod %s: %v", pod.Name, err)
	}

	return nil
}
