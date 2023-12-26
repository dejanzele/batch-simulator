package cmd

import (
	"os"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	"github.com/dejanzele/batch-simulator/cmd/simulator/config"
	"github.com/dejanzele/batch-simulator/internal/kubernetes"
)

// Help is a helper function to print command help if 0 arguments passed and command requires an argument.
func Help(cmd *cobra.Command, args []string) bool {
	if len(args) == 0 {
		_ = cmd.Help()
		return true
	}
	return false
}

// addKubeconfigFlag adds the --kubeconfig flag to the given command.
func addKubeconfigFlag(cmd *cobra.Command) {
	if home := homedir.HomeDir(); home != "" {
		cmd.PersistentFlags().StringVarP(&config.Kubeconfig, "kubeconfig", "k", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		cmd.PersistentFlags().StringVarP(&config.Kubeconfig, "kubeconfig", "k", "", "absolute path to the kubeconfig file")
	}
}

// addKubernetesConfigFlags adds flags for configuring QPS & Burst in the Kubernetes client.
func addKubernetesConfigFlags(cmd *cobra.Command) {
	cmd.Flags().Float32Var(&config.QPS, "kube-api-qps", config.QPS, "Maximum QPS to use while talking with Kubernetes API")
	cmd.Flags().IntVar(&config.Burst, "kube-api-burst", config.Burst, "Maximum burst for throttle while talking with Kubernetes API")
}

// getKubernetesConfig returns a kubernetes.Config based on the current configuration.
func getKubernetesConfig() kubernetes.Config {
	return kubernetes.Config{
		QPS:   config.QPS,
		Burst: config.Burst,
	}
}

// exitBasedOnStatus prints a message and exits with the appropriate exit code based on the given flags.
func exitBasedOnStatus(fatal, warning bool) {
	switch {
	case fatal:
		pterm.Error.Println("one or more checks encountered fatal errors")
		os.Exit(2)
	case warning:
		pterm.Warning.Println("one or more checks failed")
		os.Exit(1)
	default:
		pterm.Success.Println("all checks passed")
		os.Exit(0)
	}
}
