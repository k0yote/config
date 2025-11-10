package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/k0yote/config/gsm"
)

// Config demonstrates various field types and tag options
type Config struct {
	// Required field - will error if not found
	APIKey string `gsm:"API_KEY,required"`

	// Optional with default value
	DBHost string `gsm:"DB_HOST,default=localhost"`
	DBPort int    `gsm:"DB_PORT,default=5432"`

	// Boolean with default
	Debug bool `gsm:"DEBUG,default=false"`

	// Slice of strings
	AllowedHosts []string `gsm:"ALLOWED_HOSTS,default=localhost,127.0.0.1"`

	// Optional without default (will be empty string if not found)
	OptionalKey string `gsm:"OPTIONAL_KEY"`

	// Ignored field
	Internal string `gsm:"-"`
}

func main() {
	ctx := context.Background()

	// Example 1: Using environment variables only (no Secret Manager)
	fmt.Println("=== Example 1: Environment Variables Only ===")
	exampleWithEnvOnly(ctx)

	// Example 2: Using Secret Manager
	fmt.Println("\n=== Example 2: With Secret Manager ===")
	exampleWithSecretManager(ctx)

	// Example 3: Direct value resolution
	fmt.Println("\n=== Example 3: Direct Resolution ===")
	exampleDirectResolution(ctx)
}

func exampleWithEnvOnly(ctx context.Context) {
	// Set some environment variables
	os.Setenv("API_KEY", "my-api-key-from-env")
	os.Setenv("DB_HOST", "postgres.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DEBUG", "true")
	os.Setenv("ALLOWED_HOSTS", `["api.example.com", "www.example.com"]`)

	// Create loader without Secret Manager
	loader := gsm.NewLoader(nil, gsm.WithSecretManagerEnabled(false))

	var cfg Config
	if err := loader.Load(ctx, &cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("APIKey: %s\n", cfg.APIKey)
	fmt.Printf("DBHost: %s\n", cfg.DBHost)
	fmt.Printf("DBPort: %d\n", cfg.DBPort)
	fmt.Printf("Debug: %v\n", cfg.Debug)
	fmt.Printf("AllowedHosts: %v\n", cfg.AllowedHosts)
	fmt.Printf("OptionalKey: %s\n", cfg.OptionalKey)
}

func exampleWithSecretManager(ctx context.Context) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		fmt.Println("Skipping: GCP_PROJECT_ID not set")
		return
	}

	client, err := gsm.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create Secret Manager client: %v", err)
		return
	}
	defer client.Close()

	// Create loader with Secret Manager enabled
	loader := gsm.NewLoader(client)

	var cfg Config
	if err := loader.Load(ctx, &cfg); err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}

	fmt.Printf("APIKey: %s\n", cfg.APIKey)
	fmt.Printf("DBHost: %s\n", cfg.DBHost)
	fmt.Printf("DBPort: %d\n", cfg.DBPort)
}

func exampleDirectResolution(ctx context.Context) {
	// Set an environment variable
	os.Setenv("MY_SECRET", "secret-from-env")

	// Create resolver without Secret Manager
	resolver := gsm.NewResolver(nil, gsm.WithSecretManagerEnabled(false))

	// Resolve a value with the sm:// format
	value, err := resolver.Resolve(ctx, "sm://MY_SECRET||default-value")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Resolved value: %s\n", value)

	// Resolve a value that doesn't exist (will use default)
	value2, err := resolver.Resolve(ctx, "sm://NON_EXISTENT||fallback-value")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Resolved with default: %s\n", value2)

	// Resolve a slice
	os.Setenv("MY_ARRAY", "value1,value2,value3")
	values, err := resolver.ResolveSlice(ctx, []string{"sm://MY_ARRAY"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Resolved slice: %v\n", values)
}
