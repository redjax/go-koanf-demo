// This is a simple Go application demonstrating how to use the Koanf library for configuration management.
// It loads configuration from multiple sources (config file, environment variables, CLI flags) and prints
// the loaded configuration to the console.
//
// This reference is for simpler applications. For more complex apps, see the internal/config package which
// provides a more robust and flexible configuration loader.
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

// Config struct - define your app configuration here
type Config struct {
	AppName string `koanf:"app_name"`
	Port    int    `koanf:"port"`
	Debug   bool   `koanf:"debug"`
	Timeout string `koanf:"timeout"`
}

func main() {
	// Setup flags
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	fs.String("config", "config.yaml", "path to config file")
	fs.String("app_name", "", "application name")
	fs.Int("port", 0, "server port")
	fs.Bool("debug", false, "enable debug mode")
	fs.String("timeout", "", "request timeout")
	fs.Parse(os.Args[1:])

	// Load configuration
	cfg, err := loadConfig(fs)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use configuration
	fmt.Printf("=== Simple Config Demo ===\n")
	fmt.Printf("App Name: %s\n", cfg.AppName)
	fmt.Printf("Port:     %d\n", cfg.Port)
	fmt.Printf("Debug:    %t\n", cfg.Debug)
	fmt.Printf("Timeout:  %s\n", cfg.Timeout)
	fmt.Println()

	// Your app logic here...
	if cfg.Debug {
		fmt.Println("Debug mode is enabled")
	}
	fmt.Printf("Would start server on port %d...\n", cfg.Port)
}

// loadConfig loads configuration from file, env vars, and CLI flags
// Priority: CLI flags > Env vars > Config file > Defaults
func loadConfig(fs *flag.FlagSet) (*Config, error) {
	k := koanf.New(".")

	// 1. Set defaults
	k.Set("app_name", "simple-app")
	k.Set("port", 8080)
	k.Set("debug", false)
	k.Set("timeout", "30s")

	// 2. Load from config file (if exists)
	configFile, _ := fs.GetString("config")
	if _, err := os.Stat(configFile); err == nil {
		var parser koanf.Parser
		switch {
		case hasExt(configFile, ".yaml", ".yml"):
			parser = yaml.Parser()
		case hasExt(configFile, ".json"):
			parser = json.Parser()
		case hasExt(configFile, ".toml"):
			parser = toml.Parser()
		default:
			return nil, fmt.Errorf("unsupported config file format: %s", configFile)
		}

		if err := k.Load(file.Provider(configFile), parser); err != nil {
			return nil, fmt.Errorf("error loading config file: %w", err)
		}
	}

	// 3. Load from environment variables (APP_ prefix)
	k.Load(env.Provider("APP_", ".", func(s string) string {
		return s[4:] // Strip APP_ prefix
	}), nil)

	// 4. Load from CLI flags (highest priority)
	k.Load(posflag.Provider(fs, ".", k), nil)

	// Parse into struct
	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate
	if cfg.Port < 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", cfg.Port)
	}
	if _, err := time.ParseDuration(cfg.Timeout); err != nil {
		return nil, fmt.Errorf("invalid timeout format: %s", cfg.Timeout)
	}

	return &cfg, nil
}

// hasExt checks if filename has any of the given extensions
func hasExt(filename string, exts ...string) bool {
	for _, ext := range exts {
		if len(filename) >= len(ext) && filename[len(filename)-len(ext):] == ext {
			return true
		}
	}
	return false
}
