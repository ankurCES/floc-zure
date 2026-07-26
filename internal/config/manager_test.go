package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ankurCES/floc-zure/pkg/models"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.v == nil {
		t.Fatal("viper instance is nil")
	}
}

func TestLoad_Defaults(t *testing.T) {
	m := NewManager()
	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected output_format=json, got %s", cfg.OutputFormat)
	}
	if cfg.Location != "eastus" {
		t.Errorf("expected location=eastus, got %s", cfg.Location)
	}
	if cfg.Verbose {
		t.Error("expected verbose=false by default")
	}
}

func TestGet_BeforeLoad(t *testing.T) {
	m := NewManager()
	if m.Get() != nil {
		t.Error("expected Get() to return nil before Load()")
	}
}

func TestGet_AfterLoad(t *testing.T) {
	m := NewManager()
	_, _ = m.Load()
	cfg := m.Get()
	if cfg == nil {
		t.Fatal("expected Get() to return config after Load()")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".azfloci.yaml")

	m := NewManager()
	// Point viper to temp dir
	m.v.AddConfigPath(tmpDir)
	m.v.SetConfigFile(cfgPath)

	cfg := &models.Config{
		Subscription: "sub-save-test",
		Location:     "westus2",
		OutputFormat: "table",
		Verbose:      true,
		Tags:         map[string]string{"env": "test"},
		Defaults: models.ConfigDefaults{
			ResourceGroup: "rg-test",
			Location:      "westus2",
		},
	}

	err := m.Save(cfg)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load it back with a fresh manager
	m2 := NewManager()
	m2.v.SetConfigFile(cfgPath)
	loaded, err := m2.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Subscription != "sub-save-test" {
		t.Errorf("expected subscription sub-save-test, got %s", loaded.Subscription)
	}
	if loaded.Location != "westus2" {
		t.Errorf("expected location westus2, got %s", loaded.Location)
	}
	if loaded.OutputFormat != "table" {
		t.Errorf("expected output_format table, got %s", loaded.OutputFormat)
	}
	if !loaded.Verbose {
		t.Error("expected verbose=true")
	}
}

func TestSetDefault(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".azfloci.yaml")

	m := NewManager()
	m.v.SetConfigFile(cfgPath)

	// Must save initial config first so WriteConfig has a file
	_ = m.Save(&models.Config{Location: "eastus"})

	err := m.SetDefault("location", "northeurope")
	if err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Reload and check
	m2 := NewManager()
	m2.v.SetConfigFile(cfgPath)
	loaded, _ := m2.Load()
	if loaded.Location != "northeurope" {
		t.Errorf("expected location northeurope, got %s", loaded.Location)
	}
}

func TestGetConfigPath_Default(t *testing.T) {
	m := NewManager()
	p := m.GetConfigPath()
	if p == "" {
		t.Error("GetConfigPath returned empty string")
	}
	if filepath.Base(p) != ".azfloci.yaml" {
		t.Errorf("expected filename .azfloci.yaml, got %s", filepath.Base(p))
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("AZFLOCI_LOCATION", "japaneast")
	m := NewManager()
	cfg, err := m.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Location != "japaneast" {
		t.Errorf("expected env override location=japaneast, got %s", cfg.Location)
	}
}
