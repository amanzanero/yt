package cmd

import (
	"fmt"
	"github.com/amanzanero/yt/youtube"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var winnerCount int

func init() {
	localCmd := &cobra.Command{
		Use:     "winner",
		Short:   "Print out youtube winner",
		Long:    `This command prints out random winner(s)`,
		Example: "yt winner https://www.youtube.com/watch?v=oYBGPVwNK2c&ab_channel=Katherout",
		Args:    cobra.MinimumNArgs(1),
		Run:     winnerCmd,
	}
	localCmd.Flags().IntVarP(&winnerCount, "count", "c", 1, "Total number of winners. Each user is considered equally regardless of number of comments.")
	rootCmd.AddCommand(localCmd)
}

func winnerCmd(_ *cobra.Command, args []string) {
	tokenProvider, err := youtube.NewTokenProvider(config)
	cobra.CheckErr(err)

	youtubeService, err := youtube.New(youtube.WithTokenSource(tokenProvider))
	if err != nil {
		color.Red("yt: %v", err)
		os.Exit(1)
	}

	start := time.Now()
	winners, err := youtubeService.RandomCommenters(args[0], winnerCount)
	total := time.Since(start).Milliseconds()
	if err != nil {
		color.Red("yt: %v", err)
	}
	for i, winner := range winners {
		color.Blue("Winner #%d: \"%s\"\n", i+1, winner)
	}
	fmt.Printf("took %dms\n", total)

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
