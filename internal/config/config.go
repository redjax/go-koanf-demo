// Package config provides a generic, reusable configuration loader using koanf.
// This file can be copy/pasted into other projects without modification.
// App-specific configuration structs should be defined in separate files.
//
// OS-SPECIFIC CONFIG PATHS:
// To use standard OS config directories instead of ./config.yaml, you can add a helper
// function to get the appropriate config directory for each OS:
//
//	func getDefaultConfigPath(appName, filename string) string {
//	    // Use Go's built-in os.UserConfigDir() which returns:
//	    //   - Linux:   $XDG_CONFIG_HOME or ~/.config
//	    //   - macOS:   ~/Library/Application Support
//	    //   - Windows: %AppData% (Roaming)
//	    configDir, err := os.UserConfigDir()
//	    if err != nil {
//	        // Fallback to current directory if we can't determine config dir
//	        return filename
//	    }
//	    // Build path: ~/.config/myapp/config.yaml (Linux)
//	    //             ~/Library/Application Support/myapp/config.yaml (macOS)
//	    //             C:\Users\Username\AppData\Roaming\myapp\config.yaml (Windows)
//	    return filepath.Join(configDir, appName, filename)
//	}
//
// Then in appConfig.go's LoadConfig(), use this as the default:
//
//	defaultPath := getDefaultConfigPath("myapp", "config.yaml")
//	fs.String("config", defaultPath, "Path to configuration file")
//
// The app will automatically create/look for config in:
//   - ~/.config/myapp/config.yaml (Linux)
//   - ~/Library/Application Support/myapp/config.yaml (macOS)
//   - %AppData%\myapp\config.yaml (Windows)
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// Loader is a generic configuration loader that supports multiple sources.
// It follows a priority order: CLI flags > Environment variables > Config file > Defaults
type Loader struct {
	k          *koanf.Koanf
	configFile string
	envPrefix  string
	flagSet    *pflag.FlagSet
	parser     koanf.Parser
	loaded     bool
}

// LoaderOption is a functional option for configuring the Loader
type LoaderOption func(*Loader) error

// NewLoader creates a new configuration loader with sensible defaults.
// Use LoaderOptions to customize behavior.
func NewLoader(opts ...LoaderOption) (*Loader, error) {
	l := &Loader{
		k:         koanf.New("."),
		envPrefix: "",
		flagSet:   pflag.NewFlagSet("config", pflag.ContinueOnError),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(l); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return l, nil
}

// WithConfigFile sets the configuration file path and auto-detects the format.
// Supported formats: .yaml, .yml, .json, .toml, .env
func WithConfigFile(path string) LoaderOption {
	return func(l *Loader) error {
		if path == "" {
			return nil
		}

		l.configFile = path

		// Auto-detect parser based on file extension
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".yaml", ".yml":
			l.parser = yaml.Parser()
		case ".json":
			l.parser = json.Parser()
		case ".toml":
			l.parser = toml.Parser()
		case ".env":
			l.parser = dotenv.Parser()
		default:
			return fmt.Errorf("unsupported config file format: %s", ext)
		}

		return nil
	}
}

// WithEnvPrefix sets the environment variable prefix.
// For example, with prefix "APP", environment variable "APP_DATABASE_HOST"
// will map to config key "database.host"
func WithEnvPrefix(prefix string) LoaderOption {
	return func(l *Loader) error {
		l.envPrefix = prefix
		return nil
	}
}

// WithFlagSet allows you to provide a custom pflag.FlagSet for CLI argument parsing
func WithFlagSet(fs *pflag.FlagSet) LoaderOption {
	return func(l *Loader) error {
		l.flagSet = fs
		return nil
	}
}

// Load executes the configuration loading in priority order:
// 1. Default values (if any were set via Set)
// 2. Configuration file (if specified)
// 3. Environment variables
// 4. CLI flags (if flagSet was parsed)
func (l *Loader) Load() error {
	if l.loaded {
		return fmt.Errorf("configuration already loaded")
	}

	// Load from config file if specified
	if l.configFile != "" {
		if _, err := os.Stat(l.configFile); err == nil {
			if err := l.k.Load(file.Provider(l.configFile), l.parser); err != nil {
				return fmt.Errorf("failed to load config file %s: %w", l.configFile, err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check config file %s: %w", l.configFile, err)
		}
		// If file doesn't exist, continue silently (optional config file)
	}

	// Load from environment variables
	envProvider := env.Provider(l.envPrefix, ".", func(s string) string {
		// Transform environment variable names to config keys
		// Example: APP_DATABASE_HOST -> database.host
		s = strings.TrimPrefix(s, l.envPrefix)
		s = strings.TrimPrefix(s, "_")
		return strings.Replace(strings.ToLower(s), "_", ".", -1)
	})

	if err := l.k.Load(envProvider, nil); err != nil {
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Load from CLI flags (if parsed)
	if l.flagSet != nil && l.flagSet.Parsed() {
		if err := l.k.Load(posflag.Provider(l.flagSet, ".", l.k), nil); err != nil {
			return fmt.Errorf("failed to load CLI flags: %w", err)
		}
	}

	l.loaded = true
	return nil
}

// Unmarshal unmarshals the loaded configuration into the provided struct.
// The struct should have koanf tags for mapping.
func (l *Loader) Unmarshal(target interface{}) error {
	if !l.loaded {
		return fmt.Errorf("configuration not loaded yet, call Load() first")
	}

	if err := l.k.Unmarshal("", target); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Set sets a default value for a config key
func (l *Loader) Set(key string, value interface{}) error {
	return l.k.Set(key, value)
}

// Get retrieves a value by key
func (l *Loader) Get(key string) interface{} {
	return l.k.Get(key)
}

// String returns a string value for the given key
func (l *Loader) String(key string) string {
	return l.k.String(key)
}

// Int returns an int value for the given key
func (l *Loader) Int(key string) int {
	return l.k.Int(key)
}

// Bool returns a bool value for the given key
func (l *Loader) Bool(key string) bool {
	return l.k.Bool(key)
}

// Koanf returns the underlying koanf instance for advanced usage
func (l *Loader) Koanf() *koanf.Koanf {
	return l.k
}

// All returns all configuration as a map
func (l *Loader) All() map[string]interface{} {
	return l.k.All()
}

// Print prints all configuration keys and values (useful for debugging)
func (l *Loader) Print() {
	fmt.Println("Configuration:")
	for k, v := range l.k.All() {
		fmt.Printf("  %s = %v\n", k, v)
	}
}
