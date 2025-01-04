package main

import (
	"encoding/xml"
	"html/template"
	"time"

	_ "embed"
)

// Post Related

//go:embed templates/landing.html
var landingTemplate string

//go:embed templates/post.html
var postTemplate string

type Post struct {
	Title         string        `yaml:"title"`
	Description   string        `yaml:"description"`
	Date          string        `yaml:"date"`
	DateObj       *time.Time    `yaml:"-"`
	IsRedirect    bool          `yaml:"is_redirect"`
	RedirectURL   string        `yaml:"redirect_url"`
	FormattedDate string        `yaml:"-"`
	Content       template.HTML `yaml:"-"`
	Slug          string        `yaml:"-"` // derived from file base path
	FilePath      string        `yaml:"-"`
	Route         string        `yaml:"-"`
}

type PostList []Post

// Sitemap related
type Sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod"`
	ChangeFreq string `xml:"changefreq"`
	Priority   string `xml:"priority"`
}

// RSS related
type RSS struct {
	XMLName xml.Name   `xml:"rss"`
	XMLNS   string     `xml:"xmlns:atom,attr"`
	Version string     `xml:"version,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	TTL         int       `xml:"ttl"` // Added TTL field
	AtomLink    AtomLink  `xml:"atom:link"`
	Items       []RSSItem `xml:"item"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        string `xml:"guid"` // Added GUID field
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}
