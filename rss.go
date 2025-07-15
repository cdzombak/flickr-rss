package main

import (
	"fmt"
	"html"
	"io"
	"strings"
	"time"
)

type RSSFeed struct {
	Title       string
	Link        string
	Description string
	Items       []RSSItem
}

type RSSItem struct {
	Title       string
	Link        string
	Description string
	PubDate     string
	GUID        string
	Enclosure   *RSSEnclosure
}

type RSSEnclosure struct {
	URL    string
	Type   string
	Length string
}

func GenerateRSSFeed(photos []FlickrPhoto, username string) *RSSFeed {
	feed := &RSSFeed{
		Title:       fmt.Sprintf("Flickr Photos from %s", username),
		Link:        fmt.Sprintf("https://www.flickr.com/people/%s/", username),
		Description: fmt.Sprintf("Latest photos from Flickr user %s", username),
		Items:       make([]RSSItem, 0, len(photos)),
	}

	for _, photo := range photos {
		// Use photo owner for link if available, otherwise fallback to feed username
		linkOwner := username
		if photo.Owner != "" {
			linkOwner = photo.Owner
		}
		
		item := RSSItem{
			Title:       photo.Title,
			Link:        fmt.Sprintf("https://www.flickr.com/photos/%s/%s/", linkOwner, photo.ID),
			Description: generateItemDescription(photo),
			PubDate:     formatPubDate(photo.DateTaken),
			GUID:        photo.ID,
		}

		// Add enclosure if we have a large URL
		if photo.URLLarge != "" {
			item.Enclosure = &RSSEnclosure{
				URL:    photo.URLLarge,
				Type:   "image/jpeg",
				Length: "0", // We don't know the actual length
			}
		} else if photo.URL != "" {
			item.Enclosure = &RSSEnclosure{
				URL:    photo.URL,
				Type:   "image/jpeg",
				Length: "0",
			}
		}

		feed.Items = append(feed.Items, item)
	}

	return feed
}

func generateItemDescription(photo FlickrPhoto) string {
	var desc strings.Builder
	
	// Use large image for display
	imageURL := photo.URLLarge
	if imageURL == "" {
		// Fallback to medium if large not available
		imageURL = photo.URL
		if imageURL == "" {
			// Final fallback to constructing URL from photo metadata
			imageURL = fmt.Sprintf("https://farm%d.staticflickr.com/%s/%s_%s_m.jpg", 
				photo.Farm, photo.Server, photo.ID, photo.Secret)
		}
	}
	
	desc.WriteString(fmt.Sprintf(`<img src="%s" alt="%s" />`, 
		html.EscapeString(imageURL), html.EscapeString(photo.Title)))
	
	// Add description if available
	if photo.Description.Content != "" {
		desc.WriteString("<br/><br/>")
		desc.WriteString(html.EscapeString(photo.Description.Content))
	}
	
	return desc.String()
}

func formatPubDate(dateTaken string) string {
	// Flickr returns dates in format "2023-07-15 12:34:56"
	if dateTaken == "" {
		return time.Now().Format(time.RFC1123Z)
	}
	
	// Parse the date
	t, err := time.Parse("2006-01-02 15:04:05", dateTaken)
	if err != nil {
		return time.Now().Format(time.RFC1123Z)
	}
	
	return t.Format(time.RFC1123Z)
}

func (feed *RSSFeed) WriteXML(w io.Writer) error {
	// Write XML header
	if _, err := fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
<title>%s</title>
<link>%s</link>
<description>%s</description>
<language>en-us</language>
<lastBuildDate>%s</lastBuildDate>
<atom:link href="%s" rel="self" type="application/rss+xml" />
`,
		html.EscapeString(feed.Title),
		html.EscapeString(feed.Link),
		html.EscapeString(feed.Description),
		time.Now().Format(time.RFC1123Z),
		html.EscapeString(feed.Link)); err != nil {
		return err
	}

	// Write items
	for _, item := range feed.Items {
		if _, err := fmt.Fprintf(w, `
<item>
<title>%s</title>
<link>%s</link>
<description><![CDATA[%s]]></description>
<pubDate>%s</pubDate>
<guid>%s</guid>`,
			html.EscapeString(item.Title),
			html.EscapeString(item.Link),
			item.Description, // Already HTML escaped in generateItemDescription
			item.PubDate,
			html.EscapeString(item.GUID)); err != nil {
			return err
		}

		// Add enclosure if present
		if item.Enclosure != nil {
			if _, err := fmt.Fprintf(w, `
<enclosure url="%s" type="%s" length="%s" />`,
				html.EscapeString(item.Enclosure.URL),
				html.EscapeString(item.Enclosure.Type),
				html.EscapeString(item.Enclosure.Length)); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintf(w, "\n</item>"); err != nil {
			return err
		}
	}

	// Close channel and rss tags
	if _, err := fmt.Fprintf(w, "\n</channel>\n</rss>\n"); err != nil {
		return err
	}

	return nil
}