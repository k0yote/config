# GSM - Google Secret Manager Configuration Library

[![Go Reference](https://pkg.go.dev/badge/github.com/k0yote/config/gsm.svg)](https://pkg.go.dev/github.com/k0yote/config/gsm)
[![Go Report Card](https://goreportcard.com/badge/github.com/k0yote/config/gsm)](https://goreportcard.com/report/github.com/k0yote/config/gsm)

A simple and flexible Go library for managing application configuration using Google Cloud Secret Manager with environment variable fallback.

## Installation

```bash
go get github.com/k0yote/config/gsm
```

## Quick Start

```go
import "github.com/k0yote/config/gsm"

type Config struct {
    APIKey string `gsm:"API_KEY,required"`
    DBHost string `gsm:"DB_HOST,default=localhost"`
    DBPort int    `gsm:"DB_PORT,default=5432"`
}

client, _ := gsm.NewClient(ctx, "your-project-id")
defer client.Close()

loader := gsm.NewLoader(client)
var cfg Config
loader.Load(ctx, &cfg)
```

## Documentation

For full documentation, see the **[gsm package README](./gsm/README.md)** or visit [pkg.go.dev](https://pkg.go.dev/github.com/k0yote/config/gsm).

## Features

- 🔐 Google Cloud Secret Manager integration
- 🌍 Environment variable fallback
- 🏷️ Struct tag-based configuration
- 🎯 Configurable priority: env vars → Secret Manager → defaults
- 🪶 Minimal dependencies
- 📦 Production ready

## License

MIT License
