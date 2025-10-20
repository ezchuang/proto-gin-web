package feed

import (
	"bytes"
	"encoding/xml"
	"time"
)

// Channel captures the top-level metadata for an RSS feed.
type Channel struct {
	Title       string
	Link        string
	Description string
}

// Item represents a single RSS item entry.
type Item struct {
	Title       string
	Link        string
	Description string
	PubDate     *time.Time
}

// BuildRSS renders a standards-compliant RSS 2.0 document.
func BuildRSS(ch Channel, items []Item) ([]byte, error) {
	type itemXML struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		Description string `xml:"description,omitempty"`
		PubDate     string `xml:"pubDate,omitempty"`
	}
	type channelXML struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description,omitempty"`
		Items       []itemXML `xml:"item"`
	}
	type rssXML struct {
		XMLName xml.Name   `xml:"rss"`
		Version string     `xml:"version,attr"`
		Channel channelXML `xml:"channel"`
	}

	out := rssXML{
		Version: "2.0",
		Channel: channelXML{
			Title:       ch.Title,
			Link:        ch.Link,
			Description: ch.Description,
		},
	}
	for _, it := range items {
		entry := itemXML{
			Title:       it.Title,
			Link:        it.Link,
			Description: it.Description,
		}
		if it.PubDate != nil {
			entry.PubDate = it.PubDate.UTC().Format(time.RFC1123Z)
		}
		out.Channel.Items = append(out.Channel.Items, entry)
	}

	buf := &bytes.Buffer{}
	buf.WriteString(xml.Header)

	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(out); err != nil {
		return nil, err
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
