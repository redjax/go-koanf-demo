package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourname/go-conf-demo/internal/commands/debugCommand"
	"github.com/yourname/go-conf-demo/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "go-conf-demo",
	Short: "A demo application showcasing koanf configuration management",
	Long: `go-conf-demo demonstrates a reusable configuration pattern using koanf.

This CLI shows how to integrate the generic config loader with Cobra commands,
supporting multiple config formats (YAML, JSON, TOML, .env) with proper priority
ordering: CLI flags > Environment variables > Config file > Defaults.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration before any command runs
		// Note: LoadConfig creates its own flag set and parses os.Args.
		// This means flags are parsed twice (once by Cobra, once by pflag in LoadConfig),
		// but it allows the config module to remain independent and reusable.
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file path")
	rootCmd.PersistentFlags().String("app.environment", "", "application environment (development, staging, production)")
	rootCmd.PersistentFlags().Bool("app.debug", false, "enable debug mode")
	rootCmd.PersistentFlags().Int("server.port", 0, "server port")
	rootCmd.PersistentFlags().String("logging.level", "", "log level (debug, info, warn, error)")

	// Register subcommands
	// Pass GetConfig as a function to avoid import cycles
	rootCmd.AddCommand(debugCommand.NewDebugCommand(GetConfig))
}

// GetConfig returns the loaded configuration
func GetConfig() *config.Config {
	return cfg
}
