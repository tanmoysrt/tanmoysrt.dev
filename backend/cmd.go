package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(syncCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate static HTML from markdown files",
	Long:  `Generate static HTML from markdown files in the dist directory`,
	Run: func(cmd *cobra.Command, args []string) {
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
		if _, err := os.Stat("posts/assets"); err == nil {
			fmt.Println("Copying assets...")
			err = cp.Copy("assets", "dist/assets")
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
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %v\n", err)
			return
		}

		// Check if we're in posts directory and move up if needed
		if filepath.Base(cwd) == "posts" {
			err = os.Chdir("..")
			if err != nil {
				fmt.Printf("Error moving up directory: %v\n", err)
				return
			}
			defer os.Chdir(cwd) // Return to original directory when done
		}

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
