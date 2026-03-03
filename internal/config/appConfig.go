//
// App-specific configuration module.
//
// While the config.go file is generic and copy/paste-able, this file contains app-specific configuration structs and loading logic.
// You should create a new appConfig.go for each new project and define your own configuration fields based on your application's needs.
//

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
)

// Define the environment variable prefix for this application.
// You can change this to something specific to your app, e.g. "MYAPP".
var envPrefix = "APP"

// Redefine the Config struct here with your app-specific fields.
// This is just an example structure. Modify it to fit your application's configuration needs.
type Config struct {
	// App-level settings
	App AppConfig `koanf:"app"`

	// Database configuration
	Database DatabaseConfig `koanf:"database"`

	// Server configuration
	Server ServerConfig `koanf:"server"`

	// Logging configuration
	Logging LogConfig `koanf:"logging"`

	// Feature flags
	Features FeatureFlags `koanf:"features"`
}

// AppConfig contains general application settings
type AppConfig struct {
	Name        string `koanf:"name"`
	Environment string `koanf:"environment"`
	Debug       bool   `koanf:"debug"`
	Version     string `koanf:"version"`
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Host           string        `koanf:"host"`
	Port           int           `koanf:"port"`
	User           string        `koanf:"user"`
	Password       string        `koanf:"password"`
	Name           string        `koanf:"name"`
	MaxConnections int           `koanf:"max_connections"`
	ConnectTimeout time.Duration `koanf:"connect_timeout"`
	EnableSSL      bool          `koanf:"enable_ssl"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	Host            string        `koanf:"host"`
	Port            int           `koanf:"port"`
	ReadTimeout     time.Duration `koanf:"read_timeout"`
	WriteTimeout    time.Duration `koanf:"write_timeout"`
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
	TLSEnabled      bool          `koanf:"tls_enabled"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"` // "json" or "text"
	Output string `koanf:"output"` // "stdout", "stderr", or file path
}

// FeatureFlags contains feature toggles
type FeatureFlags struct {
	EnableMetrics    bool `koanf:"enable_metrics"`
	EnableTracing    bool `koanf:"enable_tracing"`
	EnableRateLimits bool `koanf:"enable_rate_limits"`
}

// LoadConfig is a convenience function that loads configuration using the generic loader.
// This demonstrates the recommended pattern for your applications.
//
// PRODUCTION TIP - OS-SPECIFIC CONFIG PATHS:
// For production apps, instead of passing "config.yaml" directly, you should use
// OS-specific config directories.
//
//	// Add this helper function to this file:
//	func getDefaultConfigPath() string {
//	    // Get OS-specific config directory:
//	    //   Linux:   $XDG_CONFIG_HOME or ~/.config
//	    //   macOS:   ~/Library/Application Support
//	    //   Windows: %AppData%\Roaming
//	    configDir, err := os.UserConfigDir()
//	    if err != nil {
//	        return "config.yaml" // Fallback to current directory
//	    }
//
//	    // Recommended structure: <ConfigDir>/<AppName>/config.yaml
//	    // Examples:
//	    //   ~/.config/myapp/config.yaml (Linux)
//	    //   ~/Library/Application Support/myapp/config.yaml (macOS)
//	    //   C:\Users\YourName\AppData\Roaming\myapp\config.yaml (Windows)
//	    appConfigDir := filepath.Join(configDir, "myapp")
//
//	    // Optional: Create directory if it doesn't exist
//	    // os.MkdirAll(appConfigDir, 0755)
//
//	    return filepath.Join(appConfigDir, "config.yaml")
//	}
//
//	// Then use it as the default:
//	func LoadConfig() (*Config, error) {  // Remove configFile parameter
//	    defaultConfigPath := getDefaultConfigPath()
//	    fs.String("config", defaultConfigPath, "Path to configuration file")
//	    // ... rest of the function
//	}
//
// Users can still override with:
//   - CLI flag:       --config=/custom/path/config.yaml
//   - Environment:    APP_CONFIG_FILE=/custom/path/config.yaml
//
// NOTE: The current implementation uses "config.yaml" in the working directory
// for demonstration purposes. In production, implement the above pattern.
func LoadConfig(configFile string) (*Config, error) {
	// Create a flag set for CLI arguments
	fs := pflag.NewFlagSet("config", pflag.ContinueOnError)

	// Define CLI flags that can override config file and env vars
	// PRODUCTION: Change this line to use getDefaultConfigPath() as shown above
	fs.String("config", configFile, "Path to configuration file")
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

	// Get the config file path from flags (in case it was overridden)
	configPath, _ := fs.GetString("config")

	// Create the generic loader with options
	loader, err := NewLoader(
		WithConfigFile(configPath),
		WithEnvPrefix(envPrefix), // Uses the envPrefix var defined at the top of this file
		WithFlagSet(fs),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create config loader: %w", err)
	}

	// Set default values (lowest priority)
	setDefaults(loader)

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

// setDefaults sets default values for configuration.
// You can have different default sets for different environments (dev, staging, prod)
// by creating separate functions like setDefaultsProduction() and calling the appropriate
// one based on the environment.
func setDefaults(loader *Loader) {
	// App defaults
	loader.Set("app.name", "go-conf-demo")
	loader.Set("app.environment", "development")
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
	loader.Set("database.enable_ssl", false)

	// Server defaults
	loader.Set("server.host", "0.0.0.0")
	loader.Set("server.port", 8080)
	loader.Set("server.read_timeout", "30s")
	loader.Set("server.write_timeout", "30s")
	loader.Set("server.shutdown_timeout", "10s")
	loader.Set("server.tls_enabled", false)

	// Logging defaults
	loader.Set("logging.level", "info")
	loader.Set("logging.format", "json")
	loader.Set("logging.output", "stdout")

	// Feature flags defaults
	loader.Set("features.enable_metrics", true)
	loader.Set("features.enable_tracing", false)
	loader.Set("features.enable_rate_limits", true)
}

// Validates the configuration values.
// You should modify this function to include validation logic specific to your application's configuration needs.
func (c *Config) Validate() error {
	// Validate environment
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
		"test":        true,
	}
	if !validEnvs[c.App.Environment] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, production, or test)", c.App.Environment)
	}

	// Validate port ranges
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be 1-65535)", c.Server.Port)
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("invalid database port: %d (must be 1-65535)", c.Database.Port)
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.Logging.Level)
	}

	return nil
}

// String returns a string representation of the config (with sensitive data masked)
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{App: %s, Env: %s, Debug: %t, Server: %s:%d, DB: %s:%d/%s}",
		c.App.Name,
		c.App.Environment,
		c.App.Debug,
		c.Server.Host,
		c.Server.Port,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
	)
}
