package cmd

import (
	"os"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/logger"

	"github.com/spf13/cobra"
)

// rootCmd represents the root command
var rootCmd = &cobra.Command{
	Use:   "sim",
	Short: "kwok-based batch simulation tool",
	Long: `This command-line interface (CLI) tool facilitates the simulation of batch scheduling scenarios,
leveraging Kubernetes (k8s) and Kwok technologies.
It's designed for users who need to model and understand various batch processing workflows within a k8s environment.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Init(getLogLevel())
	},
	Run: func(cmd *cobra.Command, args []string) {
		if Help(cmd, args) {
			os.Exit(0)
		}
	},
}

func NewRootCmd() *cobra.Command {
	rootCmd.PersistentFlags().BoolVarP(&config.Verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "d", false, "enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&config.Silent, "silent", "s", false, "disable internal logging")
	rootCmd.PersistentFlags().BoolVar(&config.NoGUI, "no-gui", false, "disable printing graphical elements")
	rootCmd.AddCommand(NewCheckCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewRemoveCmd())
	rootCmd.AddCommand(NewRunCmd())
	rootCmd.AddCommand(NewCleanCmd())
	rootCmd.MarkFlagsMutuallyExclusive("verbose", "debug", "silent")
	return rootCmd
}
