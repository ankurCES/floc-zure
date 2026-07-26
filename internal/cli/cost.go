package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ankurCES/floc-zure/internal/cost"
	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Cost estimation commands",
}

var costEstimateCmd = &cobra.Command{
	Use:   "estimate",
	Short: "Estimate monthly costs for resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		format, _ := cmd.Flags().GetString("format")
		if file == "" {
			return fmt.Errorf("--file is required (JSON array of resources)")
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		e := cost.NewEstimator()
		report, err := e.EstimateFromJSON(data)
		if err != nil {
			return err
		}
		if format == "json" {
			out, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Print(cost.FormatText(report))
		}
		return nil
	},
}

func init() {
	costEstimateCmd.Flags().String("file", "", "JSON file with resource list")
	costEstimateCmd.Flags().String("format", "text", "Output format: text or json")
	costCmd.AddCommand(costEstimateCmd)
	rootCmd.AddCommand(costCmd)
}
