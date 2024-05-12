package cmd

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/dejanzele/batch-simulator/internal/simulator"
	"github.com/dejanzele/batch-simulator/internal/simulator/resources"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/k8s"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch simulation until all pods complete",
	Long:  `This command watches the simulation until all pods complete.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Print config section
		blip()
		printKWOKConfig()

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

		pterm.Info.Println("initializing kubernetes resource manager...")
		manager := k8s.NewManager(client, &k8s.ManagerConfig{Namespace: config.Namespace})
		pterm.Success.Println("kubernetes resource manager initialized successfully!")

		// watch section
		blip()
		pterm.DefaultSection.Println("watch")

		simulatorJobName := ""
		if len(args) > 0 {
			simulatorJobName = args[0]
		}

		simulationJob, err := getSimulatorJob(cmd.Context(), client, simulatorJobName)
		if err != nil {
			pterm.Error.Printf("failed to get simulator job: %v\n", err)
			os.Exit(1)
		}

		now := time.Now()
		if simulationJob == nil {
			pterm.Info.Println("no simulator job found, using now as start time")
		} else {
			pterm.Info.Printf(
				"found simulator job %s/%s created at %s\n",
				simulationJob.Namespace, simulationJob.Name, simulationJob.CreationTimestamp.String(),
			)
			now = simulationJob.CreationTimestamp.Time
		}

		pterm.Info.Println("waiting for simulation pods to complete...")
		if err := manager.WaitForPodsToComplete(cmd.Context(), resources.LabelSelectorFakePod, slog.Default()); err != nil {
			pterm.Error.Printf("failed to wait for pods to complete: %v", err)
		}

		// status section
		blip()
		pterm.DefaultSection.Println("status")
		end := time.Now()
		pterm.Info.Printf("simulation watch started at %s\n", now.String())
		pterm.Info.Printf("simulation watch ended at %s\n", end.String())
		pterm.Info.Printf("simulation watch duration: %s\n", time.Since(now).String())
	},
}

func getSimulatorJob(ctx context.Context, client kubernetes.Interface, jobName string) (simulatorJob *batchv1.Job, err error) {
	if jobName != "" {
		simulatorJob, err = getSimulatorJobByName(ctx, client, jobName)
		if err != nil {
			return nil, err
		}
	}
	if simulatorJob == nil {
		simulatorJob, err = findSimulatorJob(ctx, client)
	}
	return simulatorJob, err
}

func getSimulatorJobByName(ctx context.Context, client kubernetes.Interface, jobName string) (*batchv1.Job, error) {
	job, err := client.BatchV1().Jobs(config.Namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return job, nil
}

func findSimulatorJob(ctx context.Context, client kubernetes.Interface) (*batchv1.Job, error) {
	jobs, err := client.BatchV1().Jobs(config.Namespace).List(ctx, metav1.ListOptions{LabelSelector: simulator.LabelSelectorSimulator})
	if err != nil {
		return nil, err
	}
	if len(jobs.Items) == 0 {
		return nil, nil
	}
	return &jobs.Items[0], nil
}

func NewWatchCmd() *cobra.Command {
	watchCmd.Flags().StringVarP(&config.Namespace, "namespace", "n", config.Namespace, "namespace in which to create simulation resources")

	return watchCmd
}
