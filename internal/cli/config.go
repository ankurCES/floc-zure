package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ankurCES/floc-zure/internal/config"
	"github.com/ankurCES/floc-zure/pkg/models"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage azfloci configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration file with defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := config.NewManager()
		cfg := &models.Config{
			Location:     "eastus",
			OutputFormat: "json",
			Verbose:      false,
		}
		if err := mgr.Save(cfg); err != nil {
			return fmt.Errorf("init config: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Config initialized at %s\n", mgr.GetConfigPath())
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set KEY=VALUE",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parts := strings.SplitN(args[0], "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("expected KEY=VALUE format, got %q", args[0])
		}
		key, value := parts[0], parts[1]

		mgr := config.NewManager()
		if _, err := mgr.Load(); err != nil {
			// If no config file, init one first
			_ = mgr.Save(&models.Config{Location: "eastus", OutputFormat: "json"})
		}
		if err := mgr.SetDefault(key, value); err != nil {
			return fmt.Errorf("set config: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Set %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get KEY",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		mgr := config.NewManager()
		cfg, err := mgr.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		val := configLookup(cfg, key)
		fmt.Fprintln(cmd.OutOrStdout(), val)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := config.NewManager()
		cfg, err := mgr.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		entries := map[string]string{
			"subscription":           cfg.Subscription,
			"location":               cfg.Location,
			"output_format":          cfg.OutputFormat,
			"verbose":                fmt.Sprintf("%t", cfg.Verbose),
			"defaults.resource_group": cfg.Defaults.ResourceGroup,
			"defaults.location":      cfg.Defaults.Location,
		}

		keys := make([]string, 0, len(entries))
		for k := range entries {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			v := entries[k]
			if v == "" {
				v = "(not set)"
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-25s %s\n", k, v)
		}

		if len(cfg.Tags) > 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "\nTags:")
			for k, v := range cfg.Tags {
				fmt.Fprintf(cmd.OutOrStdout(), "  %-20s %s\n", k, v)
			}
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd, configSetCmd, configGetCmd, configListCmd)
	rootCmd.AddCommand(configCmd)
}

// configLookup retrieves a value from Config by dot-path key.
func configLookup(cfg *models.Config, key string) string {
	switch key {
	case "subscription":
		return cfg.Subscription
	case "location":
		return cfg.Location
	case "output_format":
		return cfg.OutputFormat
	case "verbose":
		return fmt.Sprintf("%t", cfg.Verbose)
	case "defaults.resource_group":
		return cfg.Defaults.ResourceGroup
	case "defaults.location":
		return cfg.Defaults.Location
	default:
		return "(unknown key)"
	}
}
