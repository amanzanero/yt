package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var (
	NotFoundError   = errors.New("youtube video was not fount")
	BadRequestError = errors.New("something went wrong on our end")
	ApiError        = errors.New("something went wrong with the youtube api")
	UnknownError    = errors.New("something went wrong on our end")
)

func init()  {
	rand.Seed(time.Now().UnixNano())
}

type Service struct {
	apiKey string
}

type appConfig struct {
	YoutubeApiKey string `json:"youtubeApiKey"`
}

func LoadConfig(filename string) (*appConfig, error) {
	secretsJson, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	bytes, readErr := ioutil.ReadAll(secretsJson)
	if readErr != nil {
		return nil, readErr
	}

	config := new(appConfig)
	marshalErr := json.Unmarshal(bytes, config)
	if marshalErr != nil {
		return nil, marshalErr
	}

	return config, nil
}

type Option func(*Service)

func WithApiKey(key string) Option {
	return func(service *Service) {
		service.apiKey = key
	}
}

func New(opts ...Option) Service {
	srvc := new(Service)

	for _, opt := range opts {
		opt(srvc)
	}

	return *srvc
}

type Comment struct {
	Author  string
	Comment string
}

type commentResponse struct {
	Items []struct {
		Snippet struct {
			TopLevelComment struct {
				Snippet struct {
					TextOriginal string `json:"textOriginal"`
					Author       string `json:"authorDisplayName"`
				} `json:"snippet"`
			} `json:"topLevelComment"`
		} `json:"snippet"`
	} `json:"items"`
	NextPageToken string `json:"nextPageToken"`
}

func (s *Service) toplevelCommentsUrl(videoId, nextPageToken string, count int) string {
	apiUrl := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/commentThreads?videoId=%s&key=%s&textFormat=plainText&part=snippet&maxResults=%s",
		videoId,
		s.apiKey,
		strconv.Itoa(count),
	)
	if nextPageToken != "" {
		return apiUrl + fmt.Sprintf("&pageToken=%s", nextPageToken)
	}
	return apiUrl
}

func (s Service) requestComments(videoId, nextPageToken string, count int) (*commentResponse, error) {
	parsedResp := new(commentResponse)
	apiUrl := s.toplevelCommentsUrl(videoId, nextPageToken, count)
	resp, err := http.Get(apiUrl)
	defer func() { _ = resp.Body.Close() }()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == 404 {
		return nil, NotFoundError
	}
	if resp.StatusCode == 500 {
		return nil, ApiError
	}
	if resp.StatusCode == 400 {
		return nil, BadRequestError
	}
	if resp.StatusCode == 500 {
		return nil, UnknownError
	}

	decodeErr := json.NewDecoder(resp.Body).Decode(parsedResp)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return parsedResp, nil
}

func (s *Service) fetchTopLevelComments(videoId string, count int) ([]Comment, error) {
	nextPageToken := ""

	page := 1
	topLevelComments := make([]Comment, 0)
	fmt.Printf("\rgetting comment page %d", page)
	comments, err := s.requestComments(videoId, nextPageToken, count)
	if err != nil {
		return nil, err
	}
	for _, item := range comments.Items {
		topLevelComments = append(topLevelComments, Comment{
			Author:  item.Snippet.TopLevelComment.Snippet.Author,
			Comment: item.Snippet.TopLevelComment.Snippet.TextOriginal,
		})
	}
	nextPageToken = comments.NextPageToken

	for nextPageToken != "" {
		page++
		fmt.Printf("\rgetting comment page %d", page)
		nextComments, err := s.requestComments(videoId, nextPageToken, count)
		if err != nil {
			return nil, err
		}
		for _, item := range nextComments.Items {
			topLevelComments = append(topLevelComments, Comment{
				Author:  item.Snippet.TopLevelComment.Snippet.Author,
				Comment: item.Snippet.TopLevelComment.Snippet.TextOriginal,
			})
		}
		nextPageToken = nextComments.NextPageToken
	}
	fmt.Println()

	return topLevelComments, nil
}

func (s *Service) fetchTopLevelCommenters(videoId string) ([]string, error) {
	nextPageToken := ""

	page := 1
	topLevelCommenters := make(map[string]bool)
	fmt.Printf("\rgetting comment page %d", page)
	comments, err := s.requestComments(videoId, nextPageToken, 100)
	if err != nil {
		return nil, err
	}
	for _, item := range comments.Items {
		topLevelCommenters[item.Snippet.TopLevelComment.Snippet.Author] = true
	}
	nextPageToken = comments.NextPageToken

	for nextPageToken != "" {
		page++
		fmt.Printf("\rgetting comment page %d", page)
		nextComments, err := s.requestComments(videoId, nextPageToken, 100)
		if err != nil {
			return nil, err
		}
		for _, item := range nextComments.Items {
			topLevelCommenters[item.Snippet.TopLevelComment.Snippet.Author] = true
		}
		nextPageToken = nextComments.NextPageToken
	}
	fmt.Println()

	commenters := make([]string, len(topLevelCommenters))
	i := 0
	for author := range topLevelCommenters {
		commenters[i] = author
		i++
	}

	return commenters, nil
}

func (s *Service) ListComments(videoUrl string, maxCount int) ([]Comment, error) {
	videoId, urlErr := parseVideoUrl(videoUrl)
	if urlErr != nil {
		return nil, urlErr
	}

	topLevelComments, fetchErr := s.fetchTopLevelComments(videoId, maxCount)
	if fetchErr != nil {
		return nil, fetchErr
	}

	return topLevelComments, nil
}

func (s *Service) RandomCommenters(videoUrl string, winnerCount int) ([]string, error) {
	videoId, urlErr := parseVideoUrl(videoUrl)
	if urlErr != nil {
		return nil, urlErr
	}

	topLevelCommenters, fetchErr := s.fetchTopLevelCommenters(videoId)
	if fetchErr != nil {
		return nil, fetchErr
	}
	if len(topLevelCommenters) < winnerCount {
		return topLevelCommenters, nil
	}

		winners := make([]int, winnerCount)
	iter := 0
	for iter < winnerCount {
		winner := rand.Intn(len(topLevelCommenters))
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
		winnerNames[i] = topLevelCommenters[val]
	}

	return winnerNames, nil
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
