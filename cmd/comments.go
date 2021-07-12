package cmd

import (
	"fmt"
	"github.com/amanzanero/yt/youtube"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var count int

func init() {
	localCmd := &cobra.Command{
		Use:     "comments",
		Short:   "Print out youtube comments",
		Long:    `This command prints out all comments for a given youtube video`,
		Example: "yt comments https://www.youtube.com/watch?v=oYBGPVwNK2c&ab_channel=Katherout",
		Args:    cobra.MinimumNArgs(1),
		Run:     commentsCmd,
	}
	localCmd.Flags().IntVar(&count, "count", 10, "max number of comments")
	rootCmd.AddCommand(localCmd)
}

func commentsCmd(_ *cobra.Command, args []string) {
	config, configErr := youtube.LoadConfig("secrets.json")
	if configErr != nil {
		log.Fatalln(configErr)
	}

	start := time.Now()
	youtubeService := youtube.New(youtube.WithApiKey(config.YoutubeApiKey))
	comments, err := youtubeService.ListComments(args[0], count)
	stop := time.Since(start).Milliseconds()
	if err != nil {
		log.Fatalln(err)
	}
	for i, comment := range comments {
		fmt.Printf("%d: %s\n", i, comment)
	}
	fmt.Printf("took %dms\n", stop)
}
