package cmd

import (
	"fmt"
	"github.com/amanzanero/yt/youtube"
	"github.com/spf13/cobra"
	"log"
	"time"
)

func init() {
	localCmd := &cobra.Command{
		Use:     "winner",
		Short:   "Print out youtube winner",
		Long:    `This command prints out a random winner`,
		Example: "yt winner https://www.youtube.com/watch?v=oYBGPVwNK2c&ab_channel=Katherout",
		Args:    cobra.MinimumNArgs(1),
		Run:     winnerCmd,
	}
	rootCmd.AddCommand(localCmd)
}

func winnerCmd(_ *cobra.Command, args []string) {
	config, configErr := youtube.LoadConfig("secrets.json")
	if configErr != nil {
		log.Fatalln(configErr)
	}

	youtubeService := youtube.New(youtube.WithApiKey(config.YoutubeApiKey))

	start := time.Now()
	winner, err := youtubeService.RandomCommenter(args[0])
	total := time.Since(start).Milliseconds()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Winner: " + winner)
	fmt.Printf("took %dms\n", total)
}
