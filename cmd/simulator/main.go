package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra/doc"

	"github.com/dejanzele/batch-simulator/cmd/simulator/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if len(os.Args) > 1 && os.Args[1] == "docgen" {
		err := doc.GenMarkdownTree(rootCmd, "docs")
		if err != nil {
			slog.Error("failed to generate docs", "error", err)
			os.Exit(3)
		}
		os.Exit(0)
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(2)
	}
}
