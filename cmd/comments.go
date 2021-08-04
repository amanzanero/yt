package cmd

import (
	"fmt"
	"github.com/amanzanero/yt/youtube"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var count int

func init() {
	localCmd := &cobra.Command{
		Use:     "comments",
		Short:   "Print out youtube comments",
		Long:    `This command prints out all comments for a given youtube video`,
		Example: "yt comments https://www.youtube.com/watch?v=BIk1zUy8ehU&ab_channel=LexFridman",
		Args:    cobra.MinimumNArgs(1),
		Run:     commentsCmd,
	}
	localCmd.Flags().IntVarP(&count, "count", "c", 10, "max number of comments")
	rootCmd.AddCommand(localCmd)
}

func commentsCmd(_ *cobra.Command, args []string) {
	tokenProvider, err := youtube.NewTokenProvider(config)
	cobra.CheckErr(err)

	youtubeService, err := youtube.New(youtube.WithTokenSource(tokenProvider))
	if err != nil {
		color.Red("yt: %v", err)
		os.Exit(1)
	}

	start := time.Now()
	comments, err := youtubeService.ListComments(args[0], count)
	stop := time.Since(start).Milliseconds()
	if err != nil {
		color.Red("yt: %v", err)
	}
	for i, comment := range comments {
		fmt.Printf("%d: [%s] %s\n", i+1, comment.AuthorDisplayName, comment.TextOriginal)
	}
	fmt.Printf("took %dms\n", stop)

	// store token in case it refreshed
	token, err := tokenProvider.Token()
	if err != nil {
		color.Red("yt: could not save token")
		os.Exit(1)
	}
	err = youtube.WriteToken(token)
	if err != nil {
		color.Red("yt: could not save token")
		os.Exit(1)
	}
}
