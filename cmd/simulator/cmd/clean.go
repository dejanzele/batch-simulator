package cmd

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/kubernetes"
	"github.com/dejanzele/batch-simulator/internal/kwok"
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
		client, err := kubernetes.NewClient(&config.Kubeconfig, cfg)
		if err != nil {
			pterm.Error.Printf("failed to initialize k8s client: %v", err)
			os.Exit(1)
		}
		pterm.Success.Println("kubernetes client initialized successfully!")

		pterm.Info.Println("initializing kubernetes resource manager...")
		manager := kubernetes.NewManager(client, kubernetes.ManagerConfig{})
		pterm.Success.Println("kubernetes resource manager initialized successfully!")

		// clean section
		blip()
		pterm.DefaultSection.Println("clean")
		wg := sync.WaitGroup{}
		wg.Add(2)

		// Create a multi printer for managing multiple printers
		multi := pterm.DefaultMultiPrinter
		_, _ = multi.Start()

		// wait for nodes & pods to fully terminate
		async := false
		go func() {
			defer wg.Done()
			spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("cleaning up nodes...")
			if err := manager.DeleteNodes(cmd.Context(), kwok.LabelSelector, async); err != nil {
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

		go func() {
			defer wg.Done()
			spinner, _ := pterm.DefaultSpinner.WithWriter(multi.NewWriter()).Start("cleaning up pods...")
			if err := manager.DeletePods(cmd.Context(), kwok.LabelSelector, async); err != nil {
				warning = true
				if errors.Is(err, context.DeadlineExceeded) {
					warning = true
					spinner.Warning("timed out waiting for all pods to terminate")
				} else {
					fatal = true
					spinner.Fail("failed to cleanup nodes")
					pterm.Error.Printf("%v", err)
				}
				return
			}
			spinner.Success("all pods fully terminated!")
		}()

		wg.Wait()

		// stop the pterm multi writer
		_, _ = multi.Stop()

		// status section
		blip()
		pterm.DefaultSection.Println("status")
		exitBasedOnStatus(fatal, warning)
	},
}

func NewCleanCmd() *cobra.Command {
	return cleanCmd
}
