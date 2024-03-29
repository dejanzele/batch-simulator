package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/k8s"
	"github.com/dejanzele/batch-simulator/internal/simulator"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install required simulator components",
	Long: `This command is responsible for setting up the essential components of the simulator.
It encompasses two key steps:
1. installing the KWOK Operator
2. installing the KWOK Stages which manage node and pod lifecycles

These installations are crucial for preparing the simulation environment,
ensuring all necessary functionalities are in place and operational.`,
	Run: func(cmd *cobra.Command, args []string) {
		failed := false

		pterm.DefaultHeader.Println("initializing components...")

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

		// install section
		blip()
		pterm.DefaultSection.Println("install")

		spinner, _ := pterm.DefaultSpinner.Start("installing kwok operator...")
		output, err := simulator.InstallOperator(cmd.Context(), config.KWOKNamespace)
		if err != nil {
			spinner.Fail("failed to install kwok operator")
			pterm.Error.Printf("%v\n", err)
			os.Exit(2)
		}
		spinner.Success("kwok operator installed successfully!")
		pterm.Println(string(output))

		spinner, _ = pterm.DefaultSpinner.Start("waiting for operator to become available...")
		output, running, err := simulator.CheckIsOperatorRunning(cmd.Context(), client, config.KWOKNamespace)
		if err != nil {
			spinner.Fail("failed to check if kwok operator is running")
			pterm.Error.Printf("%v\n", err)
			os.Exit(2)
		}
		if !running {
			failed = true
			spinner.Warning(string(output))
		} else {
			spinner.Success(string(output))
		}

		spinner, _ = pterm.DefaultSpinner.Start("installing kwok stages...")
		output, err = simulator.CreateStages(cmd.Context())
		if err != nil {
			failed = true
			spinner.Fail("failed to install kwok stages")
			pterm.Error.Printf("%v\n", err)
			os.Exit(2)
		}
		spinner.Success("kwok stages installed successfully!")
		pterm.Println(string(output))

		spinner, _ = pterm.DefaultSpinner.Start("checking are kwok stages created...")
		installed, missing, err := simulator.CheckAreStagesCreated(cmd.Context(), dynamicClient)
		if err != nil {
			failed = true
			spinner.Fail("failed to check if kwok stages are created")
			pterm.Error.Printf("%v\n", err)
		}
		if !installed {
			failed = true
			spinner.Warning("stages not created: %v\n", missing)
		} else {
			spinner.Success("kwok stages created successfully!")
		}

		spinner, _ = pterm.DefaultSpinner.Start("installing necessary RBAC resources...")
		if err := simulator.CreateRBAC(cmd.Context(), client, config.Namespace); err != nil {
			failed = true
			spinner.Fail("failed to install RBAC resources")
			pterm.Error.Printf("%v\n", err)
		} else {
			spinner.Success("RBAC resources installed successfully!")
		}

		// status section
		blip()
		pterm.DefaultSection.Println("status")
		if failed {
			pterm.Warning.Println("one or more components failed to install")
		} else {
			pterm.Success.Println("all components installed successfully")
		}
	},
}

func NewInitCmd() *cobra.Command {
	addKubeconfigFlag(installCmd)
	addKubernetesConfigFlags(installCmd)
	return installCmd
}
