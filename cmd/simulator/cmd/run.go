package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/dejanzele/batch-simulator/internal/simulator"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/k8s"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a simulation",
	Long: `This command initiates the simulation process, creating simulated nodes and pods at a user-configurable rate.
It involves initializing Kubernetes clients, setting up a resource manager with specified rate limits for pods and nodes,
and managing the simulation lifecycle.
The process is designed to mimic real-world Kubernetes environments for testing and analysis purposes.`,
	Run: func(cmd *cobra.Command, args []string) {
		pterm.DefaultHeader.Println("running simulation...")

		// Print config section
		blip()
		printSimulationConfig()

		// init section
		blip()
		pterm.DefaultSection.Println("init")

		pterm.Info.Println("initializing kubernetes clients...")
		cfg := getKubernetesConfig()
		client, err := k8s.NewClient(&config.Kubeconfig, cfg)
		if err != nil {
			pterm.Error.Printf("failed to initialize k8s client: %v", err)
			os.Exit(1)
		}
		pterm.Success.Println("kubernetes client initialized successfully!")

		if config.Remote {
			pterm.Success.Println("running simulation in remote Kubernetes cluster")
			err = runRemote(cmd.Context(), client)
		} else {
			pterm.Success.Println("running simulation from local machine")
			err = runLocal(cmd.Context(), client)
		}
		if err != nil {
			pterm.Error.Printf("failed to run simulation: %v", err)
			os.Exit(1)
		}
		// status section
		blip()
		pterm.DefaultSection.Println("status")
		pterm.Success.Println("simulator finished successfully!")
	},
}

func runRemote(ctx context.Context, client kubernetes.Interface) error {
	args := []string{
		"--node-creator-frequency", config.NodeCreatorFrequency.String(),
		"--node-creator-requests", fmt.Sprintf("%d", config.NodeCreatorRequests),
		"--node-creator-limit", fmt.Sprintf("%d", config.NodeCreatorLimit),
		"--pod-creator-frequency", config.PodCreatorFrequency.String(),
		"--pod-creator-requests", fmt.Sprintf("%d", config.PodCreatorRequests),
		"--pod-creator-limit", fmt.Sprintf("%d", config.PodCreatorLimit),
		"--job-creator-frequency", config.JobCreatorFrequency.String(),
		"--job-creator-requests", fmt.Sprintf("%d", config.JobCreatorRequests),
		"--job-creator-limit", fmt.Sprintf("%d", config.JobCreatorLimit),
		"--namespace", config.Namespace,
		"--no-gui",
		"--verbose",
	}
	job := simulator.NewSimulatorJob(args)
	_, err := client.BatchV1().Jobs(config.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create simulator job: %v", err)
	}

	pterm.Info.Println("waiting for simulator job pod to become ready...")
	if err := k8s.WaitForJobPodsReady(ctx, client, config.Namespace, job.Name, config.DefaultPollTimeout); err != nil {
		return fmt.Errorf("failed to wait for simulator job pods to become ready: %v", err)
	}

	pterm.Info.Println("streaming simulator job pod logs...")
	if err := k8s.WatchJobPodLogs(ctx, client, config.Namespace, job.Name, os.Stdout); err != nil {
		return fmt.Errorf("failed to watch simulator job pod logs: %v", err)
	}

	return nil
}

func runLocal(ctx context.Context, client kubernetes.Interface) error {
	pterm.Success.Println("kubernetes client initialized successfully!")

	pterm.Info.Println("initializing kubernetes resource manager...")
	managerConfig := k8s.ManagerConfig{
		Namespace: config.Namespace,
		PodRateLimiterConfig: k8s.RateLimiterConfig{
			Frequency: config.PodCreatorFrequency,
			Requests:  config.PodCreatorRequests,
			Limit:     config.PodCreatorLimit,
		},
		NodeRateLimiterConfig: k8s.RateLimiterConfig{
			Frequency: config.NodeCreatorFrequency,
			Requests:  config.NodeCreatorRequests,
			Limit:     config.NodeCreatorLimit,
		},
		JobRateLimiterConfig: k8s.RateLimiterConfig{
			Frequency: config.JobCreatorFrequency,
			Requests:  config.JobCreatorRequests,
			Limit:     config.JobCreatorLimit,
		},
	}
	manager := k8s.NewManager(client, &managerConfig)
	pterm.Success.Println("kubernetes resource manager initialized successfully!")

	// run section
	blip()
	pterm.DefaultSection.Println("run")
	var callback func()
	wg := sync.WaitGroup{}
	if !config.NoGUI {
		wg.Add(1)
		callback = func() { wg.Done() }
		go printMetricsEvery(ctx, 1*time.Second, manager, callback)
	}
	_ = manager.Start(ctx)
	wg.Wait()

	return nil
}

func NewRunCmd() *cobra.Command {
	runCmd.Flags().DurationVar(&config.NodeCreatorFrequency, "node-creator-frequency", config.NodeCreatorFrequency, "frequency at which to create nodes")
	runCmd.Flags().IntVar(&config.NodeCreatorRequests, "node-creator-requests", config.NodeCreatorRequests, "number of node creation requests to make in each iteration")
	runCmd.Flags().IntVar(&config.NodeCreatorLimit, "node-creator-limit", config.NodeCreatorLimit, "maximum number of nodes to create")
	runCmd.Flags().DurationVar(&config.PodCreatorFrequency, "pod-creator-frequency", config.PodCreatorFrequency, "Frequency at which to create pods")
	runCmd.Flags().IntVar(&config.PodCreatorRequests, "pod-creator-requests", config.PodCreatorRequests, "number of pod creation requests to make in each iteration")
	runCmd.Flags().IntVar(&config.PodCreatorLimit, "pod-creator-limit", config.PodCreatorLimit, "maximum number of pods to create")
	runCmd.Flags().DurationVar(&config.JobCreatorFrequency, "job-creator-frequency", config.JobCreatorFrequency, "frequency at which to create jobs")
	runCmd.Flags().IntVar(&config.JobCreatorRequests, "job-creator-requests", config.JobCreatorRequests, "number of job creation requests to make in each iteration")
	runCmd.Flags().IntVar(&config.JobCreatorLimit, "job-creator-limit", config.JobCreatorLimit, "maximum number of jobs to create")
	runCmd.Flags().StringVarP(&config.Namespace, "namespace", "n", config.Namespace, "namespace in which to create simulation resources")
	runCmd.Flags().BoolVarP(&config.Remote, "remote", "r", config.Remote, "run the simulator in a Kubernetes cluster")
	runCmd.Flags().IntVar(&config.PodSpecSize, "pod-spec-size", config.PodSpecSize, "size of the pod spec in bytes")

	return runCmd
}
