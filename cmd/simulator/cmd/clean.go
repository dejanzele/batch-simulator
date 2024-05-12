package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"

	"k8s.io/utils/strings/slices"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/k8s"
	"github.com/dejanzele/batch-simulator/internal/simulator"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean deletes all resources (nodes, pods...) created by the simulator",
	Long: `This command removes all pods and nodes generated during simulations.
It ensures a thorough cleanup by initializing Kubernetes clients, managing resources effectively,
and deleting nodes and pods tagged for simulation.
It's a comprehensive approach to maintaining a clean and efficient simulation environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fatal := false
		warning := false

		pterm.DefaultHeader.Println("cleaning up simulation data...")

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

		// clean section
		blip()
		pterm.DefaultSection.Println("clean")

		resourceCount := len(config.Resources)
		if resourceCount == 0 {
			config.Resources = []string{"nodes", "pods", "jobs"}
			resourceCount = 3
		}

		pterm.Printf("cleaning up following resources: %v\n", config.Resources)

		wg := sync.WaitGroup{}
		wg.Add(resourceCount)

		// Create a multi printer for managing multiple printers
		multi := pterm.DefaultMultiPrinter
		_, _ = multi.Start()

		// wait for nodes & pods to fully terminate
		var errorList []error
		async := false

		if slices.Contains(config.Resources, "nodes") || slices.Contains(config.Resources, "node") {
			go func() {
				defer wg.Done()
				spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("cleaning up nodes...")
				if err := manager.DeleteNodes(cmd.Context(), simulator.LabelSelector, async); err != nil {
					errorList = append(errorList, err)
					if errors.Is(err, context.DeadlineExceeded) {
						warning = true
						spinner.Warning("timed out waiting for all nodes to terminate")
					} else {
						fatal = true
						spinner.Fail("failed to cleanup nodes")
						pterm.Error.Printf("%v", err)
					}
					return
				}
				spinner.Success("all nodes fully terminated!")
			}()
		}

		if slices.Contains(config.Resources, "pods") || slices.Contains(config.Resources, "pod") {
			go func() {
				defer wg.Done()
				spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("cleaning up pods...")
				if err := manager.DeletePods(cmd.Context(), simulator.LabelSelector, async); err != nil {
					errorList = append(errorList, err)
					if errors.Is(err, context.DeadlineExceeded) {
						warning = true
						spinner.Warning("timed out waiting for all pods to terminate")
					} else {
						fatal = true
						spinner.Fail("failed to cleanup pods")
					}
					return
				}
				spinner.Success("all pods fully terminated!")
			}()
		}

		if slices.Contains(config.Resources, "jobs") || slices.Contains(config.Resources, "job") {
			go func() {
				defer wg.Done()
				spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("cleaning up jobs...")
				if err := manager.DeleteJobs(cmd.Context(), simulator.LabelSelector, async); err != nil {
					errorList = append(errorList, err)
					if errors.Is(err, context.DeadlineExceeded) {
						warning = true
						spinner.Warning("timed out waiting for all jobs to terminate")
					} else {
						fatal = true
						spinner.Fail("failed to cleanup jobs")
						pterm.Error.Printf("%v", err)
					}
					return
				}
				spinner.Success("all jobs fully terminated!")
			}()
		}

		wg.Wait()

		// stop the pterm multi writer
		_, _ = multi.Stop()

		// status section
		blip()
		pterm.DefaultSection.Println("status")
		if len(errorList) > 0 {
			for _, err := range errorList {
				pterm.Error.Printf("%v\n", err)
			}
		}
		exitBasedOnStatus(fatal, warning)
	},
}

func NewCleanCmd() *cobra.Command {
	cleanCmd.Flags().StringVarP(&config.Namespace, "namespace", "n", config.Namespace, "namespace in which to create simulation resources")
	cleanCmd.Flags().StringSliceVarP(&config.Resources, "resources", "r", config.Resources, "resources to delete (nodes, pods, jobs)")

	validate := func() {
		for _, r := range config.Resources {
			switch r {
			case "nodes", "node", "pods", "pod", "jobs", "job":
				continue
			default:
				slog.Error("unsupported resource type:" + r + ", --resources|-r supports only node(s),job(s),pod(s)")
				os.Exit(1)
			}
		}
	}
	cobra.OnInitialize(validate)

	return cleanCmd
}
