# flickr-rss

Generate an RSS feed of a Flickr user's photostream or your Flickr Friends & Family feed.

## Features

- **User feeds**: Generate RSS feeds from any Flickr user's photos
  - **Public photos only** when using API key alone
  - **Includes friends/family photos** if you're in their network (requires OAuth)
- **Friends & family feeds**: Generate feeds from your friends & family timeline (requires OAuth)
- **Clean, high-res output:** output RSS items contain the Large-size image only; the image is also attached as an RSS Enclosure
- Output to stdout or save to file

## Usage

### User Feeds

Generate RSS feeds from any Flickr user's photos:

```bash
# Public photos only (API key sufficient)
flickr-rss generate username -c creds.yml

# Include friends/family photos (if you're their friend/family; requires OAuth authentication)
flickr-rss generate username -c creds-with-oauth.yml

# Using user ID or profile URL
flickr-rss generate 12345678@N00 -c creds.yml
flickr-rss generate "https://www.flickr.com/photos/username/" -c creds.yml

# Specify number of photos (default: 20)
flickr-rss generate username --count 100 -c creds.yml

# Save to file
flickr-rss generate username -c creds.yml -o feed.xml
```

**Note**: If you're authenticated (have OAuth tokens) and are friends/family with the user, their private photos shared with you will be included in the feed. Without authentication, or if you're not in the target user's friends/family, only public photos are included.

### Friends & Family Feeds

Generate feeds from your friends & family timeline (requires OAuth authentication):

```bash
# Friends & family feed
flickr-rss generate -ff -c creds.yml

# Specify count for friends & family (max 50 photos)
flickr-rss generate -ff --count 30 -c creds.yml
```

### Authentication

1. **Get API credentials**: Visit [Flickr App Garden](https://www.flickr.com/services/apps/create/) and create a non-commercial API key

2. **Create credentials file**, then stop here if you don't need OAuth creds:
```yaml
# creds.yml
api_key: your_api_key_here
api_secret: your_api_secret_here
```

3. **Authenticate via OAuth**:
```bash
# Authenticate via OAuth to get your OAuth credentials:
flickr-rss auth --api-key YOUR_API_KEY --api-secret YOUR_API_SECRET --save-creds creds.yml
```

### Reference

```
flickr-rss generate [username|userid|profile_url]
```

Generate RSS feed for a Flickr user or friends & family timeline.

**Flags:**
- `-ff, --friends-family`: Generate friends & family feed instead of user feed
- `--count`: Number of photos to include (default: 20, max 50 for friends & family)
- `-c, --creds-file`: Path to YAML credentials file
- `-o, --output`: Output file (default: stdout)
- `-v, --verbose`: Verbose output

```
flickr-rss auth
```

Authenticate with Flickr OAuth and save credentials.

**Flags:**
- `--api-key`: Flickr API key
- `--api-secret`: Flickr API secret
- `--save-creds`: Save credentials to specified YAML file

## Installation

### macOS via Homebrew

```shell
brew install cdzombak/oss/flickr-rss
```

### Debian via apt repository

[Install my Debian repository](https://www.dzombak.com/blog/2025/06/updated-instructions-for-installing-my-debian-package-repositories/) if you haven't already:

```shell
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://dist.cdzombak.net/keys/dist-cdzombak-net.gpg -o /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo chmod 644 /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo mkdir -p /etc/apt/sources.list.d
sudo curl -fsSL https://dist.cdzombak.net/cdzombak-oss.sources -o /etc/apt/sources.list.d/cdzombak-oss.sources
sudo chmod 644 /etc/apt/sources.list.d/cdzombak-oss.sources
sudo apt update
```

Then install `flickr-rss` via `apt-get`:

```shell
sudo apt-get install flickr-rss
```

### Manual installation from build artifacts

Pre-built binaries for Linux and macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/flickr-rss/releases). Debian packages for each release are available as well.

### Build and install locally

```shell
git clone https://github.com/cdzombak/flickr-rss.git
cd flickr-rss
make build

cp out/flickr-rss $INSTALL_DIR
```

## Docker images

Docker images are available for a variety of Linux architectures from [Docker Hub](https://hub.docker.com/r/cdzombak/flickr-rss) and [GHCR](https://github.com/cdzombak/dirshard/pkgs/container/flickr-rss). Images are based on the `scratch` image and are as small as possible.

Run them via, for example:

```shell
docker run --rm cdzombak/flickr-rss:1 [OPTIONS]
docker run --rm ghcr.io/cdzombak/flickr-rss:1 [OPTIONS]

## License

GNU General Public License v3.0; see [LICENSE](LICENSE) in this repository.

## Author

[Claude Code](https://www.anthropic.com/claude-code) wrote this code with management by Chris Dzombak ([dzombak.com](https://www.dzombak.com) / [github.com/cdzombak](https://www.github.com/cdzombak)).

<br /><br />![I believe in RSS](i%20believe%20in%20rss.png)
