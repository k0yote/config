# GSM - Google Secret Manager Configuration Library

A simple and flexible Go library for managing application configuration using Google Cloud Secret Manager with environment variable fallback.

## Features

- 🔐 **Google Cloud Secret Manager Integration** - Seamlessly fetch secrets from GCP Secret Manager
- 🌍 **Environment Variable Support** - Automatic fallback to environment variables
- 🏷️ **Struct Tag-Based Configuration** - Load configuration using simple struct tags
- 🎯 **Flexible Resolution** - Configurable priority: env vars → Secret Manager → defaults
- 🪶 **Minimal Dependencies** - Only depends on the GCP Secret Manager SDK
- 📦 **Production Ready** - Designed for pkg.go.dev distribution

## Installation

```bash
go get github.com/k0yote/config/gsm
```

## Quick Start

### Basic Usage with Struct Tags

```go
package main

import (
    "context"
    "log"
    "github.com/k0yote/config/gsm"
)

type Config struct {
    APIKey   string `gsm:"API_KEY,required"`
    DBHost   string `gsm:"DB_HOST,default=localhost"`
    DBPort   int    `gsm:"DB_PORT,default=5432"`
    Debug    bool   `gsm:"DEBUG,default=false"`
    Features []string `gsm:"FEATURES"`
}

func main() {
    ctx := context.Background()

    // Create client
    client, err := gsm.NewClient(ctx, "your-gcp-project-id")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Load configuration
    loader := gsm.NewLoader(client)
    var cfg Config
    if err := loader.Load(ctx, &cfg); err != nil {
        log.Fatal(err)
    }

    // Use configuration
    log.Printf("API Key: %s", cfg.APIKey)
}
```

### Direct Value Resolution

```go
// Create resolver (can use nil for client if only using env vars)
resolver := gsm.NewResolver(client)

// Resolve with format: "sm://SECRET_NAME||default_value"
value, err := resolver.Resolve(ctx, "sm://API_KEY||default-key")

// Resolve arrays
values, err := resolver.ResolveSlice(ctx, []string{"sm://ALLOWED_HOSTS"})
```

### Without Secret Manager (Environment Variables Only)

```go
// Create loader without Secret Manager
loader := gsm.NewLoader(nil, gsm.WithSecretManagerEnabled(false))

var cfg Config
if err := loader.Load(ctx, &cfg); err != nil {
    log.Fatal(err)
}
```

## Configuration Format

### Secret Reference Format

Use the `sm://` prefix to indicate secret references:

```
sm://SECRET_NAME||default_value
```

- `SECRET_NAME`: Name of the environment variable or Secret Manager secret
- `default_value`: Fallback value if not found (optional)

### Resolution Priority

Values are resolved in this order:

1. **Environment Variables** - Checked first
2. **Google Cloud Secret Manager** - If enabled and env var not found
3. **Default Value** - From the configuration if provided

### Struct Tags

Tag format: `` `gsm:"SECRET_NAME,option1,option2"` ``

**Options:**
- `default=VALUE` - Default value if not found
- `required` - Returns error if value is not found
- `-` - Skip this field

**Supported Types:**
- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool`
- `[]string`

### Array Values

Environment variables can contain arrays in two formats:

**JSON format:**
```bash
export ALLOWED_HOSTS='["host1.com", "host2.com"]'
```

**CSV format:**
```bash
export ALLOWED_HOSTS="host1.com,host2.com"
```

## Configuration Options

### Loader Options

```go
loader := gsm.NewLoader(client,
    gsm.WithEnvPrefix("APP_"),              // Add prefix to env var lookups
    gsm.WithSecretManagerEnabled(true),     // Enable/disable Secret Manager
)
```

### WithEnvPrefix

Adds a prefix to all environment variable lookups:

```go
// With prefix "APP_", looking up "DB_HOST" checks "APP_DB_HOST"
loader := gsm.NewLoader(client, gsm.WithEnvPrefix("APP_"))
```

### WithSecretManagerEnabled

Control whether Secret Manager is used:

```go
// Disable Secret Manager (only use env vars and defaults)
loader := gsm.NewLoader(nil, gsm.WithSecretManagerEnabled(false))
```

## Examples

See the [examples](./examples/basic/main.go) directory for more comprehensive examples.

## Authentication

The library uses Google Application Default Credentials (ADC). Set up authentication by either:

1. **Service Account Key:**
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
   ```

2. **Run in GCP environment** (GCE, Cloud Run, GKE, etc.) where ADC is automatically available

3. **Use `gcloud` CLI:**
   ```bash
   gcloud auth application-default login
   ```

## Error Handling

The library provides specific error types for better error handling:

```go
import "errors"

if err := loader.Load(ctx, &cfg); err != nil {
    var reqErr *gsm.RequiredFieldError
    if errors.As(err, &reqErr) {
        log.Printf("Required field missing: %s", reqErr.FieldName)
    }

    var notFoundErr *gsm.SecretNotFoundError
    if errors.As(err, &notFoundErr) {
        log.Printf("Secret not found: %s", notFoundErr.SecretName)
    }
}
```

**Error Types:**
- `ErrSecretNotFound` - Secret not found and no default provided
- `ErrInvalidTarget` - Invalid target for Load() (must be pointer to struct)
- `ErrRequiredFieldMissing` - Required field has no value
- `ErrInvalidFormat` - Invalid secret reference format
- `ErrUnsupportedType` - Unsupported field type

## Best Practices

1. **Use struct tags for complex configurations** - More maintainable than manual resolution
2. **Mark critical fields as required** - Fail fast if essential config is missing
3. **Provide sensible defaults** - For non-critical configuration
4. **Use environment variables for local development** - Keep Secret Manager for production
5. **Close the client when done** - Always `defer client.Close()`

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestResolve
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
