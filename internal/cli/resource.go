package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/ankurCES/floc-zure/internal/azure"
	"github.com/ankurCES/floc-zure/internal/resource"
	"github.com/spf13/cobra"
)

var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage Azure resources",
}

var resourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources (optionally filtered by resource group)",
	RunE: func(cmd *cobra.Command, args []string) error {
		rg, _ := cmd.Flags().GetString("resource-group")
		mgr := resource.NewManager(azure.NewCLIExecutorImpl())
		resources, err := mgr.ListResources(context.Background(), rg)
		if err != nil {
			return err
		}
		return printJSON(cmd, resources)
	},
}

var resourceShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show a resource by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("ids")
		mgr := resource.NewManager(azure.NewCLIExecutorImpl())
		res, err := mgr.GetResource(context.Background(), id)
		if err != nil {
			return err
		}
		return printJSON(cmd, res)
	},
}

var resourceDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a resource by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("ids")
		mgr := resource.NewManager(azure.NewCLIExecutorImpl())
		if err := mgr.DeleteResource(context.Background(), id); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Resource %q deleted\n", id)
		return nil
	},
}

var resourceTagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Add/update tags on a resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("ids")
		tagsRaw, _ := cmd.Flags().GetStringSlice("tags")
		tags := make(map[string]string)
		for _, t := range tagsRaw {
			parts := strings.SplitN(t, "=", 2)
			if len(parts) == 2 {
				tags[parts[0]] = parts[1]
			}
		}
		mgr := resource.NewManager(azure.NewCLIExecutorImpl())
		if err := mgr.TagResource(context.Background(), id, tags); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Tags updated on %q\n", id)
		return nil
	},
}

func init() {
	resourceListCmd.Flags().StringP("resource-group", "g", "", "Filter by resource group")

	resourceShowCmd.Flags().String("ids", "", "Full ARM resource ID (required)")
	_ = resourceShowCmd.MarkFlagRequired("ids")

	resourceDeleteCmd.Flags().String("ids", "", "Full ARM resource ID (required)")
	_ = resourceDeleteCmd.MarkFlagRequired("ids")

	resourceTagCmd.Flags().String("ids", "", "Full ARM resource ID (required)")
	resourceTagCmd.Flags().StringSlice("tags", nil, "Tags as key=value pairs (required)")
	_ = resourceTagCmd.MarkFlagRequired("ids")
	_ = resourceTagCmd.MarkFlagRequired("tags")

	resourceCmd.AddCommand(resourceListCmd, resourceShowCmd, resourceDeleteCmd, resourceTagCmd)
	rootCmd.AddCommand(resourceCmd)
}
