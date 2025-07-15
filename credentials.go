package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Credentials struct {
	APIKey          string `yaml:"api_key"`
	APISecret       string `yaml:"api_secret"`
	OAuthToken      string `yaml:"oauth_token"`
	OAuthTokenSecret string `yaml:"oauth_token_secret"`
}

func loadCredsIfProvided() (*Credentials, error) {
	if credsFile == "" {
		return &Credentials{
			APIKey:          apiKey,
			APISecret:       apiSecret,
			OAuthToken:      oauthToken,
			OAuthTokenSecret: oauthSecret,
		}, nil
	}

	data, err := os.ReadFile(credsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read credentials file %s: %w", credsFile, err)
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials file %s: %w", credsFile, err)
	}

	// Override with CLI flags if provided
	if apiKey != "" {
		creds.APIKey = apiKey
	}
	if apiSecret != "" {
		creds.APISecret = apiSecret
	}
	if oauthToken != "" {
		creds.OAuthToken = oauthToken
	}
	if oauthSecret != "" {
		creds.OAuthTokenSecret = oauthSecret
	}

	return &creds, nil
}

func saveCredentials(creds *Credentials, filename string) error {
	data, err := yaml.Marshal(creds)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials file %s: %w", filename, err)
	}

	return nil
}

func (c *Credentials) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	if c.APISecret == "" {
		return fmt.Errorf("API secret is required")
	}
	return nil
}

func (c *Credentials) HasOAuth() bool {
	return c.OAuthToken != "" && c.OAuthTokenSecret != ""
}