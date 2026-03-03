// EXAMPLE: Production-ready configuration with OS-specific paths
// This file shows how to implement OS-specific config directories.
// Copy the relevant parts into appConfig.go when you're ready for production.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

// getDefaultConfigPath returns the OS-specific default config file path.
// This follows platform conventions:
//   - Linux:   ~/.config/myapp/config.yaml (respects $XDG_CONFIG_HOME)
//   - macOS:   ~/Library/Application Support/myapp/config.yaml
//   - Windows: %AppData%\Roaming\myapp\config.yaml
func getDefaultConfigPath() string {
	// os.UserConfigDir() returns the appropriate directory for the current OS
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to current directory if we can't determine the config dir
		// (this should rarely happen on modern systems)
		return "config.yaml"
	}

	// Create app-specific subdirectory
	// Replace "myapp" with your actual application name
	appName := "myapp"
	appConfigDir := filepath.Join(configDir, appName)

	// Optional: Create the directory if it doesn't exist
	// Uncomment this if you want the directory auto-created
	// if err := os.MkdirAll(appConfigDir, 0755); err != nil {
	//     return "config.yaml" // Fallback if we can't create directory
	// }

	// Return full path to config file
	return filepath.Join(appConfigDir, "config.yaml")
}

// getDefaultConfigPathWithFallbacks provides multiple fallback locations.
// It checks locations in order and returns the first existing file,
// or the primary location if no config exists yet.
func getDefaultConfigPathWithFallbacks() string {
	appName := "myapp"

	// Build list of locations to check (in priority order)
	var locations []string

	// 1. OS-specific config directory (highest priority for config location)
	if configDir, err := os.UserConfigDir(); err == nil {
		locations = append(locations, filepath.Join(configDir, appName, "config.yaml"))
	}

	// 2. Current working directory (for development)
	if cwd, err := os.Getwd(); err == nil {
		locations = append(locations, filepath.Join(cwd, "config.yaml"))
	}

	// 3. Executable's directory (useful for portable apps)
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		locations = append(locations, filepath.Join(exeDir, "config.yaml"))
	}

	// Return first existing file
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// No config found, return primary location (will be created if needed)
	if len(locations) > 0 {
		return locations[0]
	}

	return "config.yaml" // Ultimate fallback
}

// LoadConfigProduction demonstrates production-ready config loading with OS-specific paths.
// This is what you'd use in a real application.
func LoadConfigProduction() (*Config, error) {
	// Get the default config path for this OS
	defaultConfigPath := getDefaultConfigPath()

	// Or use the version with fallbacks:
	// defaultConfigPath := getDefaultConfigPathWithFallbacks()

	// Create a flag set for CLI arguments
	fs := pflag.NewFlagSet("config", pflag.ContinueOnError)

	// Use OS-specific path as default, but allow override via --config flag
	fs.String("config", defaultConfigPath, "Path to configuration file")
	fs.String("app.environment", "", "Application environment (dev, staging, prod)")
	fs.Bool("app.debug", false, "Enable debug mode")
	fs.String("server.host", "", "Server host address")
	fs.Int("server.port", 0, "Server port")
	fs.String("database.host", "", "Database host")
	fs.Int("database.port", 0, "Database port")
	fs.String("logging.level", "", "Log level (debug, info, warn, error)")

	// Parse CLI arguments
	if err := fs.Parse(os.Args[1:]); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	// Get the final config file path (may be overridden by --config flag)
	configPath, _ := fs.GetString("config")

	// Create the generic loader with options
	loader, err := NewLoader(
		WithConfigFile(configPath),
		WithEnvPrefix("APP"),
		WithFlagSet(fs),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	// Set default values (lowest priority)
	setDefaultsProduction(loader)

	// Load configuration from all sources
	if err := loader.Load(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Unmarshal into the Config struct
	var cfg Config
	if err := loader.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaultsProduction sets production-appropriate default values
func setDefaultsProduction(loader *Loader) {
	// App defaults
	loader.Set("app.name", "myapp")
	loader.Set("app.environment", "production")
	loader.Set("app.debug", false)
	loader.Set("app.version", "1.0.0")

	// Database defaults
	loader.Set("database.host", "localhost")
	loader.Set("database.port", 5432)
	loader.Set("database.user", "postgres")
	loader.Set("database.password", "")
	loader.Set("database.name", "myapp")
	loader.Set("database.max_connections", 25)
	loader.Set("database.connect_timeout", "30s")
	loader.Set("database.enable_ssl", true) // SSL on by default in production

	// Server defaults
	loader.Set("server.host", "0.0.0.0")
	loader.Set("server.port", 8080)
	loader.Set("server.read_timeout", "30s")
	loader.Set("server.write_timeout", "30s")
	loader.Set("server.shutdown_timeout", "10s")
	loader.Set("server.tls_enabled", true) // TLS on by default in production

	// Logging defaults
	loader.Set("logging.level", "info")
	loader.Set("logging.format", "json")
	loader.Set("logging.output", "stdout")

	// Feature flags defaults
	loader.Set("features.enable_metrics", true)
	loader.Set("features.enable_tracing", true)
	loader.Set("features.enable_rate_limits", true)
}

// USAGE EXAMPLE:
//
// In your main.go, instead of:
//   cfg, err := config.LoadConfig("config.yaml")
//
// Use:
//   cfg, err := config.LoadConfigProduction()
//
// This will automatically look for config in:
//   Linux:   ~/.config/myapp/config.yaml
//   macOS:   ~/Library/Application Support/myapp/config.yaml
//   Windows: %AppData%\Roaming\myapp\config.yaml
//
// Users can still override with:
//   ./myapp --config=/custom/path/config.yaml
//
// Or set environment variable:
//   export APP_CONFIG_FILE=/custom/path/config.yaml (would need to add this flag)

// INSTALLATION INSTRUCTIONS FOR USERS:
//
// When distributing your app, document that users should:
//
// 1. Create config directory (done automatically if you uncomment os.MkdirAll above):
//    Linux:   mkdir -p ~/.config/myapp
//    macOS:   mkdir -p ~/Library/Application\ Support/myapp
//    Windows: mkdir %AppData%\myapp
//
// 2. Copy config file to that location:
//    Linux:   cp config.yaml.example ~/.config/myapp/config.yaml
//    macOS:   cp config.yaml.example ~/Library/Application\ Support/myapp/config.yaml
//    Windows: copy config.yaml.example %AppData%\myapp\config.yaml
//
// 3. Edit the config file with their settings
//
// 4. Run the app (it finds config automatically):
//    ./myapp
