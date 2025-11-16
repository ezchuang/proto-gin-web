package seo

import (
    "bytes"
    "encoding/xml"
    "time"
)

// RSS models a minimal RSS 2.0 feed structure with common fields.
// See: https://validator.w3.org/feed/docs/rss2.html
type RSS struct {
    Title         string
    Link          string
    Description   string
    Language      string
    PubDate       *time.Time
    LastBuildDate *time.Time
    Items         []RSSItem
}

// RSSItem represents an individual item in the feed.
type RSSItem struct {
    Title           string
    Link            string
    Description     string
    PubDate         *time.Time
    GUID            string
    GUIDIsPermaLink bool
    Categories      []string
    Author          string // optional email or name
}

// Build renders the RSS feed to XML bytes with an XML declaration.
func (r RSS) Build() ([]byte, error) {
    // XML structs for encoding
    type guidXML struct {
        IsPermaLink string `xml:"isPermaLink,attr,omitempty"`
        Value       string `xml:",chardata"`
    }
    type itemXML struct {
        Title       string   `xml:"title,omitempty"`
        Link        string   `xml:"link,omitempty"`
        Description string   `xml:"description,omitempty"`
        PubDate     string   `xml:"pubDate,omitempty"`
        GUID        guidXML  `xml:"guid,omitempty"`
        Category    []string `xml:"category,omitempty"`
        Author      string   `xml:"author,omitempty"`
    }
    type channelXML struct {
        Title         string    `xml:"title"`
        Link          string    `xml:"link"`
        Description   string    `xml:"description"`
        Language      string    `xml:"language,omitempty"`
        PubDate       string    `xml:"pubDate,omitempty"`
        LastBuildDate string    `xml:"lastBuildDate,omitempty"`
        Items         []itemXML `xml:"item"`
    }
    type rssXML struct {
        XMLName xml.Name   `xml:"rss"`
        Version string     `xml:"version,attr"`
        Channel channelXML `xml:"channel"`
    }

    toRFC1123 := func(t *time.Time) string {
        if t == nil {
            return ""
        }
        // RFC1123Z is the common format for RSS pubDate
        return t.UTC().Format(time.RFC1123Z)
    }

    out := rssXML{Version: "2.0"}
    out.Channel = channelXML{
        Title:         r.Title,
        Link:          r.Link,
        Description:   r.Description,
        Language:      r.Language,
        PubDate:       toRFC1123(r.PubDate),
        LastBuildDate: toRFC1123(r.LastBuildDate),
    }
    for _, it := range r.Items {
        guid := guidXML{Value: it.GUID}
        if it.GUID != "" {
            if it.GUIDIsPermaLink {
                guid.IsPermaLink = "true"
            } else {
                guid.IsPermaLink = "false"
            }
        }
        out.Channel.Items = append(out.Channel.Items, itemXML{
            Title:       it.Title,
            Link:        it.Link,
            Description: it.Description,
            PubDate:     toRFC1123(it.PubDate),
            GUID:        guid,
            Category:    it.Categories,
            Author:      it.Author,
        })
    }

    buf := &bytes.Buffer{}
    buf.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
    enc := xml.NewEncoder(buf)
    enc.Indent("", "  ")
    if err := enc.Encode(out); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

// BuildSimple provides a convenience for quick feeds.
func BuildSimple(title, link, description string, items []RSSItem) ([]byte, error) {
    now := time.Now()
    feed := RSS{
        Title:         title,
        Link:          link,
        Description:   description,
        LastBuildDate: &now,
        Items:         items,
    }
    return feed.Build()
}

