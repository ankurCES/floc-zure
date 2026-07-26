package router

import (
	"testing"
)

func TestDispatch_MatchesTwoWordCommand(t *testing.T) {
	r := New()
	called := false
	r.Register("account show", func(args []string) int {
		called = true
		return 0
	})
	code := r.Dispatch([]string{"account", "show"})
	if !called {
		t.Error("handler not called")
	}
	if code != 0 {
		t.Errorf("exit code: %d", code)
	}
}

func TestDispatch_MatchesOneWordCommand(t *testing.T) {
	r := New()
	called := false
	r.Register("version", func(args []string) int {
		called = true
		return 0
	})
	code := r.Dispatch([]string{"version"})
	if !called {
		t.Error("handler not called")
	}
	if code != 0 {
		t.Errorf("exit code: %d", code)
	}
}

func TestDispatch_PassesRemainingArgs(t *testing.T) {
	r := New()
	var gotArgs []string
	r.Register("group create", func(args []string) int {
		gotArgs = args
		return 0
	})
	r.Dispatch([]string{"group", "create", "--name", "rg1", "--location", "eastus"})
	if len(gotArgs) != 4 {
		t.Fatalf("expected 4 remaining args, got %d: %v", len(gotArgs), gotArgs)
	}
}

func TestDispatch_StripsOutputFlag(t *testing.T) {
	r := New()
	var gotArgs []string
	r.Register("account show", func(args []string) int {
		gotArgs = args
		return 0
	})
	r.Dispatch([]string{"account", "show", "--output", "json"})
	if len(gotArgs) != 0 {
		t.Errorf("--output should be stripped, got %v", gotArgs)
	}
}

func TestDispatch_UnknownCommand(t *testing.T) {
	r := New()
	code := r.Dispatch([]string{"foo", "bar"})
	if code != 2 {
		t.Errorf("expected exit 2, got %d", code)
	}
}

func TestDispatch_NoArgs(t *testing.T) {
	r := New()
	code := r.Dispatch(nil)
	if code != 2 {
		t.Errorf("expected exit 2, got %d", code)
	}
}

func TestDispatch_Help(t *testing.T) {
	r := New()
	code := r.Dispatch([]string{"--help"})
	if code != 0 {
		t.Errorf("expected exit 0, got %d", code)
	}
}

func TestParseFlag(t *testing.T) {
	args := []string{"--name", "rg1", "--location", "eastus"}
	v, ok := ParseFlag(args, "--name")
	if !ok || v != "rg1" {
		t.Errorf("got %q %v", v, ok)
	}
	v, ok = ParseFlag(args, "--missing")
	if ok {
		t.Errorf("should not find --missing, got %q", v)
	}
}

func TestParseFlag_EqualsForm(t *testing.T) {
	args := []string{"--name=rg1"}
	v, ok := ParseFlag(args, "--name")
	if !ok || v != "rg1" {
		t.Errorf("got %q %v", v, ok)
	}
}

func TestParseFlagBool(t *testing.T) {
	args := []string{"--yes", "--no-wait"}
	if !ParseFlagBool(args, "--yes") {
		t.Error("--yes not found")
	}
	if ParseFlagBool(args, "--verbose") {
		t.Error("--verbose should not be found")
	}
}

func TestParseTags(t *testing.T) {
	args := []string{"--name", "rg1", "--tags", "env=test", "purpose=sim", "--location", "eastus"}
	tags := ParseTags(args)
	if tags["env"] != "test" {
		t.Errorf("env: %v", tags)
	}
	if tags["purpose"] != "sim" {
		t.Errorf("purpose: %v", tags)
	}
	if _, has := tags["location"]; has {
		t.Error("location should not be in tags")
	}
}
