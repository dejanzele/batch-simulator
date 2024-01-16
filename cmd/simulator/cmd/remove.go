package cmd

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/simulator"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Uninstall simulator components",
	Long: `This command facilitates the removal of key simulator components, ensuring a clean and orderly uninstallation process.
It executes two main actions:
1. uninstallation of the KWOK Operator
2. removal of the KWOK Stages.

These steps are crucial for reverting the simulation environment to its original state.`,
	Run: func(cmd *cobra.Command, args []string) {
		failed := false

		pterm.DefaultHeader.Println("uninstalling components...")

		// uninstall section
		blip()
		pterm.DefaultSection.Println("uninstall")
		pterm.Info.Println("uninstalling kwok operator...")
		output, err := simulator.UninstallOperator(cmd.Context(), config.KWOKNamespace)
		if err != nil {
			failed = true
			pterm.Error.Printf("failed to uninstall kwok operator: %v\n", err)
		}
		pterm.Println(string(output))

		pterm.Info.Println("uninstalling kwok stages...")
		output, err = simulator.DeleteStages(cmd.Context())
		if err != nil {
			failed = true
			pterm.Error.Printf("failed to uninstall kwok stages: %v\n", err)
		}
		pterm.Println(string(output))

		// status section
		blip()
		pterm.DefaultSection.Println("status")
		if failed {
			pterm.Error.Println("one or more components failed to uninstall")
		} else {
			pterm.Success.Println("all components uninstalled successfully")
		}
	},
}

func NewRemoveCmd() *cobra.Command {
	return removeCmd
}
