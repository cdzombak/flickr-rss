# flickr-rss

Generate RSS feeds for Flickr photos. Create feeds from any user's public photos or your friends & family timeline.

## Features

- **User feeds**: Generate RSS feeds from any Flickr user's photos
  - **Public photos only** when using API key alone
  - **Includes friends/family photos** when authenticated and you're in their network
- **Friends & family feeds**: Generate feeds from your friends & family timeline (requires OAuth)
- **Profile URL support**: Use Flickr profile URLs directly (e.g., `https://www.flickr.com/photos/username/`)
- **High-resolution images**: Uses large-size photos for both display and enclosures
- **Customizable count**: Specify how many photos to include (supports hundreds for user feeds, up to 50 for friends & family)
- **YAML credential management**: Store API keys and OAuth tokens in a config file
- **Flexible output**: Output to stdout or save to file

## Installation

### Build from source

```bash
git clone https://github.com/cdzombak/flickr-rss.git
cd flickr-rss
make build
```

## Usage

### User Feeds

Generate RSS feeds from any Flickr user's photos:

```bash
# Public photos only (API key sufficient)
flickr-rss generate username -c creds.yml

# Include friends/family photos (requires OAuth authentication)
flickr-rss generate username -c creds-with-oauth.yml

# Using user ID or profile URL
flickr-rss generate 12345678@N00 -c creds.yml
flickr-rss generate "https://www.flickr.com/photos/username/" -c creds.yml

# Specify number of photos (default: 20)
flickr-rss generate username --count 100 -c creds.yml

# Save to file
flickr-rss generate username -c creds.yml -o feed.xml
```

**Note**: If you're authenticated (have OAuth tokens) and are friends/family with the user, their private photos shared with you will be included in the feed. Without authentication, only public photos are included.

### Friends & Family Feeds

Generate feeds from your friends & family timeline (requires OAuth authentication):

```bash
# Friends & family feed (max 50 photos)
flickr-rss generate -ff -c creds.yml

# Specify count for friends & family
flickr-rss generate -ff --count 30 -c creds.yml
```

### Authentication Setup

1. **Get API credentials**: Visit [Flickr App Garden](https://www.flickr.com/services/apps/create/) and create a non-commercial API key

2. **Create credentials file**:
```yaml
# creds.yml
api_key: your_api_key_here
api_secret: your_api_secret_here
oauth_token: your_oauth_token_here          # optional for user feeds, required for friends & family feeds
oauth_token_secret: your_oauth_token_secret_here  # optional for user feeds, required for friends & family feeds
```

3. **Authenticate for enhanced access**:
```bash
# Authenticate to access friends & family feeds and private photos in user feeds
flickr-rss auth --api-key YOUR_API_KEY --api-secret YOUR_API_SECRET --save-creds creds.yml
```

## Command Reference

### `flickr-rss generate [username|userid|profile_url]`

Generate RSS feed for a Flickr user or friends & family timeline.

**Flags:**
- `-ff, --friends-family`: Generate friends & family feed instead of user feed
- `--count`: Number of photos to include (default: 20, max 50 for friends & family)
- `-c, --creds-file`: Path to YAML credentials file
- `-o, --output`: Output file (default: stdout)
- `-v, --verbose`: Verbose output

### `flickr-rss auth`

Authenticate with Flickr OAuth and save credentials.

**Flags:**
- `--api-key`: Flickr API key
- `--api-secret`: Flickr API secret
- `--save-creds`: Save credentials to specified YAML file

## RSS Feed Format

Each RSS feed includes:
- **Channel**: Feed title, description, and link
- **Items**: One per photo with title, link, embedded image, description, date, and enclosure

## License

GNU General Public License v3.0

![I believe in RSS!](./i believe in rss.png)
