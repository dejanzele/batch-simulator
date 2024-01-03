package cmd

import (
	"os"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/kubernetes"
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
		client, err := kubernetes.NewClient(&config.Kubeconfig, cfg)
		if err != nil {
			pterm.Error.Printf("failed to initialize k8s client: %v", err)
			os.Exit(1)
		}
		pterm.Success.Println("kubernetes client initialized successfully!")

		pterm.Info.Println("initializing kubernetes resource manager...")
		managerConfig := kubernetes.ManagerConfig{
			Namespace: config.Namespace,
			PodRateLimiterConfig: kubernetes.RateLimiterConfig{
				Frequency: config.PodCreatorFrequency,
				Requests:  config.PodCreatorRequests,
				Limit:     config.PodCreatorLimit,
			},
			NodeRateLimiterConfig: kubernetes.RateLimiterConfig{
				Frequency: config.NodeCreatorFrequency,
				Requests:  config.NodeCreatorRequests,
				Limit:     config.NodeCreatorLimit,
			},
			JobRateLimiterConfig: kubernetes.RateLimiterConfig{
				Frequency: config.JobCreatorFrequency,
				Requests:  config.JobCreatorRequests,
				Limit:     config.JobCreatorLimit,
			},
		}
		manager := kubernetes.NewManager(client, managerConfig)
		pterm.Success.Println("kubernetes resource manager initialized successfully!")

		// run section
		blip()
		pterm.DefaultSection.Println("run")
		var callback func()
		wg := sync.WaitGroup{}
		if !config.NoGUI {
			wg.Add(1)
			callback = func() { wg.Done() }
			go printMetricsEvery(cmd.Context(), 1*time.Second, manager, callback)
		}
		_ = manager.Start(cmd.Context())
		wg.Wait()

		// status section
		blip()
		pterm.DefaultSection.Println("status")
		pterm.Success.Println("simulator finished successfully!")
	},
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

	return runCmd
}
