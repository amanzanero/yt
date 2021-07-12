package youtube

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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
}

func (y *Service) fetchTopLevelComments(videoId string, count int) (*commentResponse, error) {
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/commentThreads?videoId=%s&key=%s&textFormat=plainText&part=snippet&maxResults=%s",
		videoId,
		y.apiKey,
		strconv.Itoa(count),
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	parsedResp := new(commentResponse)
	decodeErr := json.NewDecoder(resp.Body).Decode(parsedResp)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return parsedResp, nil
}

func (y *Service) ListComments(videoUrl string, maxCount int) ([]Comment, error) {
	videoId, urlErr := parseVideoUrl(videoUrl)
	if urlErr != nil {
		return nil, urlErr
	}

	topLevelComments, fetchErr := y.fetchTopLevelComments(videoId, maxCount)
	if fetchErr != nil {
		return nil, fetchErr
	}
	comments := make([]Comment, len(topLevelComments.Items))
	for i, comment := range topLevelComments.Items {
		comments[i].Comment = comment.Snippet.TopLevelComment.Snippet.TextOriginal
		comments[i].Author = comment.Snippet.TopLevelComment.Snippet.Author
	}
	return comments, nil
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
