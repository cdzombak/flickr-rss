package main

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	flickrRequestTokenURL = "https://www.flickr.com/services/oauth/request_token"
	flickrAuthorizeURL    = "https://www.flickr.com/services/oauth/authorize"
	flickrAccessTokenURL  = "https://www.flickr.com/services/oauth/access_token"
)

type OAuthClient struct {
	apiKey       string
	apiSecret    string
	requestToken string
	tokenSecret  string
}

func NewOAuthClient(apiKey, apiSecret string) *OAuthClient {
	return &OAuthClient{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

func (c *OAuthClient) GetRequestToken() error {
	params := map[string]string{
		"oauth_callback":         "oob",
		"oauth_consumer_key":     c.apiKey,
		"oauth_nonce":           generateNonce(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":       strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_version":         "1.0",
	}

	signature := c.generateSignature("GET", flickrRequestTokenURL, params, "")
	params["oauth_signature"] = signature

	authHeader := c.buildAuthHeader(params)

	client := &http.Client{}
	req, err := http.NewRequest("GET", flickrRequestTokenURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		return fmt.Errorf("failed to read response: %w", err)
	}

	responseData := string(body[:n])
	values, err := url.ParseQuery(responseData)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	c.requestToken = values.Get("oauth_token")
	c.tokenSecret = values.Get("oauth_token_secret")

	if c.requestToken == "" || c.tokenSecret == "" {
		return fmt.Errorf("failed to get request token from response: %s", responseData)
	}

	return nil
}

func (c *OAuthClient) GetAuthorizationURL() string {
	return fmt.Sprintf("%s?oauth_token=%s&perms=read", flickrAuthorizeURL, c.requestToken)
}

func (c *OAuthClient) GetAccessToken(verifier string) (*Credentials, error) {
	params := map[string]string{
		"oauth_consumer_key":     c.apiKey,
		"oauth_nonce":           generateNonce(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":       strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_token":           c.requestToken,
		"oauth_verifier":        verifier,
		"oauth_version":         "1.0",
	}

	signature := c.generateSignature("GET", flickrAccessTokenURL, params, c.tokenSecret)
	params["oauth_signature"] = signature

	authHeader := c.buildAuthHeader(params)

	client := &http.Client{}
	req, err := http.NewRequest("GET", flickrAccessTokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	body := make([]byte, 1024)
	n, err := resp.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	responseData := string(body[:n])
	values, err := url.ParseQuery(responseData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	accessToken := values.Get("oauth_token")
	accessTokenSecret := values.Get("oauth_token_secret")

	if accessToken == "" || accessTokenSecret == "" {
		return nil, fmt.Errorf("failed to get access token from response: %s", responseData)
	}

	return &Credentials{
		APIKey:          c.apiKey,
		APISecret:       c.apiSecret,
		OAuthToken:      accessToken,
		OAuthTokenSecret: accessTokenSecret,
	}, nil
}

func (c *OAuthClient) generateSignature(method, url string, params map[string]string, tokenSecret string) string {
	// Sort parameters
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build parameter string
	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", percentEncode(k), percentEncode(params[k])))
	}
	paramString := strings.Join(paramPairs, "&")

	// Build signature base string
	baseString := fmt.Sprintf("%s&%s&%s", method, percentEncode(url), percentEncode(paramString))

	// Build signing key
	signingKey := fmt.Sprintf("%s&%s", percentEncode(c.apiSecret), percentEncode(tokenSecret))

	// Generate HMAC-SHA1 signature
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(baseString))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

func (c *OAuthClient) buildAuthHeader(params map[string]string) string {
	var authParams []string
	for k, v := range params {
		if strings.HasPrefix(k, "oauth_") {
			authParams = append(authParams, fmt.Sprintf(`%s="%s"`, k, percentEncode(v)))
		}
	}
	return "OAuth " + strings.Join(authParams, ", ")
}

func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func percentEncode(s string) string {
	return url.QueryEscape(s)
}

func performOAuthFlow(apiKey, apiSecret string) (*Credentials, error) {
	client := NewOAuthClient(apiKey, apiSecret)

	fmt.Println("Step 1: Getting request token...")
	if err := client.GetRequestToken(); err != nil {
		return nil, fmt.Errorf("failed to get request token: %w", err)
	}

	authURL := client.GetAuthorizationURL()
	fmt.Printf("\nStep 2: Please visit this URL to authorize the application:\n%s\n\n", authURL)
	fmt.Print("After authorizing, enter the verification code: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	verifier := strings.TrimSpace(scanner.Text())

	if verifier == "" {
		return nil, fmt.Errorf("verification code is required")
	}

	fmt.Println("\nStep 3: Getting access token...")
	creds, err := client.GetAccessToken(verifier)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	fmt.Println("Authentication successful!")
	return creds, nil
}