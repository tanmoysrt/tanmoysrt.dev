package main

import (
	"encoding/xml"
	"fmt"
	"time"
)

func GenerateSitemap(posts PostList) (string, error) {
	sitemap := Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
	}

	sitemap.URLs = make([]SitemapURL, 0)
	// add the homepage
	sitemap.URLs = append(sitemap.URLs, SitemapURL{
		Loc:        "https://tanmoysrt.dev/",
		LastMod:    time.Now().Format("2006-01-02"),
		ChangeFreq: "daily",
		Priority:   "1.0",
	})

	// add the posts
	for _, u := range posts {
		sitemap.URLs = append(sitemap.URLs, SitemapURL{
			Loc:        fmt.Sprintf("https://tanmoysrt.dev%s", u.Route),
			LastMod:    u.Date,
			ChangeFreq: "weekly",
			Priority:   "0.5",
		})
	}

	// Marshal the sitemap to XML
	output, err := xml.MarshalIndent(sitemap, "", "  ")
	if err != nil {
		return "", err
	}

	// Add XML declaration at the top
	xmlHeader := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	output = append(xmlHeader, output...)

	return string(output), nil
}
