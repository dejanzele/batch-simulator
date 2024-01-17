package cmd

import (
	"os"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/k8s"
	"github.com/dejanzele/batch-simulator/internal/simulator"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check are required components installed & configured",
	Long: `This command conducts comprehensive checks for essential components necessary for the system's operation,
including the presence of 'kubectl', 'kwok', and various stages.
It ensures that all required tools and configurations are in place and functioning correctly,
offering a quick and efficient way to validate the setup.`,
	Run: func(cmd *cobra.Command, args []string) {
		fatal := false
		warning := false

		pterm.DefaultHeader.Println("starting checks...")

		// config section
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
		dynamicClient, err := k8s.NewDynamicClient(&config.Kubeconfig, cfg)
		if err != nil {
			pterm.Error.Printf("failed to initialize dynamic k8s client: %v", err)
			os.Exit(1)
		}
		pterm.Success.Println("kubernetes client initialized successfully!")

		// checks section
		blip()
		pterm.DefaultSection.Println("checks")
		spinner, _ := pterm.DefaultSpinner.Start("is kubectl installed?")
		time.Sleep(500 * time.Millisecond)
		_, ok := simulator.CheckIsKubectlInstalled(cmd.Context())
		if !ok {
			warning = true
			spinner.Warning("kubectl is not installed")
		} else {
			spinner.Success("kubectl is installed")
		}

		time.Sleep(500 * time.Millisecond)

		spinner, _ = pterm.DefaultSpinner.Start("is kwok cli installed?")
		time.Sleep(500 * time.Millisecond)
		_, ok = simulator.CheckIsKWOKInstalled(cmd.Context())
		if !ok {
			warning = true
			spinner.Warning("kwok cli is not installed")
		} else {
			spinner.Success("kwok is installed")
		}

		time.Sleep(500 * time.Millisecond)

		spinner, _ = pterm.DefaultSpinner.Start("are stages created?")
		time.Sleep(500 * time.Millisecond)
		created, missing, err := simulator.CheckAreStagesCreated(cmd.Context(), dynamicClient)
		if err != nil {
			fatal = true
			spinner.Fail("failed to check if stages are created")
			pterm.Error.Printf("%v\n", err)
		}
		if !created {
			warning = true
			spinner.Warning("required kwok stages are not installed! run 'simulator init' to install required components.")
			pterm.Warning.Printf("following stages are missing: %v\n", missing)
		} else {
			spinner.Success("all stages are installed")
		}

		time.Sleep(500 * time.Millisecond)

		spinner, _ = pterm.DefaultSpinner.Start("is kwok-operator running?")
		time.Sleep(500 * time.Millisecond)
		_, running, err := simulator.CheckIsOperatorRunning(cmd.Context(), client, config.KWOKNamespace)
		if err != nil {
			fatal = true
			spinner.Fail("failed to check is kwok-operator running")
			pterm.Error.Printf("%v\n", err)
		}
		if !running {
			warning = true
			spinner.Warning("kwok-operator is not running")
		} else {
			spinner.Success("kwok-operator is running")
		}

		// status section
		blip()
		pterm.DefaultSection.Println("status")
		if warning {
			pterm.Warning.Println("run 'simulator install' to install required components")
		}
		exitBasedOnStatus(fatal, warning)
	},
}

func NewCheckCmd() *cobra.Command {
	addKubeconfigFlag(checkCmd)
	addKubernetesConfigFlags(checkCmd)
	checkCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "verbose output")
	return checkCmd
}
