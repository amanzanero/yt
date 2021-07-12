package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

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

func (s *Service) fetchTopLevelComments(videoId string, count int) ([]Comment, error) {
	nextPageToken := ""

	getCommentPage := func() (*commentResponse, error) {
		parsedResp := new(commentResponse)
		apiUrl := s.toplevelCommentsUrl(videoId, nextPageToken, count)
		resp, err := http.Get(apiUrl)
		defer func() { _ = resp.Body.Close() }()
		if err != nil {
			return nil, err
		}

		decodeErr := json.NewDecoder(resp.Body).Decode(parsedResp)
		if decodeErr != nil {
			return nil, decodeErr
		}
		nextPageToken = parsedResp.NextPageToken

		return parsedResp, nil
	}

	page := 1
	topLevelComments := make([]Comment, 0)
	log.Printf("getting page #%d, nextPageToken:%s", page, nextPageToken)
	comments, err := getCommentPage()
	if err != nil {
		return nil, err
	}
	for _, item := range comments.Items {
		topLevelComments = append(topLevelComments, Comment{
			Author:  item.Snippet.TopLevelComment.Snippet.Author,
			Comment: item.Snippet.TopLevelComment.Snippet.TextOriginal,
		})
	}

	for nextPageToken != "" {
		page++
		log.Printf("getting page #%d, nextPageToken:%s", page, nextPageToken)
		nextComments, err := getCommentPage()
		if err != nil {
			return nil, err
		}
		for _, item := range nextComments.Items {
			topLevelComments = append(topLevelComments, Comment{
				Author:  item.Snippet.TopLevelComment.Snippet.Author,
				Comment: item.Snippet.TopLevelComment.Snippet.TextOriginal,
			})
		}
	}

	return topLevelComments, nil
}

func (s *Service) fetchTopLevelCommenters(videoId string) ([]string, error) {
	nextPageToken := ""

	getCommentPage := func() (*commentResponse, error) {
		parsedResp := new(commentResponse)
		apiUrl := s.toplevelCommentsUrl(videoId, nextPageToken, 100)
		resp, err := http.Get(apiUrl)
		defer func() { _ = resp.Body.Close() }()
		if err != nil {
			return nil, err
		}

		decodeErr := json.NewDecoder(resp.Body).Decode(parsedResp)
		if decodeErr != nil {
			return nil, decodeErr
		}
		nextPageToken = parsedResp.NextPageToken

		return parsedResp, nil
	}

	page := 1
	topLevelCommenters := make(map[string]bool)
	log.Printf("getting page #%d, nextPageToken:%s", page, nextPageToken)
	comments, err := getCommentPage()
	if err != nil {
		return nil, err
	}
	for _, item := range comments.Items {
		topLevelCommenters[item.Snippet.TopLevelComment.Snippet.Author] = true
	}

	for nextPageToken != "" {
		page++
		log.Printf("getting page #%d, nextPageToken:%s", page, nextPageToken)
		nextComments, err := getCommentPage()
		if err != nil {
			return nil, err
		}
		for _, item := range nextComments.Items {
			topLevelCommenters[item.Snippet.TopLevelComment.Snippet.Author] = true
		}
	}

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

func (s *Service) RandomCommenter(videoUrl string) (string, error) {
	videoId, urlErr := parseVideoUrl(videoUrl)
	if urlErr != nil {
		return "", urlErr
	}

	topLevelCommenters, fetchErr := s.fetchTopLevelCommenters(videoId)
	if fetchErr != nil {
		return "", fetchErr
	}

	winner := rand.Intn(len(topLevelCommenters))

	return topLevelCommenters[winner], nil
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
