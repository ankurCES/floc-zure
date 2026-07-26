package config

import (
	"os"
	"path/filepath"

	"github.com/ankurCES/floc-zure/pkg/models"
	"github.com/spf13/viper"
)

type ManagerImpl struct {
	v   *viper.Viper
	cfg *models.Config
}

func NewManager() *ManagerImpl {
	v := viper.New()
	v.SetConfigName(".azfloci")
	v.SetConfigType("yaml")
	v.AddConfigPath("$HOME")
	v.AddConfigPath(".")
	v.SetEnvPrefix("AZFLOCI")
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("output_format", "json")
	v.SetDefault("location", "eastus")
	v.SetDefault("verbose", false)

	return &ManagerImpl{v: v}
}

func (m *ManagerImpl) Load() (*models.Config, error) {
	_ = m.v.ReadInConfig() // OK if missing
	var cfg models.Config
	if err := m.v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	m.cfg = &cfg
	return &cfg, nil
}

func (m *ManagerImpl) Save(cfg *models.Config) error {
	path := m.GetConfigPath()
	m.v.Set("subscription", cfg.Subscription)
	m.v.Set("location", cfg.Location)
	m.v.Set("output_format", cfg.OutputFormat)
	m.v.Set("verbose", cfg.Verbose)
	m.v.Set("tags", cfg.Tags)
	m.v.Set("defaults", cfg.Defaults)
	return m.v.WriteConfigAs(path)
}

func (m *ManagerImpl) Get() *models.Config {
	return m.cfg
}

func (m *ManagerImpl) SetDefault(key, value string) error {
	m.v.Set(key, value)
	return m.v.WriteConfig()
}

func (m *ManagerImpl) GetConfigPath() string {
	if f := m.v.ConfigFileUsed(); f != "" {
		return f
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".azfloci.yaml")
}
