package main

import (
	"fmt"
	"os"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateCmd)
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate static HTML from markdown files",
	Long:  `Generate static HTML from markdown files in the dist directory`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = os.RemoveAll("dist")
		// Create dist directory
		fmt.Println("Creating dist directory...")
		err := os.Mkdir("dist", 0777)
		if err != nil {
			panic(err)
		}
		// Create the assets directory
		fmt.Println("Creating dist/assets directory...")
		err = os.Mkdir("dist/assets", 0777)
		if err != nil {
			panic(err)
		}
		// Read posts
		fmt.Println("Reading posts...")
		posts, err := ReadPosts()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Found %d posts\n", len(posts))
		// // Compile landing page
		fmt.Println("Compiling landing page...")
		landingPage, err := CompileTemplate(posts, landingTemplate)
		if err != nil {
			panic(err)
		}
		// Write to file
		err = os.WriteFile("dist/index.html", []byte(landingPage), 0777)
		if err != nil {
			panic(err)
		}

		fmt.Println("Compiling posts...")
		for _, post := range posts {
			fmt.Printf("[COMPILE] %s\n", post.Slug)
			postPage, err := CompileTemplate(post, postTemplate)
			if err != nil {
				panic(err)
			}

			fmt.Printf("[WRITE] %s.html\n", post.Slug)
			err = os.WriteFile(fmt.Sprintf("dist/%s.html", post.Slug), []byte(postPage), 0777)
			if err != nil {
				panic(err)
			}
			fmt.Printf("[DONE] %s\n", post.Slug)
		}

		// If `posts/assets` exists, copy it to `dist/assets`
		if _, err := os.Stat("posts/assets"); err == nil {
			fmt.Println("Copying assets...")
			err = cp.Copy("posts/assets", "dist/assets")
			if err != nil {
				panic(err)
			}
		}
	},
}
