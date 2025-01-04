package main

import (
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"
)

func markdownToHTML(md string) string {
	// parse markdown
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlContent := string(markdown.Render(doc, renderer))
	return htmlContent
}

func ParseInfoAndContent(content string) (*Post, error) {
	content = strings.TrimSpace(content)
	frontMatterNotFound := errors.New("no front matter found")

	if !strings.HasPrefix(content, "---") {
		return nil, frontMatterNotFound
	}
	content = content[3:]

	parts := strings.SplitN(content, "---", 2)
	if len(parts) < 2 {
		return nil, frontMatterNotFound
	}
	post := Post{}
	err := yaml.Unmarshal([]byte(parts[0]), &post)
	if err != nil {
		return nil, errors.New("error parsing YAML")
	}
	// format date
	parsedDate, err := time.Parse("2006-01-02", post.Date)
	if err != nil {
		return nil, fmt.Errorf("error parsing date: %v", err)
	}
	post.FormattedDate = parsedDate.Format("January 2, 2006")

	// convert content to HTML
	post.Content = template.HTML(markdownToHTML(strings.TrimSpace(parts[1])))

	return &post, nil

}
