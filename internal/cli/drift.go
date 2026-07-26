package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ankurCES/floc-zure/internal/drift"
	"github.com/spf13/cobra"
)

func defaultSimStatePath() string {
	if p := os.Getenv("AZFLOCI_SIM_STATE"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".azfloci", "sim-state.json")
}

var driftCmd = &cobra.Command{
	Use:   "drift",
	Short: "Detect configuration drift in simulator state",
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Capture current simulator state as a snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		statePath, _ := cmd.Flags().GetString("state")
		output, _ := cmd.Flags().GetString("output")
		label, _ := cmd.Flags().GetString("label")
		if statePath == "" { statePath = defaultSimStatePath() }
		if output == "" { output = "snapshot.json" }
		if label == "" { label = "snapshot" }
		snap, err := drift.CaptureFromFile(statePath, label)
		if err != nil { return err }
		if err := drift.SaveSnapshot(snap, output); err != nil { return err }
		fmt.Printf("Snapshot saved: %s (%d resources)\n", output, len(snap.Resources))
		return nil
	},
}

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two snapshots and report drift",
	RunE: func(cmd *cobra.Command, args []string) error {
		before, _ := cmd.Flags().GetString("before")
		after, _ := cmd.Flags().GetString("after")
		format, _ := cmd.Flags().GetString("format")
		if before == "" || after == "" { return fmt.Errorf("--before and --after are required") }
		snap1, err := drift.LoadSnapshot(before)
		if err != nil { return fmt.Errorf("load before: %w", err) }
		snap2, err := drift.LoadSnapshot(after)
		if err != nil { return fmt.Errorf("load after: %w", err) }
		report := drift.Compare(snap1, snap2)
		if format == "json" {
			data, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(data))
		} else { fmt.Print(drift.FormatText(report)) }
		return nil
	},
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Snapshot current state and compare against a baseline",
	RunE: func(cmd *cobra.Command, args []string) error {
		statePath, _ := cmd.Flags().GetString("state")
		baseline, _ := cmd.Flags().GetString("baseline")
		format, _ := cmd.Flags().GetString("format")
		if statePath == "" { statePath = defaultSimStatePath() }
		if baseline == "" { return fmt.Errorf("--baseline is required") }
		snap1, err := drift.LoadSnapshot(baseline)
		if err != nil { return fmt.Errorf("load baseline: %w", err) }
		snap2, err := drift.CaptureFromFile(statePath, "current")
		if err != nil { return err }
		report := drift.Compare(snap1, snap2)
		if format == "json" {
			data, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(data))
		} else { fmt.Print(drift.FormatText(report)) }
		return nil
	},
}

func init() {
	snapshotCmd.Flags().String("state", "", "Path to simulator state file")
	snapshotCmd.Flags().String("output", "snapshot.json", "Output snapshot file")
	snapshotCmd.Flags().String("label", "snapshot", "Snapshot label")
	compareCmd.Flags().String("before", "", "Before snapshot file")
	compareCmd.Flags().String("after", "", "After snapshot file")
	compareCmd.Flags().String("format", "text", "Output format: text or json")
	reportCmd.Flags().String("state", "", "Path to simulator state file")
	reportCmd.Flags().String("baseline", "", "Baseline snapshot file")
	reportCmd.Flags().String("format", "text", "Output format: text or json")
	driftCmd.AddCommand(snapshotCmd, compareCmd, reportCmd)
	rootCmd.AddCommand(driftCmd)
}
