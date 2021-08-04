package youtube

import (
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"math/rand"
	"net/url"
	"time"
)

var maxResults = 10

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Service struct {
	apiKey      string
	tokenSource oauth2.TokenSource
	ytService   *youtube.Service
}

type Option func(*Service)

func WithTokenSource(tokenSource oauth2.TokenSource) Option {
	return func(service *Service) {
		service.tokenSource = tokenSource
	}
}

func New(opts ...Option) (*Service, error) {
	srvc := new(Service)

	for _, opt := range opts {
		opt(srvc)
	}

	if srvc.tokenSource == nil {
		return nil, errors.New("yt: missing token source")
	}

	ytService, err := youtube.NewService(context.Background(), option.WithTokenSource(srvc.tokenSource))
	if err != nil {
		return nil, errors.New("yt: error initializing youtube API %v")
	}
	srvc.ytService = ytService

	return srvc, nil
}

func (s *Service) ListComments(videoUrl string, count int) ([]*youtube.CommentSnippet, error) {
	videoId, urlErr := parseVideoUrl(videoUrl)
	if urlErr != nil {
		return nil, urlErr
	}

	shutdown := make(chan bool)
	commChan := make(chan struct {
		comments []*youtube.CommentSnippet
		err      error
	})
	tickerLogger("loading youtube comments", shutdown)
	go func() {
		comments, err := s.getNumberOfComments(videoId, count)
		commChan <- struct {
			comments []*youtube.CommentSnippet
			err      error
		}{comments, err}
		close(shutdown)
	}()
	result := <-commChan
	return result.comments, result.err
}

func (s *Service) RandomCommenters(videoUrl string, winnerCount int) ([]string, error) {
	videoId, urlErr := parseVideoUrl(videoUrl)
	if urlErr != nil {
		return nil, urlErr
	}

	shutdown := make(chan bool)
	commChan := make(chan struct {
		comments []*youtube.CommentSnippet
		err      error
	})
	tickerLogger("loading youtube commenters", shutdown)
	go func() {
		comments, err := s.getComments(videoId)
		commChan <- struct {
			comments []*youtube.CommentSnippet
			err      error
		}{comments, err}
		close(shutdown)
	}()
	result := <-commChan
	if result.err != nil {
		return nil, result.err
	}
	commenters := result.comments

	winners := make([]int, winnerCount)
	iter := 0
	for iter < winnerCount {
		winner := rand.Intn(len(commenters))
		unique := true
		for _, val := range winners {
			if winner == val {
				unique = false
				break
			}
		}

		if unique {
			winners[iter] = winner
			iter++
		}
	}
	winnerNames := make([]string, winnerCount)
	for i, val := range winners {
		winnerNames[i] = commenters[val].AuthorDisplayName
	}

	return winnerNames, nil
}

func (s *Service) getComments(videoId string) ([]*youtube.CommentSnippet, error) {
	comments := make([]*youtube.CommentSnippet, 0)

	commentRequest := s.ytService.CommentThreads.List([]string{"snippet"}).VideoId(videoId).MaxResults(int64(maxResults))
	resp, err := commentRequest.Do()
	if err != nil {
		return nil, err
	}
	for _, i := range resp.Items {
		comments = append(comments, i.Snippet.TopLevelComment.Snippet)
	}
	for resp.NextPageToken != "" {
		resp, err = commentRequest.PageToken(resp.NextPageToken).Do()
		if err != nil {
			return nil, err
		}
		for _, i := range resp.Items {
			comments = append(comments, i.Snippet.TopLevelComment.Snippet)
		}
	}

	return comments, nil
}

func (s *Service) getNumberOfComments(videoId string, count int) ([]*youtube.CommentSnippet, error) {
	comments := make([]*youtube.CommentSnippet, 0)

	limit := count
	if count > maxResults {
		limit = maxResults
	}

	commentRequest := s.ytService.CommentThreads.List([]string{"snippet"}).VideoId(videoId)
	resp, err := commentRequest.MaxResults(int64(limit)).Do()
	if err != nil {
		return nil, err
	}
	for _, i := range resp.Items {
		comments = append(comments, i.Snippet.TopLevelComment.Snippet)
	}
	for resp.NextPageToken != "" {
		if len(comments) >= count {
			break
		}

		// reset the limit if possible
		if count-len(comments) < maxResults {
			limit = count - len(comments)
		}

		resp, err = commentRequest.PageToken(resp.NextPageToken).MaxResults(int64(limit)).Do()
		if err != nil {
			return nil, err
		}
		for _, i := range resp.Items {
			comments = append(comments, i.Snippet.TopLevelComment.Snippet)
		}
	}

	if len(comments) < count {
		return comments, nil
	}
	return comments[0:count], nil
}

func tickerLogger(message string, shutdown <-chan bool) {
	ticker := time.Tick(time.Millisecond)
	start := time.Now()
	green := color.New(color.FgGreen).SprintfFunc()
	go func() {
		for {
			select {
			case <-ticker:
				fmt.Printf("\r%s", green("%s... (%dms)", message, time.Since(start).Milliseconds()))
			case <-shutdown:
				fmt.Println()
				return
			}
		}
	}()
}

var InvalidUrlErr = errors.New("youtube url was not in correct format")

func parseVideoUrl(videoUrl string) (string, error) {
	u, err := url.ParseRequestURI(videoUrl)
	if err != nil {
		return "", err
	}
	if u.Host != "www.youtube.com" {
		return "", InvalidUrlErr
	}
	q, parseErr := url.ParseQuery(u.RawQuery)
	if parseErr != nil {
		return "", parseErr
	}

	return q.Get("v"), nil
}
