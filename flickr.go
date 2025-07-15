package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type FlickrClient struct {
	credentials *Credentials
	httpClient  *http.Client
}

type FlickrPhoto struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description struct {
		Content string `json:"_content"`
	} `json:"description"`
	DateTaken string `json:"datetaken"`
	URL       string `json:"url_m"`
	URLLarge  string `json:"url_l"`
	Secret    string `json:"secret"`
	Server    string `json:"server"`
	Farm      int    `json:"farm"`
	Owner     string `json:"owner"`
	Username  string `json:"username"`
}

type FlickrResponse struct {
	Photos struct {
		Photo []FlickrPhoto `json:"photo"`
	} `json:"photos"`
	Stat    string `json:"stat"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewFlickrClient(creds *Credentials) *FlickrClient {
	return &FlickrClient{
		credentials: creds,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *FlickrClient) GetUserPhotos(userID string, count int) ([]FlickrPhoto, error) {
	var allPhotos []FlickrPhoto
	perPage := 500 // Maximum allowed by Flickr API
	page := 1
	
	for len(allPhotos) < count {
		// Calculate how many photos to request for this page
		remaining := count - len(allPhotos)
		if remaining > perPage {
			remaining = perPage
		}
		
		photos, hasMore, err := c.getUserPhotosPage(userID, remaining, page)
		if err != nil {
			return nil, err
		}
		
		allPhotos = append(allPhotos, photos...)
		
		// Stop if we have enough photos or no more pages
		if len(allPhotos) >= count || !hasMore || len(photos) == 0 {
			break
		}
		
		page++
	}
	
	// Trim to exact count requested
	if len(allPhotos) > count {
		allPhotos = allPhotos[:count]
	}
	
	return allPhotos, nil
}

func (c *FlickrClient) getUserPhotosPage(userID string, perPage, page int) ([]FlickrPhoto, bool, error) {
	baseURL := "https://api.flickr.com/services/rest/"
	
	params := url.Values{}
	params.Set("method", "flickr.people.getPublicPhotos")
	params.Set("api_key", c.credentials.APIKey)
	params.Set("user_id", userID)
	params.Set("format", "json")
	params.Set("nojsoncallback", "1")
	params.Set("per_page", strconv.Itoa(perPage))
	params.Set("page", strconv.Itoa(page))
	params.Set("extras", "description,date_taken,url_m,url_l,owner_name")

	reqURL := baseURL + "?" + params.Encode()
	
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return nil, false, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("failed to read response body: %w", err)
	}

	var flickrResp struct {
		Photos struct {
			Photo []FlickrPhoto `json:"photo"`
			Page  int           `json:"page"`
			Pages int           `json:"pages"`
			Total interface{}   `json:"total"`
		} `json:"photos"`
		Stat    string `json:"stat"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	
	if err := json.Unmarshal(body, &flickrResp); err != nil {
		return nil, false, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if flickrResp.Stat != "ok" {
		if flickrResp.Message != "" {
			return nil, false, fmt.Errorf("Flickr API error: %s (code %d)", flickrResp.Message, flickrResp.Code)
		}
		return nil, false, fmt.Errorf("Flickr API returned error status: %s", flickrResp.Stat)
	}

	hasMore := flickrResp.Photos.Page < flickrResp.Photos.Pages
	return flickrResp.Photos.Photo, hasMore, nil
}

func (c *FlickrClient) FindUserByUsername(username string) (string, error) {
	baseURL := "https://api.flickr.com/services/rest/"
	
	params := url.Values{}
	params.Set("method", "flickr.people.findByUsername")
	params.Set("api_key", c.credentials.APIKey)
	params.Set("username", username)
	params.Set("format", "json")
	params.Set("nojsoncallback", "1")

	reqURL := baseURL + "?" + params.Encode()
	
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		User struct {
			ID string `json:"nsid"`
		} `json:"user"`
		Stat    string `json:"stat"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if result.Stat != "ok" {
		if result.Message != "" {
			return "", fmt.Errorf("Flickr API error: %s (code %d)", result.Message, result.Code)
		}
		return "", fmt.Errorf("Flickr API returned error status: %s", result.Stat)
	}

	return result.User.ID, nil
}

func (c *FlickrClient) LookupUserByURL(profileURL string) (string, error) {
	baseURL := "https://api.flickr.com/services/rest/"
	
	params := url.Values{}
	params.Set("method", "flickr.urls.lookupUser")
	params.Set("api_key", c.credentials.APIKey)
	params.Set("url", profileURL)
	params.Set("format", "json")
	params.Set("nojsoncallback", "1")

	reqURL := baseURL + "?" + params.Encode()
	
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
		Stat    string `json:"stat"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if result.Stat != "ok" {
		if result.Message != "" {
			return "", fmt.Errorf("Flickr API error: %s (code %d)", result.Message, result.Code)
		}
		return "", fmt.Errorf("Flickr API returned error status: %s", result.Stat)
	}

	return result.User.ID, nil
}

func (c *FlickrClient) GetUserInfo(userID string) (string, error) {
	baseURL := "https://api.flickr.com/services/rest/"
	
	params := url.Values{}
	params.Set("method", "flickr.people.getInfo")
	params.Set("api_key", c.credentials.APIKey)
	params.Set("user_id", userID)
	params.Set("format", "json")
	params.Set("nojsoncallback", "1")

	reqURL := baseURL + "?" + params.Encode()
	
	resp, err := c.httpClient.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result struct {
		Person struct {
			Username struct {
				Content string `json:"_content"`
			} `json:"username"`
		} `json:"person"`
		Stat    string `json:"stat"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if result.Stat != "ok" {
		if result.Message != "" {
			return "", fmt.Errorf("Flickr API error: %s (code %d)", result.Message, result.Code)
		}
		return "", fmt.Errorf("Flickr API returned error status: %s", result.Stat)
	}

	return result.Person.Username.Content, nil
}

func (c *FlickrClient) GetContactsPhotos(count int) ([]FlickrPhoto, error) {
	// Limit to maximum supported by API
	if count > 50 {
		count = 50
	}
	
	baseURL := "https://api.flickr.com/services/rest/"
	
	oauthParams := map[string]string{
		"oauth_consumer_key":     c.credentials.APIKey,
		"oauth_nonce":           c.generateNonce(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":       strconv.FormatInt(time.Now().Unix(), 10),
		"oauth_token":           c.credentials.OAuthToken,
		"oauth_version":         "1.0",
	}

	apiParams := map[string]string{
		"method":        "flickr.photos.getContactsPhotos",
		"format":        "json",
		"nojsoncallback": "1",
		"count":         strconv.Itoa(count),
		"just_friends":  "1",
		"extras":        "description,date_taken,url_m,url_l,owner_name",
	}

	// Combine all parameters for signature
	allParams := make(map[string]string)
	for k, v := range oauthParams {
		allParams[k] = v
	}
	for k, v := range apiParams {
		allParams[k] = v
	}

	signature := c.generateSignature("GET", baseURL, allParams, c.credentials.OAuthTokenSecret)
	oauthParams["oauth_signature"] = signature

	authHeader := c.buildAuthHeader(oauthParams)

	// Build URL with API parameters only
	params := url.Values{}
	for k, v := range apiParams {
		params.Set(k, v)
	}
	reqURL := baseURL + "?" + params.Encode()
	
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", authHeader)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var flickrResp FlickrResponse
	if err := json.Unmarshal(body, &flickrResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	if flickrResp.Stat != "ok" {
		if flickrResp.Message != "" {
			return nil, fmt.Errorf("Flickr API error: %s (code %d)", flickrResp.Message, flickrResp.Code)
		}
		return nil, fmt.Errorf("Flickr API returned error status: %s", flickrResp.Stat)
	}

	return flickrResp.Photos.Photo, nil
}

func (c *FlickrClient) generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func (c *FlickrClient) generateSignature(method, baseURL string, params map[string]string, tokenSecret string) string {
	// Sort parameters
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build parameter string
	var paramPairs []string
	for _, k := range keys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", url.QueryEscape(k), url.QueryEscape(params[k])))
	}
	paramString := strings.Join(paramPairs, "&")

	// Build signature base string
	baseString := fmt.Sprintf("%s&%s&%s", method, url.QueryEscape(baseURL), url.QueryEscape(paramString))

	// Build signing key
	signingKey := fmt.Sprintf("%s&%s", url.QueryEscape(c.credentials.APISecret), url.QueryEscape(tokenSecret))

	// Generate HMAC-SHA1 signature
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(baseString))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

func (c *FlickrClient) buildAuthHeader(params map[string]string) string {
	var authParams []string
	for k, v := range params {
		if strings.HasPrefix(k, "oauth_") {
			authParams = append(authParams, fmt.Sprintf(`%s="%s"`, k, url.QueryEscape(v)))
		}
	}
	return "OAuth " + strings.Join(authParams, ", ")
}