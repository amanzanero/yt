package cmd

import (
	"github.com/amanzanero/yt/youtube"
	"github.com/spf13/cobra"
	"log"
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

	youtubeService := youtube.New(youtube.WithApiKey(config.YoutubeApiKey))

	comments, err := youtubeService.ListComments(args[0], count)
	if err != nil {
		log.Fatalln(err)
	}
	for _, comment := range comments {
		log.Println(comment.Author)
	}
}
