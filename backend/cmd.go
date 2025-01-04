package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/gosimple/slug"
	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(newPostCmd)
	rootCmd.AddCommand(syncCmd)
}

func validatePath() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}
	// Check if we're in posts directory and move up if needed
	if filepath.Base(cwd) != "posts" {
		return fmt.Errorf("please run this command from the posts directory")
	}
	return nil
}

var newPostTemplate = `
---
title: "{{ .Title }}"
description: "Write your description here"
date: "{{ .Date }}"
is_redirect: false
redirect_url: 
---

Write your post here
`

var newPostCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new post",
	Long:  `Create a new post in the posts directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validatePath(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// Ask for post title
		var title string
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter post title: ")
		fmt.Print("> ")
		if scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("Input was: %q\n", line)
			title = line
		} else {
			fmt.Println("Error reading input")
			os.Exit(1)
		}

		// If title is empty, exit
		if title == "" {
			fmt.Println("Post title cannot be empty")
			os.Exit(1)
		}

		generatedSlug := slug.Make(title)
		postFileName := fmt.Sprintf("%s.md", generatedSlug)

		// Create post file
		fmt.Printf("Creating post %s...\n", postFileName)
		t, err := template.New("post").Parse(newPostTemplate)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		var buf bytes.Buffer
		err = t.Execute(&buf, map[string]string{
			"Title": title,
			"Date":  time.Now().Format("2006-01-02"),
		})
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		err = os.WriteFile(postFileName, buf.Bytes(), 0777)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Printf("Post %s created successfully\n", postFileName)
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build static HTML site from markdown files",
	Long:  `Build static HTML site from markdown files in the dist directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validatePath(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// Read posts
		fmt.Println("Reading posts...")
		posts, err := ReadPosts()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// Remove dist directory if it exists
		_ = os.RemoveAll("dist")
		// Create dist directory
		fmt.Println("Creating dist directory...")
		err = os.Mkdir("dist", 0777)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// Create the assets directory
		fmt.Println("Creating dist/assets directory...")
		err = os.Mkdir("dist/assets", 0777)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Found %d posts\n", len(posts))
		// // Compile landing page
		fmt.Println("Compiling landing page...")
		landingPage, err := CompileTemplate(posts, landingTemplate)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// Write to file
		err = os.WriteFile("dist/index.html", []byte(landingPage), 0777)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Println("Compiling posts...")
		for _, post := range posts {
			fmt.Printf("[COMPILE] %s\n", post.Slug)
			postPage, err := CompileTemplate(post, postTemplate)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("[WRITE] %s.html\n", post.Slug)
			err = os.WriteFile(fmt.Sprintf("dist/%s.html", post.Slug), []byte(postPage), 0777)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Printf("[DONE] %s\n", post.Slug)
		}

		// If `posts/assets` exists, copy it to `dist/assets`
		if _, err := os.Stat("assets"); err == nil {
			fmt.Println("Copying assets...")
			err := cp.Copy("assets", "dist/assets")
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync posts with a git repository",
	Long:  `Sync posts with a git repository`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validatePath(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}

		err = os.Chdir("..")
		if err != nil {
			fmt.Printf("Error moving up directory: %v\n", err)
			return
		}
		defer os.Chdir(cwd) // Return to original directory when done

		// Initialize git command executor
		git := func(args ...string) error {
			cmd := exec.Command("git", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}

		// Check for changes
		statusCmd := exec.Command("git", "status", "--porcelain")
		output, err := statusCmd.Output()
		if err != nil {
			fmt.Printf("Error checking git status: %v\n", err)
			return
		}

		// Only proceed with commit if there are changes
		if len(output) > 0 {
			fmt.Println("Changes detected, committing...")

			// Add all changes
			if err := git("add", "."); err != nil {
				fmt.Printf("Error adding files: %v\n", err)
				return
			}

			// Commit changes
			if err := git("commit", "-m", "Update posts"); err != nil {
				fmt.Printf("Error committing changes: %v\n", err)
				return
			}
		} else {
			fmt.Println("No changes to commit")
		}

		// Pull with rebase
		fmt.Println("Pulling latest changes...")
		if err := git("pull", "origin", "master", "--rebase"); err != nil {
			fmt.Printf("Error pulling changes: %v\n", err)
			return
		}

		// Push changes
		fmt.Println("Pushing changes...")
		if err := git("push", "origin", "master"); err != nil {
			fmt.Printf("Error pushing changes: %v\n", err)
			return
		}

		fmt.Println("Sync completed successfully!")
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the site to Netlify",
	Long:  `Deploy the site to Netlify`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := validatePath(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// check for required environment variables
		NETLIFY_TOKEN := os.Getenv("NETLIFY_TOKEN")
		NETLIFY_BLOG_SITE_ID := os.Getenv("NETLIFY_BLOG_SITE_ID")
		if NETLIFY_TOKEN == "" {
			fmt.Println("NETLIFY_TOKEN environment variable not set")
			os.Exit(1)
		}
		if NETLIFY_BLOG_SITE_ID == "" {
			fmt.Println("NETLIFY_BLOG_SITE_ID environment variable not set")
			os.Exit(1)
		}
		// check for zip utility
		_, err := exec.LookPath("zip")
		if err != nil {
			fmt.Println("zip utility not found. Please install zip utility")
			os.Exit(1)
		}

		// build the site
		buildCmd.Run(nil, nil)

		// create zip file of dist directory
		if _, err := os.Stat("dist"); err != nil {
			fmt.Println("dist directory not found")
			os.Exit(1)
		}
		fmt.Println("Creating zip file...")
		_ = os.Remove("dist.zip")
		zipCmd := exec.Command("zip", "-r", "dist.zip", ".", "-i", "dist/*")
		err = zipCmd.Run()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer os.Remove("dist.zip")
		// deploy the site
		fmt.Println("Deploying site to Netlify...")

		httpClient := &http.Client{}
		req, err := http.NewRequest("POST", fmt.Sprintf("https://api.netlify.com/api/v1/sites/%s/deploys", NETLIFY_BLOG_SITE_ID), nil)
		if err != nil {
			fmt.Println("Error creating request")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		req.Header.Set("Content-Type", "application/zip")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", NETLIFY_TOKEN))
		// read zip file
		zipFile, err := os.Open("dist.zip")
		if err != nil {
			fmt.Println("Error reading zip file")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer zipFile.Close()
		req.Body = zipFile
		response, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("Error deploying site")
			fmt.Println(err.Error())
			os.Exit(1)
		}
		if response.StatusCode != 200 {
			fmt.Println("Error deploying site")
			fmt.Println(response.Status)
			os.Exit(1)
		}
		fmt.Println("Site deployed successfully")
	},
}
