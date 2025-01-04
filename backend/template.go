package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

func CompileTemplate(data interface{}, templateString string) (string, error) {
	t, err := template.New("post").Parse(templateString)
	if err != nil {
		panic(err)
	}

	var outputBuffer bytes.Buffer
	err = t.Execute(&outputBuffer, data)
	if err != nil {
		return "", err
	}

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)

	var minifiedOutputBuffer bytes.Buffer
	err = m.Minify("text/html", &minifiedOutputBuffer, &outputBuffer)
	if err != nil {
		return "", err
	}

	return minifiedOutputBuffer.String(), nil
}

func ReadPosts() (PostList, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %v", err)
	}
	if filepath.Base(currentPath) != "posts" {
		return nil, fmt.Errorf("please run this command from the posts directory")
	}
	// find all markdown files in the posts directory
	files, err := filepath.Glob(filepath.Join(currentPath, "*.md"))
	if err != nil {
		return nil, fmt.Errorf("error reading posts directory: %v", err)
	}
	// read each file and parse it
	var posts PostList = make([]Post, 0)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("error reading file %s: %v", file, err)
		}
		post, err := ParseInfoAndContent(string(content))
		if err != nil {
			return nil, fmt.Errorf("error parsing file %s: %v", file, err)
		}
		post.FilePath = file
		post.Slug = filepath.Base(file)
		if len(post.Slug) > 3 && post.Slug[len(post.Slug)-3:] == ".md" {
			post.Slug = post.Slug[:len(post.Slug)-3]
		}
		post.Route = fmt.Sprintf("/posts/%s", post.Slug) // Set route to /posts/:slug
		posts = append(posts, *post)
	}

	// Sort by date desc
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].DateObj.After(*posts[j].DateObj)
	})

	return posts, nil
}
