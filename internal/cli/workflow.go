package cli

import (
	"context"
	"fmt"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/ankurCES/floc-zure/internal/workflow"
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Run and manage Azure provisioning workflows",
}

var workflowRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a workflow from a YAML file",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		eng := workflow.NewEngine(azure.NewCLIExecutorImpl())

		wf, err := eng.LoadWorkflow(file)
		if err != nil {
			return fmt.Errorf("load workflow: %w", err)
		}

		if err := eng.Validate(wf); err != nil {
			return fmt.Errorf("validate workflow: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Running workflow: %s\n", wf.Name)
		result, err := eng.Execute(context.Background(), wf)
		if err != nil {
			return fmt.Errorf("execute workflow: %w", err)
		}

		for _, sr := range result.Steps {
			icon := "✓"
			if sr.Status == "failed" {
				icon = "✗"
			} else if sr.Status == "skipped" {
				icon = "○"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "  %s %s (%s", icon, sr.StepName, sr.Duration)
			if sr.Retries > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), ", %d retries", sr.Retries)
			}
			fmt.Fprintln(cmd.OutOrStdout(), ")")
			if sr.Error != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "    error: %s\n", sr.Error)
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Workflow %s: %s (%s)\n", wf.Name, result.Status, result.Duration)
		return nil
	},
}

var workflowValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a workflow YAML file without executing",
	RunE: func(cmd *cobra.Command, args []string) error {
		file, _ := cmd.Flags().GetString("file")
		eng := workflow.NewEngine(azure.NewCLIExecutorImpl())

		wf, err := eng.LoadWorkflow(file)
		if err != nil {
			return fmt.Errorf("load workflow: %w", err)
		}

		if err := eng.Validate(wf); err != nil {
			return fmt.Errorf("workflow invalid: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Workflow %q is valid (%d steps)\n", wf.Name, len(wf.Steps))
		return nil
	},
}

func init() {
	workflowRunCmd.Flags().StringP("file", "f", "", "Path to workflow YAML file (required)")
	_ = workflowRunCmd.MarkFlagRequired("file")

	workflowValidateCmd.Flags().StringP("file", "f", "", "Path to workflow YAML file (required)")
	_ = workflowValidateCmd.MarkFlagRequired("file")

	workflowCmd.AddCommand(workflowRunCmd, workflowValidateCmd)
	rootCmd.AddCommand(workflowCmd)
}
