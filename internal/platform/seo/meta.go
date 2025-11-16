package seo

import (
	"html"
	"strings"
)

// Meta models common head metadata for pages, including OpenGraph and Twitter cards.
type Meta struct {
	Title       string
	Description string
	URL         string
	SiteName    string
	ImageURL    string
	Type        string // website | article
	// Twitter
	TwitterCard    string // summary | summary_large_image
	TwitterSite    string // @site
	TwitterCreator string // @author
}

// Default returns a baseline Meta pre-populated with common defaults.
func Default(siteName, siteDescription, baseURL string) Meta {
	return Meta{
		Title:          siteName,
		Description:    siteDescription,
		URL:            baseURL,
		SiteName:       siteName,
		ImageURL:       "",
		Type:           "website",
		TwitterCard:    "summary_large_image",
		TwitterSite:    "",
		TwitterCreator: "",
	}
}

// WithPage returns a copy of m overridden with page-specific values.
func (m Meta) WithPage(title, description, pageURL, imageURL string) Meta {
	cp := m
	if title != "" {
		cp.Title = title
	}
	if description != "" {
		cp.Description = description
	}
	if pageURL != "" {
		cp.URL = pageURL
	}
	if imageURL != "" {
		cp.ImageURL = imageURL
	}
	return cp
}

// Tags renders HTML <meta> and <title> tags for inclusion in a layout head.
// Output is escaped for safety.
func (m Meta) Tags() string {
	esc := func(s string) string { return html.EscapeString(strings.TrimSpace(s)) }
	b := &strings.Builder{}
	// Basic
	b.WriteString("<title>")
	if m.Title != "" {
		b.WriteString(esc(m.Title))
	} else {
		b.WriteString(esc(m.SiteName))
	}
	b.WriteString("</title>")
	if m.Description != "" {
		b.WriteString(`<meta name="description" content="` + esc(m.Description) + `">`)
	}
	// OpenGraph
	if m.Type != "" {
		b.WriteString(`<meta property="og:type" content="` + esc(m.Type) + `">`)
	}
	if m.Title != "" {
		b.WriteString(`<meta property="og:title" content="` + esc(m.Title) + `">`)
	}
	if m.Description != "" {
		b.WriteString(`<meta property="og:description" content="` + esc(m.Description) + `">`)
	}
	if m.URL != "" {
		b.WriteString(`<meta property="og:url" content="` + esc(m.URL) + `">`)
	}
	if m.SiteName != "" {
		b.WriteString(`<meta property="og:site_name" content="` + esc(m.SiteName) + `">`)
	}
	if m.ImageURL != "" {
		b.WriteString(`<meta property="og:image" content="` + esc(m.ImageURL) + `">`)
	}
	// Twitter
	if m.TwitterCard != "" {
		b.WriteString(`<meta name="twitter:card" content="` + esc(m.TwitterCard) + `">`)
	}
	if m.TwitterSite != "" {
		b.WriteString(`<meta name="twitter:site" content="` + esc(m.TwitterSite) + `">`)
	}
	if m.TwitterCreator != "" {
		b.WriteString(`<meta name="twitter:creator" content="` + esc(m.TwitterCreator) + `">`)
	}
	if m.ImageURL != "" {
		b.WriteString(`<meta name="twitter:image" content="` + esc(m.ImageURL) + `">`)
	}
	return b.String()
}
