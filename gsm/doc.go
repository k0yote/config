// Package gsm provides a simple and flexible way to manage application configuration
// using Google Cloud Secret Manager with environment variable fallback.
//
// # Overview
//
// This package resolves configuration values with the following priority:
//  1. Environment variables
//  2. Google Cloud Secret Manager (if enabled)
//  3. Default values
//
// # Basic Usage
//
// Create a client and resolve individual values:
//
//	ctx := context.Background()
//	client, err := gsm.NewClient(ctx, "your-project-id")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Resolve a value with format: "sm://SECRET_NAME||default_value"
//	value, err := client.Resolve(ctx, "sm://API_KEY||default_key")
//
// # Struct Tag-Based Configuration
//
// Load configuration into a struct using tags:
//
//	type Config struct {
//	    APIKey    string `gsm:"API_KEY,default=test_key"`
//	    DBHost    string `gsm:"DB_HOST,required"`
//	    DBPort    int    `gsm:"DB_PORT,default=5432"`
//	    Features  []string `gsm:"FEATURES"`
//	}
//
//	var cfg Config
//	loader := gsm.NewLoader(client)
//	if err := loader.Load(ctx, &cfg); err != nil {
//	    log.Fatal(err)
//	}
//
// # Value Format
//
// Values can be specified in the format: "sm://SECRET_NAME||default_value"
//   - SECRET_NAME: The name of the environment variable or secret in Secret Manager
//   - default_value: The value to use if the secret is not found (optional)
//
// For arrays/slices, environment variables can be provided as:
//   - JSON format: MY_VAR=["value1", "value2"]
//   - CSV format: MY_VAR=value1,value2
//
// # Options
//
// Customize the loader behavior with options:
//
//	loader := gsm.NewLoader(client,
//	    gsm.WithEnvPrefix("APP_"),           // Add prefix to all env var lookups
//	    gsm.WithSecretManagerEnabled(true),  // Enable/disable Secret Manager
//	)
//
// # Struct Tags
//
// Supported tag options:
//   - "SECRET_NAME" - The name of the environment variable/secret
//   - "default=VALUE" - Default value if not found
//   - "required" - Error if value is not found
//   - "-" - Skip this field
//
// Examples:
//
//	APIKey string `gsm:"API_KEY"`                           // Required, no default
//	DBHost string `gsm:"DB_HOST,default=localhost"`         // Optional with default
//	Port   int    `gsm:"PORT,default=8080"`                 // Int with default
//	Debug  bool   `gsm:"DEBUG,default=false"`               // Bool with default
//	Tags   []string `gsm:"TAGS,default=tag1,tag2"`          // Slice with defaults
//	Ignore string `gsm:"-"`                                 // Ignored field
package gsm
