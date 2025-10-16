package seo

import (
    "bytes"
    "encoding/xml"
    "fmt"
    "net/url"
    "path"
    "strings"
    "time"
)

// URLEntry represents a single <url> element in a sitemap.
// All fields except Loc are optional and will be omitted if empty.
type URLEntry struct {
    Loc        string
    LastMod    *time.Time    // RFC3339 or date, formatted as 2006-01-02
    ChangeFreq string        // e.g., always|hourly|daily|weekly|monthly|yearly|never
    Priority   *float64      // 0.0 - 1.0
}

// Build constructs a sitemap XML document from the provided URL entries.
// The output includes the XML declaration and required urlset namespace.
func Build(entries []URLEntry) ([]byte, error) {
    type urlXML struct {
        Loc        string  `xml:"loc"`
        LastMod    string  `xml:"lastmod,omitempty"`
        ChangeFreq string  `xml:"changefreq,omitempty"`
        Priority   string  `xml:"priority,omitempty"`
    }
    type urlset struct {
        XMLName xml.Name `xml:"urlset"`
        Xmlns   string   `xml:"xmlns,attr"`
        URLs    []urlXML `xml:"url"`
    }

    out := urlset{Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9"}
    for _, e := range entries {
        u := urlXML{Loc: e.Loc}
        if e.LastMod != nil {
            // Use date-only format which is valid and commonly used in sitemaps
            u.LastMod = e.LastMod.UTC().Format("2006-01-02")
        }
        if e.ChangeFreq != "" {
            u.ChangeFreq = e.ChangeFreq
        }
        if e.Priority != nil {
            u.Priority = fmt.Sprintf("%.1f", *e.Priority)
        }
        out.URLs = append(out.URLs, u)
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

// BuildFromPaths is a convenience helper to build a sitemap for a base URL
// and a list of relative paths (e.g., ["/", "/posts/hello"]).
// - lastmod is set to now
// - changefreq defaults to "daily"
// - priority is 1.0 for "/" and 0.8 for other paths
func BuildFromPaths(base string, relPaths []string) ([]byte, error) {
    var entries []URLEntry
    now := time.Now()
    baseURL, err := url.Parse(base)
    if err != nil {
        return nil, err
    }
    for _, p := range relPaths {
        // Normalize path joining while preserving absolute URLs if provided
        var loc string
        if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
            loc = p
        } else {
            // Ensure leading slash
            if !strings.HasPrefix(p, "/") {
                p = "/" + p
            }
            u := *baseURL
            u.Path = path.Clean(path.Join("/", p))
            // Preserve trailing slash only for root
            loc = u.String()
        }
        var prio float64 = 0.8
        if p == "/" {
            prio = 1.0
        }
        entries = append(entries, URLEntry{
            Loc:        loc,
            LastMod:    &now,
            ChangeFreq: "daily",
            Priority:   &prio,
        })
    }
    return Build(entries)
}
