package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
var Version = "dev"

// rootCmd is the top-level Cobra command for azfloci.
var rootCmd = &cobra.Command{
	Use:   "azfloci",
	Short: "Azure resource orchestration CLI",
	Long:  "azfloci wraps the Azure CLI to provide workflow-driven resource management for Azure.",
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(authCmd)
}

// versionCmd prints the azfloci version string.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("azfloci %s\n", Version)
	},
}

func Execute() error {
	return rootCmd.Execute()
}

// RootCmd exposes the root command for testing.
func RootCmd() *cobra.Command {
	return rootCmd
}
