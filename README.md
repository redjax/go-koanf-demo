# Go Configuration Demo with Koanf

Reference app with a generic [`config.go`](./internal/config/config.go) using [knadh/koanf](https://github.com/knadh/koanf) that can be copied and pasted into any app as a base/starting point, and an app-specific [`appConfig.go`](./internal/config/appConfig.go) that demonstrates a configuration object specific to this app.

## Overview

This demo provides a generic configuration loader that can be reused across projects. The pattern separates the configuration loading mechanism (generic) from app-specific configuration structs (customizable per project).

If your app is simple and you don't need all the complexity of the `internal/config/config.go` example, check the [`simple` example for a self-contained, 1-file version of the config](./cmd/simple/main.go).

## Features

- Multiple format support: YAML, TOML, JSON, and .env files
- Environment variables with customizable prefix (default: `APP_`)
- CLI flag overrides
- Priority ordering: CLI flags > Environment > Config file > Defaults
- Type-safe parsing into structs with validation
- Generic loader that requires no modifications when copied

## Project Structure

```shell
go-conf-demo/
├── cmd/
│   ├── api/
│   │   └── main.go                        # API server entrypoint
│   ├── go-conf-demo/
│   │   └── main.go                        # CLI entrypoint
│   └── root.go                            # Cobra CLI root command
├── internal/
│   ├── commands/
│   │   └── debugCommand/
│   │       └── debug_cmd.go               # Debug subcommand
│   └── config/
│       ├── config.go                      # Generic loader (copy as-is)
│       ├── appConfig.go                   # App-specific config
│       └── appConfig_production_example.go # Production example with OS paths
├── main.go                                # Standalone demo application
├── example.config.yaml                    # YAML example
├── example.config.json                    # JSON example
├── example.config.toml                    # TOML example
└── .env.example                           # .env example
```

## Quick Start

Install dependencies:

```shell
go mod download
```

Run the demo:

```shell
## Use default config.yaml
go run main.go

## Use different format
go run main.go --config=config.json
go run main.go --config=config.toml

## Override with environment variables
APP_SERVER_PORT=9090 go run main.go

## Override with CLI flags
go run main.go --server.port=3000 --app.debug=true

## Combine sources (CLI has highest priority)
APP_SERVER_PORT=9090 go run main.go --server.port=3000
```

## CLI Example (Cobra)

The repo includes a Cobra CLI example demonstrating config usage in a CLI context.

Build the CLI:

```shell
go build -o go-conf-demo.exe ./cmd/go-conf-demo
```

Run the debug command:

```shell
# Use default config
./go-conf-demo.exe debug

# Use different config file
./go-conf-demo.exe debug --config=config.json

# Override with flags
./go-conf-demo.exe debug --server.port=4000 --app.debug=false

# View available commands and flags
./go-conf-demo.exe --help
```

The CLI demonstrates:

- Integration with Cobra command framework
- Persistent flags across subcommands
- Config loading in `PersistentPreRunE` hook
- Passing config to subcommands without import cycles

See [`cmd/root.go`](./cmd/root.go) and [`internal/commands/debugCommand/debug_cmd.go`](./internal/commands/debugCommand/debug_cmd.go) for implementation details.

## API Example

A simple HTTP API example shows config usage in a web service context.

Build and run the API:

```shell
# Build the API
go build -o api.exe ./cmd/api

# Run with default config
./api.exe

# Run with different config
./api.exe config.json

# Override with environment variables
APP_SERVER_PORT=8080 ./api.exe
```

Test the endpoints:

```shell
# Health check
curl http://localhost:9090/health

# View configuration
curl http://localhost:9090/config

# Root endpoint
curl http://localhost:9090/
```

The API demonstrates:

- Loading config at server startup
- Using config values for server settings (host, port, timeouts)
- Graceful shutdown with configured timeout
- Exposing sanitized config via endpoint (passwords excluded)

See [`cmd/api/main.go`](./cmd/api/main.go) for implementation details.

## How It Works

### Configuration Priority

Priority from highest to lowest:

1. CLI flags (`--server.port=3000`)
2. Environment variables (`APP_SERVER_PORT=3000`)
3. Config file (`config.yaml`)
4. Default values (set programmatically)

### Generic Loader (config.go)

The generic loader handles all koanf complexity. This file never/rarely needs modification when copying to new projects.

```go
loader, err := NewLoader(
    WithConfigFile("config.yaml"),
    WithEnvPrefix("APP"),
    WithFlagSet(flagSet),
)
loader.Load()
loader.Unmarshal(&cfg)
```

### App-Specific Config (appConfig.go)

Define your application's configuration structure here:

```go
type Config struct {
    App      AppConfig      `koanf:"app"`
    Database DatabaseConfig `koanf:"database"`
    Server   ServerConfig   `koanf:"server"`
    // Add your own sections
}
```

Each time you create a new Go app and use this configuration module as a reference, create an `appConfig.go` with configurations specific to that app. You should not need to modify `config.go`; any app-specific configurations belong in an `appConfig.go`.

## Adapting to Your Project

### Copy the Generic Loader

Copy `internal/config/config.go` to your project without modifications.

### Customize App Config

Copy `internal/config/appConfig.go` to your project and modify for your needs:

```go
type Config struct {
    App    AppConfig    `koanf:"app"`
    
    // Add your custom sections
    Redis  RedisConfig  `koanf:"redis"`
    Cache  CacheConfig  `koanf:"cache"`
}

type RedisConfig struct {
    URL string `koanf:"url"`
    TTL string `koanf:"ttl"`
}
```

### Update Defaults

Modify `setDefaults()` in `appConfig.go` and set appropriate defaults for your app:

```go
func setDefaults(loader *Loader) {
    loader.Set("redis.url", "localhost:6379")
    loader.Set("redis.ttl", "5m")
}
```

### Add Validation

Add validation rules in `Validate()` specific to your app:

```go
func (c *Config) Validate() error {
    if c.Redis.URL == "" {
        return fmt.Errorf("redis.url is required")
    }
    return nil
}
```

## Environment Variable Mapping

With `WithEnvPrefix("APP")`, environment variables map to config keys:

| Environment Variable | Config Key | Example |
| --- | --- | --- |
| `APP_SERVER_PORT` | `server.port` | `APP_SERVER_PORT=8080` |
| `APP_DATABASE_HOST` | `database.host` | `APP_DATABASE_HOST=localhost` |
| `APP_APP_DEBUG` | `app.debug` | `APP_APP_DEBUG=true` |
| `APP_LOGGING_LEVEL` | `logging.level` | `APP_LOGGING_LEVEL=debug` |

Pattern: `PREFIX_SECTION_KEY` converts to `section.key` (underscores to dots, lowercase)

## Config File Formats

The app can parse the same configuration in many different formats:

YAML:

```yaml
server:
  port: 8080
```

JSON:

```json
{"server": {"port": 8080}}
```

TOML:

```toml
[server]
port = 8080
```

ENV:

```env
SERVER_PORT=8080
```

Environment Variables:

```shell
export APP_SERVER_PORT=8080
```

## Use Cases

### CLI Application

```go
func main() {
    cfg, _ := config.LoadConfig("config.yaml")
    // Use cfg throughout your app
}
```

### HTTP Server

```go
func main() {
    cfg, _ := config.LoadConfig("config.yaml")
    server := &http.Server{
        Addr: fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
        ReadTimeout: cfg.Server.ReadTimeout,
    }
    server.ListenAndServe()
}
```

### Library

```go
type Client struct {
    config *Config
}

func NewClient(configFile string) (*Client, error) {
    cfg, err := config.LoadConfig(configFile)
    if err != nil {
        return nil, err
    }
    return &Client{config: cfg}, nil
}
```

## Production: OS-Specific Config Paths

By default, this demo uses `config.yaml` in the current working directory. For production, use OS-specific config directories.

### Standard Config Directories

Linux/Unix:

- Location: `~/.config/<appname>/`
- Example: `~/.config/myapp/config.yaml`
- Respects: `$XDG_CONFIG_HOME`

macOS:

- Location: `~/Library/Application Support/<appname>/`
- Example: `~/Library/Application Support/myapp/config.yaml`

Windows:

- Location: `%AppData%\<appname>\`
- Example: `C:\Users\Username\AppData\Roaming\myapp\config.yaml`

See the [`appConfig_production_example.go` file](./internal/config/appConfig_production_example.go) for reference.

### Implementation

Add a helper function to `appConfig.go`:

```go
func getDefaultConfigPath() string {
    configDir, err := os.UserConfigDir()
    if err != nil {
        return "config.yaml" // Fallback
    }
    appName := "myapp"
    return filepath.Join(configDir, appName, "config.yaml")
}
```

Update `LoadConfig()`:

```go
// OLD:
fs.String("config", configFile, "Path to configuration file")

// NEW:
fs.String("config", getDefaultConfigPath(), "Path to configuration file")
```

See [`appConfig_production_example.go`](./internal/config/appConfig_production_example.go) for a complete working example with multiple fallback locations.

### Multiple Fallback Locations

Check locations in order:

1. OS-specific config directory
1. Current working directory (for development)
1. Executable's directory (for portable apps)

See `getDefaultConfigPathWithFallbacks()` in [`appConfig_production_example.go`](./internal/config/appConfig_production_example.go).

### Priority Order with OS Paths

1. CLI flag: `--config=/custom/path/config.yaml`
2. Environment variable: `MYAPP_CONFIG_FILE=/custom/path/config.yaml`
3. OS-specific default: `~/.config/myapp/config.yaml`
4. Current directory: `./config.yaml`
5. Embedded defaults: `loader.Set()`

### User Installation

Linux/macOS:

```bash
mkdir -p ~/.config/myapp
cp config.yaml.example ~/.config/myapp/config.yaml
nano ~/.config/myapp/config.yaml
./myapp
```

Windows:

```powershell
New-Item -ItemType Directory -Force -Path "$env:AppData\myapp"
Copy-Item config.yaml.example "$env:AppData\myapp\config.yaml"
notepad "$env:AppData\myapp\config.yaml"
.\myapp.exe
```

## Advanced Usage

### Custom Flag Set

```go
fs := pflag.NewFlagSet("myapp", pflag.ContinueOnError)
fs.String("db-host", "", "Database host")

loader, err := NewLoader(
    WithConfigFile("config.yaml"),
    WithFlagSet(fs),
)
```

### Direct Access

```go
loader.Load()
host := loader.String("database.host")
port := loader.Int("database.port")
```

### Multiple Config Files

```go
loader, _ := NewLoader(WithConfigFile("base.yaml"))
loader.Load()
loader.Koanf().Load(file.Provider("override.yaml"), yaml.Parser())
```

### Testing

```go
func TestMyFeature(t *testing.T) {
    loader, _ := config.NewLoader()
    loader.Set("app.debug", true)
    loader.Set("server.port", 8080)
    loader.Load()
    
    var cfg config.Config
    loader.Unmarshal(&cfg)
}
```

## Key Concepts

Separation of Concerns:

- `config.go`: Generic loading mechanism (universal, never changes)
- `appConfig.go`: App-specific structure (customize per project)

Type Safety: All configuration is type-safe with proper Go types (int, bool, time.Duration, etc.)

Validation: Configuration is validated on load, catching errors early before the app starts.

## Troubleshooting

Flags not working:

- Ensure `fs.Parse(os.Args[1:])` is called
- Check flag names use dots: `server.port`

Environment variables not loading:

- Check prefix matches (`APP` by default)
- Use underscores: `APP_SERVER_PORT`
- Pattern: `APP_SECTION_KEY`

Config file not loading:

- Check file extension (`.yaml`, `.json`, `.toml`, `.env`)
- Verify file path
- Config file is optional (no error if missing)

Values not overriding:

- Remember priority: CLI > ENV > File > Defaults

## Best Practices

For production apps:

- Use OS-specific config directories
- Provide example config files
- Document config file location
- Support `--config` flag override
- Validate all configuration
- Use environment variables for secrets
- Don't commit production configs to version control
- Log which config file was loaded (debug level)
- Auto-create config directory with proper permissions

Security:

- Never commit passwords or secrets
- Use environment variables or secret management tools
- Provide `.yaml.example` files with placeholders
- Add actual config files to `.gitignore`

## References

- [koanf documentation](https://github.com/knadh/koanf)
- [12-Factor App Config](https://12factor.net/config)
- [pflag documentation](https://github.com/spf13/pflag)
