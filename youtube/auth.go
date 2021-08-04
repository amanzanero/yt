package youtube

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	cv "github.com/jimlambrt/go-oauth-pkce-code-verifier"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	configName             = ".yt.yaml"
	home                   = ""
	accessTokenKey         = "access_token"
	refreshTokenKey        = "refresh_token"
	expiryKey              = "expiry"
	tokenTypeKey           = "token_type"
	codeChallengeKey       = "code_challenge"
	codeChallengeMethodKey = "code_challenge_method"
	codeVerifierKey        = "code_verifier"
)

func init() {
	setUpViper()
}

func setUpViper() {
	var err error
	home, err = os.UserHomeDir()
	cobra.CheckErr(err)

	// Search config in home directory with name ".cobra" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(configName)
}

type TokenReader interface {
	Read() (string, error)
}

type TokenProvider struct {
	expiresAt    time.Time
	accessToken  string
	refreshToken string
	clientId     string
	authDomain   string
	redirectUrl  string
	scope        string
}

func (p *TokenProvider) Read() (string, error) {
	if time.Now().After(p.expiresAt) {
		// refresh token
	}
	return "", nil
}

func NewTokenProvider(cfg oauth2.Config) (oauth2.TokenSource, error) {
	// first try getting from local memory
	err := viper.ReadInConfig()
	if err != nil || viper.GetString(accessTokenKey) == "" || viper.GetString(refreshTokenKey) == "" {
		writeErr := viper.WriteConfigAs(fmt.Sprintf("%s/%s", home, configName))
		if writeErr != nil {
			color.Red("yt: %v", writeErr)
			os.Exit(1)
		}
		// write config then
		AuthorizeUser(cfg)
	}
	token := &oauth2.Token{
		AccessToken:  viper.GetString(accessTokenKey),
		TokenType:    viper.GetString(tokenTypeKey),
		RefreshToken: viper.GetString(refreshTokenKey),
		Expiry:       viper.GetTime(expiryKey),
	}
	ts := cfg.TokenSource(context.Background(), token)
	return ts, nil
}

func WriteToken(token *oauth2.Token) error {
	viper.Set(accessTokenKey, token.AccessToken)
	viper.Set(refreshTokenKey, token.RefreshToken)
	viper.Set(expiryKey, token.Expiry)
	viper.Set(tokenTypeKey, token.TokenType)
	return viper.WriteConfig()
}

// AuthorizeUser implements the PKCE OAuth2 flow.
func AuthorizeUser(cfg oauth2.Config) {
	var codeVerifier, _ = cv.CreateCodeVerifier()
	authorizationURL := cfg.AuthCodeURL(
		"state",
		oauth2.SetAuthURLParam(codeChallengeKey, codeVerifier.CodeChallengeS256()),
		oauth2.SetAuthURLParam(codeChallengeMethodKey, "S256"),
	)

	server := &http.Server{Addr: cfg.RedirectURL}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			color.Red("yt: Url Param 'code' is missing")
			_, _ = io.WriteString(w, "Error: could not find 'code' URL parameter\n")

			// close the HTTP server and return
			cleanup(server)
			return
		}

		// trade the authorization code and the code verifier for an access token
		token, err := cfg.Exchange(context.Background(), code, oauth2.SetAuthURLParam(codeVerifierKey, codeVerifier.String()))
		if err != nil {
			color.Red("could not get access token: %v\n", err)
			_, _ = io.WriteString(w, "Error: could not retrieve access token\n")
			// close the HTTP server and return
			cleanup(server)
			return
		}

		err = WriteToken(token)
		if err != nil {
			color.Red("could not write config file: %v\n", err)
			_, _ = io.WriteString(w, "error: could not store access token\n")
			// close the HTTP server and return
			cleanup(server)
			return
		}

		// return an indication of success to the caller
		_, _ = io.WriteString(w, `
		<html>
			<body>
				<h1>Login successful!</h1>
				<h2>You can close this window and return to the yt CLI</h2>
			</body>
		</html>`)

		fmt.Println()

		// close the HTTP server
		cleanup(server)
	})

	// parse the redirect URL for the port number
	u, err := url.Parse(cfg.RedirectURL)
	if err != nil {
		color.Red("yt: bad redirect URL: %s\n", err)
		os.Exit(1)
	}

	// set up a listener on the redirect port
	port := fmt.Sprintf(":%s", u.Port())
	l, err := net.Listen("tcp", port)
	if err != nil {
		color.Red("yt: can't listen to port %s: %s\n", port, err)
		os.Exit(1)
	}

	// open a browser window to the authorizationURL
	err = open.Start(authorizationURL)
	if err != nil {
		color.Red("yt: can't open browser to URL %s: %s\n", authorizationURL, err)
		os.Exit(1)
	}

	// start the blocking web server loop
	// this will exit when the handler gets fired and calls server.Close()
	serveErr := server.Serve(l)
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		color.Red("yt: %v\n", serveErr)
	}
}

// cleanup closes the HTTP server
func cleanup(server *http.Server) {
	// we run this as a goroutine so that this function falls through and
	// the socket to the browser gets flushed/closed before the server goes away
	go func() {
		if err := server.Close(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			color.Red("yt: %v\n", err)
		}
	}()
}
