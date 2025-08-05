package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	ec "github.com/cdzombak/exitcode_go"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "flickr-rss",
		Short: "Generate RSS feeds for Flickr user photos",
		Long: `flickr-rss generates RSS feeds containing the latest photos from a Flickr user.
Each feed item includes an embedded image and RSS enclosure for the photo.`,
	}

	generateCmd = &cobra.Command{
		Use:   "generate [username|userid|profile_url]",
		Short: "Generate RSS feed for a Flickr user or friends & family",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runGenerate,
	}

	authCmd = &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Flickr and save credentials",
		RunE:  runAuth,
	}

	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information and exit",
		RunE:  runVersion,
	}

	// Global flags
	apiKey        string
	apiSecret     string
	oauthToken    string
	oauthSecret   string
	credsFile     string
	output        string
	verbose       bool
	saveCreds     string
	friendsFamily bool
	photoCount    int

	// injected at build time:
	version string = "<dev>"
)

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(versionCmd)

	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Flickr API key")
	rootCmd.PersistentFlags().StringVar(&apiSecret, "api-secret", "", "Flickr API secret")
	rootCmd.PersistentFlags().StringVar(&oauthToken, "oauth-token", "", "OAuth token")
	rootCmd.PersistentFlags().StringVar(&oauthSecret, "oauth-token-secret", "", "OAuth token secret")
	rootCmd.PersistentFlags().StringVarP(&credsFile, "creds-file", "c", "", "Path to credentials YAML file")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output file for RSS feed (default: stdout)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Auth command specific flags
	authCmd.Flags().StringVar(&saveCreds, "save-creds", "", "Save credentials to specified YAML file")

	// Generate command specific flags
	generateCmd.Flags().BoolVar(&friendsFamily, "ff", false, "Generate feed from friends & family photos (requires OAuth)")
	generateCmd.Flags().IntVar(&photoCount, "count", 20, "Number of photos to include in the feed")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)

		// Set exit code based on error type
		var exitCode int
		switch {
		case errors.Is(err, ErrFlickrAuth):
			exitCode = ec.NoPermission
		case errors.Is(err, ErrFlickrServer):
			exitCode = ec.Unavailable
		case errors.Is(err, ErrFlickrUsage):
			exitCode = ec.Usage
		case errors.Is(err, ErrFlickrAPI):
			exitCode = ec.Failure
		case errors.Is(err, ErrFileIO):
			exitCode = ec.IOErr
		case errors.Is(err, ErrInputs):
			exitCode = ec.NotConfigured
		case errors.Is(err, ErrUsage):
			exitCode = ec.InvalidArgument
		default:
			exitCode = ec.Failure
		}

		os.Exit(exitCode)
	}
}

func runVersion(_ *cobra.Command, _ []string) error {
	fmt.Printf("flickr-rss %s\n", version)
	return nil
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Handle friends & family mode
	if friendsFamily {
		return runGenerateFriendsFamily(cmd, args)
	}

	if len(args) == 0 {
		return NewUsage("username, user ID, or profile URL is required unless using -ff flag")
	}

	userInput := args[0]

	// Load credentials
	creds, err := loadCredsIfProvided()
	if err != nil {
		return WrapInputs(err, "failed to load credentials")
	}

	if err := creds.Validate(); err != nil {
		return WrapInputs(err, "invalid credentials")
	}

	client := NewFlickrClient(creds)

	// Determine if userInput is a profile URL, username, or user ID
	var userID string
	var displayName string

	if verbose {
		fmt.Fprintf(os.Stderr, "Looking up user: %s\n", userInput)
	}

	// Check if userInput is a Flickr profile URL
	if isFlickrProfileURL(userInput) {
		if verbose {
			fmt.Fprintf(os.Stderr, "Detected Flickr profile URL, looking up user\n")
		}
		userID, err = client.LookupUserByURL(userInput)
		if err != nil {
			return WrapFlickrAPI(err, fmt.Sprintf("failed to lookup user from URL '%s'", userInput))
		}
		// Get the actual username for display purposes
		displayName, err = client.GetUserInfo(userID)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to get username, using user ID: %v\n", err)
			}
			displayName = userID
		}
	} else if containsNonNumeric(userInput) {
		// Try to find user by username (if it contains non-numeric characters, likely a username)
		userID, err = client.FindUserByUsername(userInput)
		if err != nil {
			return WrapFlickrAPI(err, fmt.Sprintf("failed to find user by username '%s'", userInput))
		}
		displayName = userInput
	} else {
		// Assume it's already a user ID
		userID = userInput
		displayName = userInput
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Using user ID: %s\n", userID)
		fmt.Fprintf(os.Stderr, "Display name: %s\n", displayName)
	}

	// Fetch latest photos
	photos, err := client.GetUserPhotos(userID, photoCount)
	if err != nil {
		return WrapFlickrAPI(err, fmt.Sprintf("failed to fetch photos for user %s", userID))
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Found %d photos\n", len(photos))
	}

	// Generate RSS feed
	feed := GenerateRSSFeed(photos, displayName)

	// Output RSS feed
	var writer io.Writer = os.Stdout
	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			return WrapFileIO(err, fmt.Sprintf("failed to create output file %s", output))
		}
		defer file.Close()
		writer = file

		if verbose {
			fmt.Fprintf(os.Stderr, "Writing RSS feed to: %s\n", output)
		}
	}

	return feed.WriteXML(writer)
}

func containsNonNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return true
		}
	}
	return false
}

func isFlickrProfileURL(urlStr string) bool {
	re := regexp.MustCompile(`^https?://(?:www\.)?flickr\.com/photos/[^/]+/?`)
	return re.MatchString(urlStr)
}

func runAuth(cmd *cobra.Command, args []string) error {
	if apiKey == "" || apiSecret == "" {
		return NewUsage("API key and secret are required for authentication. Use --api-key and --api-secret flags")
	}

	fmt.Println("Starting OAuth authentication...")
	creds, err := performOAuthFlow(apiKey, apiSecret)
	if err != nil {
		return WrapFlickrAuth(err, "authentication failed")
	}

	if saveCreds != "" {
		if err := saveCredentials(creds, saveCreds); err != nil {
			return WrapFileIO(err, "failed to save credentials")
		}
		fmt.Printf("Credentials saved to: %s\n", saveCreds)
	} else {
		fmt.Println("\nCredentials (save these for future use):")
		fmt.Printf("API Key: %s\n", creds.APIKey)
		fmt.Printf("API Secret: %s\n", creds.APISecret)
		fmt.Printf("OAuth Token: %s\n", creds.OAuthToken)
		fmt.Printf("OAuth Token Secret: %s\n", creds.OAuthTokenSecret)
	}

	return nil
}

func runGenerateFriendsFamily(cmd *cobra.Command, args []string) error {
	// Load credentials
	creds, err := loadCredsIfProvided()
	if err != nil {
		return WrapInputs(err, "failed to load credentials")
	}

	if err := creds.Validate(); err != nil {
		return WrapInputs(err, "invalid credentials")
	}

	// Verify OAuth credentials are present for friends & family access
	if creds.OAuthToken == "" || creds.OAuthTokenSecret == "" {
		return NewUsage("friends & family feed requires OAuth authentication. Run 'flickr-rss auth' first")
	}

	client := NewFlickrClient(creds)

	if verbose {
		fmt.Fprintf(os.Stderr, "Fetching friends & family photos...\n")
	}

	// Fetch latest photos from friends & family (max 50 due to API limits)
	requestCount := photoCount
	if requestCount > 50 {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: Friends & family feed limited to 50 photos (requested %d)\n", photoCount)
		}
		requestCount = 50
	}
	photos, err := client.GetContactsPhotos(requestCount)
	if err != nil {
		return WrapFlickrAPI(err, "failed to fetch friends & family photos")
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Found %d photos from friends & family\n", len(photos))
	}

	// Generate RSS feed
	feed := GenerateRSSFeed(photos, "Friends & Family")

	// Output RSS feed
	var writer io.Writer = os.Stdout
	if output != "" {
		file, err := os.Create(output)
		if err != nil {
			return WrapFileIO(err, fmt.Sprintf("failed to create output file %s", output))
		}
		defer file.Close()
		writer = file

		if verbose {
			fmt.Fprintf(os.Stderr, "Writing RSS feed to: %s\n", output)
		}
	}

	return feed.WriteXML(writer)
}
