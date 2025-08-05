package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Credentials struct {
	APIKey           string `yaml:"api_key"`
	APISecret        string `yaml:"api_secret"`
	OAuthToken       string `yaml:"oauth_token"`
	OAuthTokenSecret string `yaml:"oauth_token_secret"`
}

func loadCredsIfProvided() (*Credentials, error) {
	if credsFile == "" {
		return &Credentials{
			APIKey:           apiKey,
			APISecret:        apiSecret,
			OAuthToken:       oauthToken,
			OAuthTokenSecret: oauthSecret,
		}, nil
	}

	data, err := os.ReadFile(credsFile)
	if err != nil {
		return nil, WrapFileIO(err, fmt.Sprintf("failed to read credentials file %s", credsFile))
	}

	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, WrapInputs(err, fmt.Sprintf("failed to parse credentials file %s", credsFile))
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
		return WrapFileIO(err, "failed to marshal credentials")
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return WrapFileIO(err, fmt.Sprintf("failed to write credentials file %s", filename))
	}

	return nil
}

func (c *Credentials) Validate() error {
	if c.APIKey == "" {
		return NewInputs("API key is required")
	}
	if c.APISecret == "" {
		return NewInputs("API secret is required")
	}
	return nil
}

func (c *Credentials) HasOAuth() bool {
	return c.OAuthToken != "" && c.OAuthTokenSecret != ""
}
