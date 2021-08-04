package cmd

import (
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

var (
	rootCmd = &cobra.Command{
		Use:     "yt",
		Short:   "yt is a cli app for creators to easily interact with subscribers and community",
		Long:    `yt lets you choose contest winners and see comments! more is on the way...`,
		Version: "v0.0.0",
	}
	clientId    = "23050909541-45hoolndr2b01infrr909cd3gg9e9idq.apps.googleusercontent.com"
	redirectUrl = "http://localhost:8090"
	config      = oauth2.Config{
		ClientID:     clientId,
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://accounts.google.com/o/oauth2/v2/auth/authorize",
			TokenURL:  "https://yt-cli-321905.ue.r.appspot.com/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		Scopes:      []string{youtube.YoutubeReadonlyScope, youtube.YoutubeForceSslScope},
		RedirectURL: redirectUrl,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
