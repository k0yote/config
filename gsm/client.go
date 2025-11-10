package gsm

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// Client provides access to Google Cloud Secret Manager.
type Client struct {
	projectID string
	client    *secretmanager.Client
}

// NewClient creates a new Secret Manager client for the given GCP project.
// The client uses Application Default Credentials (ADC) for authentication.
//
// Make sure to set GOOGLE_APPLICATION_CREDENTIALS environment variable
// or run in an environment with default credentials (GCE, Cloud Run, etc).
func NewClient(ctx context.Context, projectID string) (*Client, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID cannot be empty")
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	return &Client{
		projectID: projectID,
		client:    client,
	}, nil
}

// Close closes the Secret Manager client and releases resources.
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// GetSecret retrieves a secret value from Google Cloud Secret Manager.
// The secretName should be the name of the secret (not the full resource path).
// It always fetches the latest version of the secret.
//
// Returns ErrSecretNotFound if the secret doesn't exist or cannot be accessed.
func (c *Client) GetSecret(ctx context.Context, secretName string) (string, error) {
	if secretName == "" {
		return "", fmt.Errorf("secretName cannot be empty")
	}

	// Build the resource name for the latest version
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", c.projectID, secretName)

	// Access the secret version
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := c.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", &SecretNotFoundError{SecretName: secretName}
	}

	return string(result.Payload.Data), nil
}

// ProjectID returns the GCP project ID associated with this client.
func (c *Client) ProjectID() string {
	return c.projectID
}
