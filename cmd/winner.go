package cmd

import (
	"fmt"
	"github.com/amanzanero/yt/youtube"
	"github.com/spf13/cobra"
	"log"
	"time"
)

var winnerCount int

func init() {
	localCmd := &cobra.Command{
		Use:     "winner",
		Short:   "Print out youtube winner",
		Long:    `This command prints out a random winner`,
		Example: "yt winner https://www.youtube.com/watch?v=oYBGPVwNK2c&ab_channel=Katherout",
		Args:    cobra.MinimumNArgs(1),
		Run:     winnerCmd,
	}
	localCmd.Flags().IntVar(&winnerCount, "count", 1, "Total number of winners (each commenter has an equal chance of winning regardless of comment count)")
	rootCmd.AddCommand(localCmd)
}

func winnerCmd(_ *cobra.Command, args []string) {
	config, configErr := youtube.LoadConfig("secrets.json")
	if configErr != nil {
		log.Fatalln(configErr)
	}

	youtubeService := youtube.New(youtube.WithApiKey(config.YoutubeApiKey))

	start := time.Now()
	winners, err := youtubeService.RandomCommenters(args[0], winnerCount)
	total := time.Since(start).Milliseconds()
	if err != nil {
		log.Fatalln(err)
	}
	for i, winner := range winners {
		fmt.Printf("Winner #%d: \"%s\"\n", i+1, winner)
	}
	fmt.Printf("took %dms\n", total)
}
