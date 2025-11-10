# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This repository contains **gsm** (Google Secret Manager), a reusable Go library for managing application configuration using Google Cloud Secret Manager with environment variable fallback.

The library is designed for distribution via pkg.go.dev and provides a clean, minimal API for configuration management.

## Build & Run Commands

```bash
# Install dependencies
go mod download

# Run all tests
go test -v ./gsm/...

# Run tests with coverage
go test -cover ./gsm/...

# Run a specific test
go test -v -run TestLoaderLoad ./gsm/...

# Run example
go run ./gsm/examples/basic/main.go
```

## Architecture

### Package Structure

```
gsm/
├── doc.go           # Package documentation
├── errors.go        # Error types and definitions
├── parser.go        # Secret reference format parser (sm://)
├── client.go        # Secret Manager client wrapper
├── resolver.go      # Value resolution logic
├── loader.go        # Struct tag-based configuration loader
├── *_test.go        # Unit tests (70.9% coverage)
├── examples/        # Usage examples
└── README.md        # Package documentation
```

### Core Components

**client.go**: Provides `Client` which wraps the GCP Secret Manager API client. Handles authentication via Application Default Credentials (ADC) and provides a simple interface for fetching secrets.

**parser.go**: Implements the `sm://SECRET_NAME||default_value` format parser. Returns a `SecretRef` struct containing the parsed components (secret name, default value, flags).

**resolver.go**: Implements `Resolver` which orchestrates the value resolution with priority:
1. Environment variables (checked first)
2. Secret Manager (if enabled and env var not found)
3. Default values (from the parsed reference)

Supports both single values and arrays (JSON/CSV format).

**loader.go**: Implements `Loader` which uses reflection to automatically populate struct fields based on `gsm` tags. Handles type conversion for string, int, float, bool, and []string types.

**errors.go**: Defines custom error types for better error handling:
- `SecretNotFoundError` - Secret not found
- `RequiredFieldError` - Required field missing
- `InvalidFormatError` - Invalid reference format
- `UnsupportedTypeError` - Unsupported field type

### Key Design Decisions

1. **Minimal Dependencies**: Only depends on GCP Secret Manager SDK and testify for tests. No Viper, no godotenv, no logging libraries.

2. **Flexible Configuration**: Users can disable Secret Manager entirely and use only environment variables and defaults.

3. **Type Safety**: Struct tag-based approach with compile-time type checking for common types.

4. **Array Support**: Environment variables can contain JSON arrays or CSV values, parsed automatically.

5. **Error Context**: Rich error types with context for debugging.

## Usage Patterns

### Basic Struct Tag Usage

```go
type Config struct {
    APIKey string `gsm:"API_KEY,required"`
    DBHost string `gsm:"DB_HOST,default=localhost"`
    DBPort int    `gsm:"DB_PORT,default=5432"`
}

loader := gsm.NewLoader(client)
var cfg Config
loader.Load(ctx, &cfg)
```

### Direct Value Resolution

```go
resolver := gsm.NewResolver(client)
value, err := resolver.Resolve(ctx, "sm://API_KEY||default")
```

### Options

```go
loader := gsm.NewLoader(client,
    gsm.WithEnvPrefix("APP_"),
    gsm.WithSecretManagerEnabled(true),
)
```

## Testing Strategy

Tests are comprehensive and cover:
- Parser functionality (all edge cases)
- Resolver behavior (env vars, defaults, errors)
- Loader with various field types
- Array parsing (JSON and CSV formats)
- Error cases and type validation

Run tests: `go test -v ./gsm/...`

## Publishing to pkg.go.dev

1. Ensure code is pushed to GitHub
2. Create a version tag: `git tag gsm/v1.0.0 && git push origin gsm/v1.0.0`
3. Visit `https://pkg.go.dev/github.com/k0yote/config/gsm@v1.0.0` to verify
4. Documentation will be automatically generated from doc comments

## Dependencies

- `cloud.google.com/go/secretmanager` - GCP Secret Manager SDK
- `github.com/stretchr/testify` - Testing utilities (dev dependency)

Keep dependencies minimal. Avoid adding logging, config parsing, or utility libraries.
