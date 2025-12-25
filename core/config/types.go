// Package config provides configuration loading and type definitions.
package config

// ConfigMode defines the validation mode for configuration
type ConfigMode string

const (
	ConfigModePermissive ConfigMode = "permissive"
	ConfigModeStrict     ConfigMode = "strict"
)
