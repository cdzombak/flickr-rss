# flickr-rss

Generate RSS feeds for the latest pictures from Flickr users.

## Features

- Generate RSS feeds containing the latest photos from any Flickr user
- Support for both Flickr usernames and user IDs
- Each RSS item includes:
  - Embedded image tag showing the photo
  - RSS enclosure pointing to the photo file
  - Photo title and description
  - Publication date
- Optional OAuth authentication for accessing private/friend photos
- YAML-based credential management
- Command-line interface with flexible output options

## Installation

### Build from source

```bash
git clone https://github.com/cdzombak/flickr-rss.git
cd flickr-rss
make build
```

## Usage

### Basic Usage (Public Photos Only)

For public photos, you only need a Flickr API key:

```bash
# Using username
flickr-rss generate --api-key YOUR_API_KEY --api-secret YOUR_API_SECRET username

# Using user ID  
flickr-rss generate --api-key YOUR_API_KEY --api-secret YOUR_API_SECRET 12345678@N00

# Save to file
flickr-rss generate --api-key YOUR_API_KEY --api-secret YOUR_API_SECRET username -o feed.xml
```

### Using Credentials File

Create a credentials file to avoid passing flags each time:

```yaml
# creds.yml
api_key: your_api_key_here
api_secret: your_api_secret_here
oauth_token: your_oauth_token_here          # optional
oauth_token_secret: your_oauth_token_secret_here  # optional
```

Then use it:

```bash
flickr-rss generate -c creds.yml username
```

### OAuth Authentication (For Private/Friend Photos)

To access private photos or photos from friends/family, you'll need OAuth credentials:

```bash
# Authenticate and save credentials
flickr-rss auth --api-key YOUR_API_KEY --api-secret YOUR_API_SECRET --save-creds creds.yml

# Use authenticated credentials
flickr-rss generate -c creds.yml username
```

## Getting Flickr API Credentials

1. Visit [Flickr App Garden](https://www.flickr.com/services/apps/create/)
2. Apply for a non-commercial API key
3. Note down your API Key and API Secret

## Command Reference

### Global Flags

- `--api-key`: Flickr API key
- `--api-secret`: Flickr API secret  
- `--oauth-token`: OAuth token (for authenticated requests)
- `--oauth-token-secret`: OAuth token secret (for authenticated requests)
- `-c, --creds-file`: Path to YAML credentials file
- `-o, --output`: Output file (default: stdout)
- `-v, --verbose`: Verbose output

### Commands

#### `generate <username|userid>`

Generate RSS feed for a Flickr user.

Examples:
```bash
flickr-rss generate myusername
flickr-rss generate 12345678@N00
flickr-rss generate -c creds.yml myusername -o feed.xml
```

#### `auth`

Authenticate with Flickr OAuth and optionally save credentials.

Examples:
```bash
flickr-rss auth --api-key KEY --api-secret SECRET --save-creds creds.yml
```

## RSS Feed Format

The generated RSS feed includes:

- **Channel metadata**: Title, description, link to user's Flickr profile
- **Items per photo**:
  - Title: Photo title
  - Link: Link to photo on Flickr
  - Description: HTML with embedded `<img>` tag and photo description
  - Publication date: When photo was taken
  - GUID: Unique photo ID
  - Enclosure: Direct link to photo file

## Example RSS Item

```xml
<item>
  <title>Sunset at the Beach</title>
  <link>https://www.flickr.com/photos/username/123456789/</link>
  <description><![CDATA[<img src="https://farm1.staticflickr.com/123/123456789_abc123_m.jpg" alt="Sunset at the Beach" /><br/><br/>Beautiful sunset captured at Malibu Beach]]></description>
  <pubDate>Mon, 15 Jul 2024 18:30:00 +0000</pubDate>
  <guid>123456789</guid>
  <enclosure url="https://farm1.staticflickr.com/123/123456789_abc123_l.jpg" type="image/jpeg" length="0" />
</item>
```

## License

GNU General Public License v3.0