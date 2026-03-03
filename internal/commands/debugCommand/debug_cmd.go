package debugCommand

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yourname/go-conf-demo/internal/config"
)

// NewDebugCommand creates the debug command
// Accepts a function that returns the current config, allowing the command
// to access the loaded configuration without creating import cycles.
func NewDebugCommand(getConfig func() *config.Config) *cobra.Command {
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Print the loaded configuration",
		Long: `Debug command displays the current application configuration.

This is useful for verifying that your configuration is loaded correctly
from all sources (config file, environment variables, CLI flags) with
proper priority ordering.`,
		Run: func(cobraCmd *cobra.Command, args []string) {
			runDebug(getConfig())
		},
	}

	return debugCmd
}

func runDebug(cfg *config.Config) {

	fmt.Println("=== Configuration Debug ===")
	fmt.Println()

	// App Configuration
	fmt.Println("Application:")
	fmt.Printf("  Name:        %s\n", cfg.App.Name)
	fmt.Printf("  Version:     %s\n", cfg.App.Version)
	fmt.Printf("  Environment: %s\n", cfg.App.Environment)
	fmt.Printf("  Debug:       %t\n", cfg.App.Debug)
	fmt.Println()

	// Database Configuration
	fmt.Println("Database:")
	fmt.Printf("  Host:            %s\n", cfg.Database.Host)
	fmt.Printf("  Port:            %d\n", cfg.Database.Port)
	fmt.Printf("  Database:        %s\n", cfg.Database.Name)
	fmt.Printf("  User:            %s\n", cfg.Database.User)
	fmt.Printf("  Password:        %s\n", maskPassword(cfg.Database.Password))
	fmt.Printf("  Max Connections: %d\n", cfg.Database.MaxConnections)
	fmt.Printf("  Connect Timeout: %s\n", cfg.Database.ConnectTimeout)
	fmt.Printf("  SSL Enabled:     %t\n", cfg.Database.EnableSSL)
	fmt.Println()

	// Server Configuration
	fmt.Println("Server:")
	fmt.Printf("  Host:             %s\n", cfg.Server.Host)
	fmt.Printf("  Port:             %d\n", cfg.Server.Port)
	fmt.Printf("  Read Timeout:     %s\n", cfg.Server.ReadTimeout)
	fmt.Printf("  Write Timeout:    %s\n", cfg.Server.WriteTimeout)
	fmt.Printf("  Shutdown Timeout: %s\n", cfg.Server.ShutdownTimeout)
	fmt.Printf("  TLS Enabled:      %t\n", cfg.Server.TLSEnabled)
	fmt.Println()

	// Logging Configuration
	fmt.Println("Logging:")
	fmt.Printf("  Level:  %s\n", cfg.Logging.Level)
	fmt.Printf("  Format: %s\n", cfg.Logging.Format)
	fmt.Printf("  Output: %s\n", cfg.Logging.Output)
	fmt.Println()

	// Feature Flags
	fmt.Println("Feature Flags:")
	fmt.Printf("  Metrics:     %t\n", cfg.Features.EnableMetrics)
	fmt.Printf("  Tracing:     %t\n", cfg.Features.EnableTracing)
	fmt.Printf("  Rate Limits: %t\n", cfg.Features.EnableRateLimits)
	fmt.Println()
}

func maskPassword(password string) string {
	if password == "" {
		return "(not set)"
	}
	return "********"
}

func init() {
	// Register with root command
	// This will be called from main.go
}
