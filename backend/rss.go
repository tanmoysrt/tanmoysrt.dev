package main

import (
	"encoding/xml"
	"fmt"
	"time"
)

func GenerateRSS(posts PostList) (string, error) {
	items := make([]RSSItem, 0)
	for _, post := range posts {
		items = append(items, RSSItem{
			Title:       post.Title,
			Link:        fmt.Sprintf("https://tanmoysrt.dev%s", post.Route),
			Description: post.Description,
			PubDate:     post.DateObj.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"),
			GUID:        fmt.Sprintf("https://tanmoysrt.dev%s", post.Route),
		})
	}
	rss := RSS{
		XMLName: xml.Name{Space: "", Local: "rss"},
		Version: "2.0",
		XMLNS:   "http://www.w3.org/2005/Atom",
		Channel: RSSChannel{
			Title:       "Blog | Tanmoy Sarkar",
			Link:        "https://tanmoysrt.dev/",
			Description: "Tanmoy Sarkar's Blog",
			PubDate:     time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"),
			TTL:         60, // 60 minutes
			AtomLink: AtomLink{
				Href: "https://tanmoysrt.dev/rss.xml",
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}

	// Marshal the RSS structure into XML
	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return "", err
	}

	// Add XML declaration at the top
	xmlHeader := []byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	output = append(xmlHeader, output...)

	return string(output), nil
}
