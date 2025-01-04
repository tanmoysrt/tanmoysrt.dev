package main

import (
	"html/template"

	_ "embed"
)

//go:embed templates/landing.html
var landingTemplate string

//go:embed templates/post.html
var postTemplate string

type Post struct {
	Title         string
	Description   string
	Date          string
	FormattedDate string        `yaml:"-"`
	Content       template.HTML `yaml:"-"`
	Slug          string        `yaml:"-"` // derived from file base path
	FilePath      string        `yaml:"-"`
}

type PostList []Post
