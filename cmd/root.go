package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "yt",
		Short: "yt is a cli app to get random commenters for a youtube video",
		Long:  `yt lets you`,
		Version: "v0.0.0",
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
}
