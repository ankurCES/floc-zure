// Package config handles loading, validating, and persisting azfloci configuration.
package config

import "github.com/ankurCES/floc-zure/pkg/models"

// Manager loads and persists azfloci configuration.
// Config sources (highest priority first): CLI flags, env vars, config file, defaults.
type Manager interface {
	// Load reads config from file, env, and defaults. Returns merged config.
	Load() (*models.Config, error)

	// Save persists the current config to the config file.
	Save(cfg *models.Config) error

	// Get returns the current in-memory config (call Load first).
	Get() *models.Config

	// SetDefault sets a default config value by key path (e.g., "defaults.location").
	SetDefault(key, value string) error

	// GetConfigPath returns the path to the active config file.
	GetConfigPath() string
}
