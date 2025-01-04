package main

import (
	"html/template"
	"time"

	_ "embed"
)

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
