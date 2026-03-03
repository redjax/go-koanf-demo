// Example main package that demonstrates how to use the configuration loader defined in internal/config.
//
// This is a simple application that loads configuration from multiple sources (config file, environment variables, CLI flags)
// and prints the loaded configuration to the console. It also simulates how the application would use the configuration values.
//
// You can run this application with different configuration sources to see how the priority works:
//   - Change values in config.yaml
//   - Use a different format: --config=config.json
//   - Set environment variables: APP_SERVER_PORT=9090
//   - Pass CLI flags: --server.port=3000 --app.debug=true
//   - Combine all sources to see priority in action
package main

import (
	"fmt"
	"os"

	"github.com/yourname/go-conf-demo/internal/config"
)

func main() {
	fmt.Println("=== Go Configuration Demo with koanf ===")

	// Load configuration
	// By default, it will look for config.yaml, but you can override with --config flag
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Print loaded configuration
	fmt.Println("Successfully loaded configuration!")
	fmt.Println(cfg.String())
	fmt.Println()

	// Display detailed configuration sections
	printAppConfig(cfg)
	printDatabaseConfig(cfg)
	printServerConfig(cfg)
	printLoggingConfig(cfg)
	printFeatureFlags(cfg)

	// Demonstrate that the app would use this configuration
	fmt.Println("\n=== Application Simulation ===")
	if cfg.App.Debug {
		fmt.Println("Debug mode is ENABLED")
	} else {
		fmt.Println("Debug mode is DISABLED")
	}

	// Simulate starting the server with the loaded configuration
	fmt.Printf("Starting %s v%s in %s environment\n",
		cfg.App.Name,
		cfg.App.Version,
		cfg.App.Environment,
	)
	fmt.Printf("Server would listen on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Database would connect to %s:%d\n", cfg.Database.Host, cfg.Database.Port)
	fmt.Printf("Logging at %s level to %s\n", cfg.Logging.Level, cfg.Logging.Output)

	fmt.Println("\nConfiguration demo complete!")
	fmt.Println("\nTry running with different sources:")
	fmt.Println("  • Change values in config.yaml")
	fmt.Println("  • Use a different format: --config=config.json")
	fmt.Println("  • Set environment variables: APP_SERVER_PORT=9090")
	fmt.Println("  • Pass CLI flags: --server.port=3000 --app.debug=true")
	fmt.Println("  • Combine all sources to see priority in action!")
}

// printAppConfig prints the application-level configuration.
func printAppConfig(cfg *config.Config) {
	fmt.Println("📱 Application Configuration:")
	fmt.Printf("  Name:        %s\n", cfg.App.Name)
	fmt.Printf("  Version:     %s\n", cfg.App.Version)
	fmt.Printf("  Environment: %s\n", cfg.App.Environment)
	fmt.Printf("  Debug:       %t\n", cfg.App.Debug)
	fmt.Println()
}

// printDatabaseConfig prints the database configuration.
func printDatabaseConfig(cfg *config.Config) {
	fmt.Println("💾 Database Configuration:")
	fmt.Printf("  Host:            %s\n", cfg.Database.Host)
	fmt.Printf("  Port:            %d\n", cfg.Database.Port)
	fmt.Printf("  Database:        %s\n", cfg.Database.Name)
	fmt.Printf("  User:            %s\n", cfg.Database.User)
	fmt.Printf("  Password:        %s\n", maskPassword(cfg.Database.Password))
	fmt.Printf("  Max Connections: %d\n", cfg.Database.MaxConnections)
	fmt.Printf("  Connect Timeout: %s\n", cfg.Database.ConnectTimeout)
	fmt.Printf("  SSL Enabled:     %t\n", cfg.Database.EnableSSL)
	fmt.Println()
}

// printServerConfig prints the HTTP server configuration.
func printServerConfig(cfg *config.Config) {
	fmt.Println("Server Configuration:")
	fmt.Printf("  Host:             %s\n", cfg.Server.Host)
	fmt.Printf("  Port:             %d\n", cfg.Server.Port)
	fmt.Printf("  Read Timeout:     %s\n", cfg.Server.ReadTimeout)
	fmt.Printf("  Write Timeout:    %s\n", cfg.Server.WriteTimeout)
	fmt.Printf("  Shutdown Timeout: %s\n", cfg.Server.ShutdownTimeout)
	fmt.Printf("  TLS Enabled:      %t\n", cfg.Server.TLSEnabled)
	fmt.Println()
}

// printLoggingConfig prints the logging configuration.
func printLoggingConfig(cfg *config.Config) {
	fmt.Println("Logging Configuration:")
	fmt.Printf("  Level:  %s\n", cfg.Logging.Level)
	fmt.Printf("  Format: %s\n", cfg.Logging.Format)
	fmt.Printf("  Output: %s\n", cfg.Logging.Output)
	fmt.Println()
}

// printFeatureFlags prints the feature flags configuration.
func printFeatureFlags(cfg *config.Config) {
	fmt.Println("Feature Flags:")
	fmt.Printf("  Metrics:     %t\n", cfg.Features.EnableMetrics)
	fmt.Printf("  Tracing:     %t\n", cfg.Features.EnableTracing)
	fmt.Printf("  Rate Limits: %t\n", cfg.Features.EnableRateLimits)
	fmt.Println()
}

// maskPassword is a helper function to mask sensitive information when printing configuration values.
func maskPassword(password string) string {
	if password == "" {
		return "(not set)"
	}
	return "********"
}
