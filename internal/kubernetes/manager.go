package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	batchv1 "k8s.io/api/batch/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/kwok/resources"
	"github.com/dejanzele/batch-simulator/internal/ratelimiter"
	"github.com/dejanzele/batch-simulator/internal/ratelimiter/executor"
)

const (
	defaultNamespace                      = "default"
	defaultPodRateLimiterFrequency        = 1 * time.Second
	defaultPodRateLimiterRequests   int32 = 10
	defaultNodeRateLimiterFrequency       = 1 * time.Second
	defaultNodeRateLimiterRequests  int32 = 5
	defaultJobRateLimiterFrequency        = 1 * time.Second
	defaultJobRateLimiterRequests   int32 = 5
)

// Manager is used to manage Kubernetes resources.
type Manager struct {
	// client is the Kubernetes client that should be used by the Manager.
	client kubernetes.Interface
	// namespace is the namespace in which resources should be created.
	// If no namespace is provided, the default namespace will be used.
	namespace string
	// logger is the logger that should be used by the Manager.
	// If no logger is provided, a new logger will be created.
	logger *slog.Logger
	// rateLimitedPodCreator is the rate limiter that should be used for Pod resources.
	rateLimitedPodCreator *ratelimiter.RateLimiter[*corev1.Pod]
	// rateLimitedNodeCreator is the rate limiter that should be used for Node resources.
	rateLimitedNodeCreator *ratelimiter.RateLimiter[*corev1.Node]
	// rateLimitedJobCreator is the rate limiter that should be used for Job resources.
	rateLimitedJobCreator *ratelimiter.RateLimiter[*batchv1.Job]
}

// ManagerConfig is used to configure a new Manager.
type ManagerConfig struct {
	// Namespace is the namespace in which resources should be created.
	Namespace string
	// Logger is the logger that should be used by the Manager.
	Logger *slog.Logger
	// PodRateLimiterConfig is the configuration for the rate limited PodCreator.
	PodRateLimiterConfig RateLimiterConfig
	// NodeRateLimiterConfig is the configuration for the rate limited NodeCreator.
	NodeRateLimiterConfig RateLimiterConfig
	// JobRateLimiterConfig is the configuration for the rate limited JobCreator.
	JobRateLimiterConfig RateLimiterConfig
}

// RateLimiterConfig is used to configure the rate limiter for a specific resource type.
type RateLimiterConfig struct {
	// Frequency is the frequency at which the rate limiter should be invoked.
	Frequency time.Duration
	// Requests is the number of requests that should be made per invocation.
	Requests int32
	// Limit is the maximum number of items that should be processed
	Limit int32
}

func NewManager(client kubernetes.Interface, cfg ManagerConfig) *Manager {
	defaultedConfig := defaultManagerConfig(cfg)
	nodeExecutor := executor.NewNodeCreator(client)
	nodeRateLimiter := ratelimiter.New[*corev1.Node](
		defaultedConfig.NodeRateLimiterConfig.Frequency,
		defaultedConfig.NodeRateLimiterConfig.Requests,
		nodeExecutor,
		ratelimiter.WithLimit[*corev1.Node](defaultedConfig.NodeRateLimiterConfig.Limit),
	)
	podExecutor := executor.NewPodCreator(client, defaultedConfig.Namespace)
	podRateLimiter := ratelimiter.New[*corev1.Pod](
		defaultedConfig.PodRateLimiterConfig.Frequency,
		defaultedConfig.PodRateLimiterConfig.Requests,
		podExecutor,
		ratelimiter.WithLimit[*corev1.Pod](defaultedConfig.PodRateLimiterConfig.Limit),
	)
	jobExecutor := executor.NewJobCreator(client, defaultedConfig.Namespace)
	jobRateLimiter := ratelimiter.New[*batchv1.Job](
		defaultedConfig.JobRateLimiterConfig.Frequency,
		defaultedConfig.JobRateLimiterConfig.Requests,
		jobExecutor,
		ratelimiter.WithLimit[*batchv1.Job](defaultedConfig.JobRateLimiterConfig.Limit),
	)
	m := &Manager{
		client:                 client,
		namespace:              defaultedConfig.Namespace,
		logger:                 defaultedConfig.Logger,
		rateLimitedNodeCreator: nodeRateLimiter,
		rateLimitedPodCreator:  podRateLimiter,
		rateLimitedJobCreator:  jobRateLimiter,
	}
	m.logger = slog.With("process", "manager")
	if defaultedConfig.NodeRateLimiterConfig.Limit > 0 {
		_ = m.SubmitNodes(defaultedConfig.NodeRateLimiterConfig.Limit)
	}
	if defaultedConfig.PodRateLimiterConfig.Limit > 0 {
		_ = m.SubmitPods(defaultedConfig.PodRateLimiterConfig.Limit)
	}
	if defaultedConfig.JobRateLimiterConfig.Limit > 0 {
		_ = m.SubmitJobs(defaultedConfig.JobRateLimiterConfig.Limit)
	}
	return m
}

// defaultManagerConfig returns a new ManagerConfig with default values set.
func defaultManagerConfig(cfg ManagerConfig) ManagerConfig {
	if cfg.Namespace == "" {
		cfg.Namespace = defaultNamespace
	}
	if cfg.Logger == nil {
		cfg.Logger = &slog.Logger{}
	}
	if cfg.PodRateLimiterConfig.Frequency == 0 {
		cfg.PodRateLimiterConfig.Frequency = defaultPodRateLimiterFrequency
	}
	if cfg.PodRateLimiterConfig.Requests == 0 {
		cfg.PodRateLimiterConfig.Requests = defaultPodRateLimiterRequests
	}
	if cfg.NodeRateLimiterConfig.Frequency == 0 {
		cfg.NodeRateLimiterConfig.Frequency = defaultNodeRateLimiterFrequency
	}
	if cfg.NodeRateLimiterConfig.Requests == 0 {
		cfg.NodeRateLimiterConfig.Requests = defaultNodeRateLimiterRequests
	}
	if cfg.JobRateLimiterConfig.Frequency == 0 {
		cfg.JobRateLimiterConfig.Frequency = defaultJobRateLimiterFrequency
	}
	if cfg.JobRateLimiterConfig.Requests == 0 {
		cfg.JobRateLimiterConfig.Requests = defaultJobRateLimiterRequests
	}
	return cfg
}

// Start starts the Manager and the pod & node creation rate limiters.
// It blocks until the Manager is stopped the context is cancelled or all rate limited executors have finished.
func (m *Manager) Start(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	m.logger.Info("starting kubernetes resource manager with rate limiting")
	go m.rateLimitedNodeCreator.Run(ctx)
	go m.rateLimitedPodCreator.Run(ctx)
	go m.rateLimitedJobCreator.Run(ctx)

	for {
		select {
		case err := <-m.rateLimitedNodeCreator.ErrChan():
			m.logger.Error("received error from node rate limiter", "error", err)
		case err := <-m.rateLimitedPodCreator.ErrChan():
			m.logger.Error("received error from pod rate limiter", "error", err)
		case err := <-m.rateLimitedJobCreator.ErrChan():
			m.logger.Error("received error from job rate limiter", "error", err)
		case <-ctx.Done():
			m.Stop()
			return ctx.Err()
		case <-ticker.C:
			nodeCreatorStopped := !m.rateLimitedNodeCreator.IsRunning()
			podCreatorStopped := !m.rateLimitedPodCreator.IsRunning()
			jobCreatorStopped := !m.rateLimitedJobCreator.IsRunning()
			if nodeCreatorStopped && podCreatorStopped && jobCreatorStopped {
				m.Stop()
				return nil
			}
		}
	}
}

// Stop stops the Manager.
func (m *Manager) Stop() {
	m.logger.Info("stopping kubernetes resource manager")
	m.rateLimitedPodCreator.Stop()
	m.rateLimitedNodeCreator.Stop()
	m.rateLimitedJobCreator.Stop()
}

// SubmitNodes adds the specified number of Kubernetes Node resources to the rate-limited work queue.
func (m *Manager) SubmitNodes(count int32) error {
	nodes := make([]*corev1.Node, 0, count)
	for i := 0; i < int(count); i++ {
		nodeName := fmt.Sprintf("fake-node-%d", i)
		nodes = append(nodes, resources.NewFakeNode(nodeName))
	}
	return m.rateLimitedNodeCreator.AddWorkItems(nodes...)
}

// DeleteNodes deletes all Kubernetes Node resources having provided label.
// If async is set to false, this function will block until nodes are terminated or context exceeds deadline.
func (m *Manager) DeleteNodes(ctx context.Context, labelSelector string, async bool) error {
	deleteFunc := func(ctx context.Context, client kubernetes.Interface, deleteOpts metav1.DeleteOptions, listOpts metav1.ListOptions, async bool) error {
		return client.CoreV1().Nodes().DeleteCollection(ctx, deleteOpts, listOpts)
	}
	listFunc := func(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) (bool, error) {
		nodeList, err := client.CoreV1().Nodes().List(ctx, opts)
		if err != nil {
			return false, err
		}
		return len(nodeList.Items) == 0, nil
	}
	m.logger.Info("deleting nodes", "labelSelector", labelSelector, "async", async)
	if err := deleteCollection(ctx, m.client, labelSelector, deleteFunc, listFunc, async); err != nil {
		return fmt.Errorf("failed to delete nodes with labelSelector=%s: %w", labelSelector, err)
	}

	return nil
}

// WaitForNodesToTerminate waits for the nodes with the provided labelSelector to terminate.
func (m *Manager) WaitForNodesToTerminate(ctx context.Context, client kubernetes.Interface, labelSelector string) error {
	listFunc := func(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) (bool, error) {
		nodeList, err := client.CoreV1().Nodes().List(ctx, opts)
		if err != nil {
			return false, err
		}
		return len(nodeList.Items) == 0, nil
	}
	return waitFor(ctx, client, labelSelector, listFunc)
}

// SubmitPods adds the specified number of Kubernetes Pod resources to the rate-limited work queue.
func (m *Manager) SubmitPods(count int32) error {
	pods := make([]*corev1.Pod, 0, count)
	for i := 0; i < int(count); i++ {
		podName := fmt.Sprintf("fake-pod-%d", i)
		pod := resources.NewFakePod(podName, m.namespace)
		pods = append(pods, pod)
	}
	return m.rateLimitedPodCreator.AddWorkItems(pods...)
}

// DeletePods deletes Kubernetes Pod resources having provided label.
// If async is set to false, this function will block until pods are terminated or context exceeds deadline.
func (m *Manager) DeletePods(ctx context.Context, labelSelector string, async bool) error {
	deleteFunc := func(ctx context.Context, client kubernetes.Interface, deleteOpts metav1.DeleteOptions, listOpts metav1.ListOptions, async bool) error {
		return client.CoreV1().Pods(m.namespace).DeleteCollection(ctx, deleteOpts, listOpts)
	}
	listFunc := func(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) (bool, error) {
		podList, err := client.CoreV1().Pods(m.namespace).List(ctx, opts)
		if err != nil {
			return false, err
		}
		return len(podList.Items) == 0, nil
	}
	m.logger.Info("deleting pods", "labelSelector", labelSelector, "async", async)
	if err := deleteCollection(ctx, m.client, labelSelector, deleteFunc, listFunc, async); err != nil {
		return fmt.Errorf("failed to delete pods with labelSelector=%s: %w", labelSelector, err)
	}

	return nil
}

// WaitForPodsToTerminate waits for the pods with the provided labelSelector to terminate.
func (m *Manager) WaitForPodsToTerminate(ctx context.Context, client kubernetes.Interface, labelSelector string) error {
	listFunc := func(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) (bool, error) {
		podList, err := client.CoreV1().Pods(m.namespace).List(ctx, opts)
		if err != nil {
			return false, err
		}
		return len(podList.Items) == 0, nil
	}
	return waitFor(ctx, client, labelSelector, listFunc)
}

// SubmitJobs adds the specified number of Kubernetes Job resources to the rate-limited work queue.
func (m *Manager) SubmitJobs(count int32) error {
	jobs := make([]*batchv1.Job, 0, count)
	for i := 0; i < int(count); i++ {
		jobName := fmt.Sprintf("fake-job-%d", i)
		job := resources.NewFakeJob(jobName, m.namespace)
		jobs = append(jobs, job)
	}
	return m.rateLimitedJobCreator.AddWorkItems(jobs...)
}

// DeleteJobs deletes Kubernetes Job resources having provided label.
// If async is set to false, this function will block until jobs are terminated or context exceeds deadline.
func (m *Manager) DeleteJobs(ctx context.Context, labelSelector string, async bool) error {
	deleteFunc := func(ctx context.Context, client kubernetes.Interface, deleteOpts metav1.DeleteOptions, listOpts metav1.ListOptions, async bool) error {
		deletePropagationBackground := metav1.DeletePropagationBackground
		deleteOpts.PropagationPolicy = &deletePropagationBackground
		return client.BatchV1().Jobs(m.namespace).DeleteCollection(ctx, deleteOpts, listOpts)
	}
	listFunc := func(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) (bool, error) {
		jobList, err := client.BatchV1().Jobs(m.namespace).List(ctx, opts)
		if err != nil {
			return false, err
		}
		return len(jobList.Items) == 0, nil
	}
	m.logger.Info("deleting jobs", "labelSelector", labelSelector, "async", async)
	if err := deleteCollection(ctx, m.client, labelSelector, deleteFunc, listFunc, async); err != nil {
		return fmt.Errorf("failed to delete jobs with labelSelector=%s: %w", labelSelector, err)
	}

	return nil
}

// WaitForJobsToTerminate waits for the jobs with the provided labelSelector to terminate.
func (m *Manager) WaitForJobsToTerminate(ctx context.Context, client kubernetes.Interface, labelSelector string) error {
	listFunc := func(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) (bool, error) {
		jobList, err := client.BatchV1().Jobs(m.namespace).List(ctx, opts)
		if err != nil {
			return false, err
		}
		return len(jobList.Items) == 0, nil
	}
	return waitFor(ctx, client, labelSelector, listFunc)
}

// deleteCollection deletes Kubernetes resources having provided label.
func deleteCollection(ctx context.Context, client kubernetes.Interface, labelSelector string, deleteFunc DeleteFunc, listFunc ListFunc, async bool) error {
	deleteOpts := metav1.DeleteOptions{}
	listOpts := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	if err := deleteFunc(ctx, client, deleteOpts, listOpts, async); err != nil {
		return err
	}
	if !async {
		if err := waitFor(ctx, client, labelSelector, listFunc); err != nil {
			return err
		}
	}
	return nil
}

type DeleteFunc func(ctx context.Context, client kubernetes.Interface, deleteOpts metav1.DeleteOptions, listOpts metav1.ListOptions, async bool) error
type ListFunc func(ctx context.Context, client kubernetes.Interface, opts metav1.ListOptions) (empty bool, err error)

// waitFor waits for the resources with the provided labelSelector to be empty.
func waitFor(ctx context.Context, client kubernetes.Interface, labelSelector string, listFunc ListFunc) error {
	return wait.PollUntilContextTimeout(
		ctx,
		config.DefaultPollInterval,
		2*config.DefaultPollTimeout,
		false,
		func(ctx context.Context) (done bool, err error) {
			listOpts := metav1.ListOptions{LabelSelector: labelSelector}
			empty, err := listFunc(ctx, client, listOpts)
			if err != nil {
				return false, err
			}

			return empty, nil
		},
	)
}

func (m *Manager) Metrics() (nodeCreationMetrics, podCreationMetrics, jobCreationMetrics ratelimiter.Metrics) {
	nodeCreationMetrics = m.rateLimitedNodeCreator.Metrics()
	podCreationMetrics = m.rateLimitedPodCreator.Metrics()
	jobCreationMetrics = m.rateLimitedJobCreator.Metrics()
	return
}
