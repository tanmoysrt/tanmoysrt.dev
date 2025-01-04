package main

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "blogger",
	Short: "A simple markdown to static HTML blog generator",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	rootCmd.Execute()
}
