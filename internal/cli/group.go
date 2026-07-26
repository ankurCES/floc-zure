package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/ankurCES/floc-zure/internal/resource"
	"github.com/spf13/cobra"
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage Azure resource groups",
}

var groupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a resource group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		location, _ := cmd.Flags().GetString("location")
		tagsRaw, _ := cmd.Flags().GetStringSlice("tags")

		tags := parseTags(tagsRaw)
		mgr := resource.NewManager(azure.NewCLIExecutorImpl())
		rg, err := mgr.CreateResourceGroup(context.Background(), name, location, tags)
		if err != nil {
			return err
		}
		return printJSON(cmd, rg)
	},
}

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resource groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := resource.NewManager(azure.NewCLIExecutorImpl())
		groups, err := mgr.ListResourceGroups(context.Background())
		if err != nil {
			return err
		}
		return printJSON(cmd, groups)
	},
}

var groupShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a resource group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		mgr := resource.NewManager(azure.NewCLIExecutorImpl())
		rg, err := mgr.GetResourceGroup(context.Background(), name)
		if err != nil {
			return err
		}
		return printJSON(cmd, rg)
	},
}

var groupDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a resource group",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		mgr := resource.NewManager(azure.NewCLIExecutorImpl())
		if err := mgr.DeleteResourceGroup(context.Background(), name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Resource group %q deletion initiated\n", name)
		return nil
	},
}

func init() {
	groupCreateCmd.Flags().StringP("name", "n", "", "Resource group name (required)")
	groupCreateCmd.Flags().StringP("location", "l", "eastus", "Azure region")
	groupCreateCmd.Flags().StringSlice("tags", nil, "Tags as key=value pairs")
	_ = groupCreateCmd.MarkFlagRequired("name")

	groupShowCmd.Flags().StringP("name", "n", "", "Resource group name (required)")
	_ = groupShowCmd.MarkFlagRequired("name")

	groupDeleteCmd.Flags().StringP("name", "n", "", "Resource group name (required)")
	_ = groupDeleteCmd.MarkFlagRequired("name")

	groupCmd.AddCommand(groupCreateCmd, groupListCmd, groupShowCmd, groupDeleteCmd)
	rootCmd.AddCommand(groupCmd)
}

// parseTags converts ["key=val", ...] to map.
func parseTags(raw []string) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	m := make(map[string]string, len(raw))
	for _, t := range raw {
		parts := strings.SplitN(t, "=", 2)
		if len(parts) == 2 {
			m[parts[0]] = parts[1]
		}
	}
	return m
}

// printJSON marshals v to indented JSON and writes to cmd's output.
func printJSON(cmd *cobra.Command, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}
