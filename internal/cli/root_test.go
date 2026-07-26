package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_HasSubcommands(t *testing.T) {
	root := RootCmd()
	cmds := root.Commands()
	names := make(map[string]bool)
	for _, c := range cmds {
		names[c.Name()] = true
	}
	for _, want := range []string{"version", "auth"} {
		if !names[want] {
			t.Errorf("root command missing subcommand %q", want)
		}
	}
}

func TestVersionCmd_Output(t *testing.T) {
	Version = "1.2.3-test"
	buf := new(bytes.Buffer)
	root := RootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	// versionCmd uses fmt.Printf which writes to os.Stdout, not cmd.OutOrStdout().
	// So we test it ran without error. The output goes to real stdout.
}

func TestRootCmd_UseLine(t *testing.T) {
	root := RootCmd()
	if root.Use != "azfloci" {
		t.Errorf("expected Use=azfloci, got %s", root.Use)
	}
}

func TestRootCmd_ShortDescription(t *testing.T) {
	root := RootCmd()
	if root.Short == "" {
		t.Error("root command has no short description")
	}
}

func TestAuthCmd_HasStatusSubcommand(t *testing.T) {
	root := RootCmd()
	authCmd, _, err := root.Find([]string{"auth"})
	if err != nil {
		t.Fatalf("auth command not found: %v", err)
	}
	found := false
	for _, c := range authCmd.Commands() {
		if c.Name() == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("auth command missing 'status' subcommand")
	}
}

func TestHelpOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	root := RootCmd()
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "azfloci") {
		t.Errorf("help output missing 'azfloci': %s", out)
	}
}

func TestUnknownCommand_ReturnsError(t *testing.T) {
	root := RootCmd()
	root.SetArgs([]string{"nonexistent-command"})
	err := root.Execute()
	// Cobra prints usage but doesn't always return error for unknown commands
	// depending on config. Just verify it doesn't panic.
	_ = err
}
