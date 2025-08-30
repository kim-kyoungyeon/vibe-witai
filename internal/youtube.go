package internal

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Comment struct {
	Author string
	Text   string
}

func FetchComments(videoID string) ([]Comment, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	comments := make([]Comment, 0, 300)
	nextPageToken := ""
	for len(comments) < 300 {
		url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/commentThreads?part=snippet&videoId=%s&key=%s&maxResults=100", videoID, apiKey)
		if nextPageToken != "" {
			url += "&pageToken=" + nextPageToken
		}
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		var result struct {
			Items []struct {
				Snippet struct {
					TopLevelComment struct {
						Snippet struct {
							AuthorDisplayName string `json:"authorDisplayName"`
							TextDisplay       string `json:"textDisplay"`
						} `json:"snippet"`
					} `json:"topLevelComment"`
				} `json:"snippet"`
			} `json:"items"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		for _, item := range result.Items {
			c := item.Snippet.TopLevelComment.Snippet
			comments = append(comments, Comment{Author: c.AuthorDisplayName, Text: c.TextDisplay})
			if len(comments) >= 300 {
				break
			}
		}
		if result.NextPageToken == "" || len(comments) >= 300 {
			break
		}
		nextPageToken = result.NextPageToken
	}
	return comments, nil
}

// ParseVideoID: 입력값에서 유튜브 영상 ID 추출 (URL 또는 ID)
func ParseVideoID(input string) string {
	// 1. ID만 입력된 경우
	if len(input) == 11 && !strings.Contains(input, "/") {
		return input
	}
	// 2. URL에서 추출
	re := regexp.MustCompile(`(?i)(?:v=|youtu.be/|embed/|shorts/)([\w-]{11})`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// FetchRandomPopularVideoID: 인기 영상 중 랜덤으로 하나의 ID 반환
func FetchRandomPopularVideoID() (string, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=id&chart=mostPopular&maxResults=20&regionCode=KR&key=%s", apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Items) == 0 {
		return "", fmt.Errorf("No popular videos found")
	}
	rand.Seed(time.Now().UnixNano())
	idx := rand.Intn(len(result.Items))
	return result.Items[idx].ID, nil
}

// VideoMeta: 썸네일, 채널명, 제목 등 메타데이터 구조체
type VideoMeta struct {
	Title     string
	Channel   string
	Thumbnail string
}

// FetchVideoMeta: 영상 ID로 메타데이터(제목, 채널명, 썸네일) 조회
func FetchVideoMeta(videoID string) (VideoMeta, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet&id=%s&key=%s", videoID, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return VideoMeta{}, err
	}
	defer resp.Body.Close()
	var result struct {
		Items []struct {
			Snippet struct {
				Title        string `json:"title"`
				ChannelTitle string `json:"channelTitle"`
				Thumbnails   struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
					Medium struct {
						URL string `json:"url"`
					} `json:"medium"`
					High struct {
						URL string `json:"url"`
					} `json:"high"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return VideoMeta{}, err
	}
	if len(result.Items) == 0 {
		return VideoMeta{}, fmt.Errorf("영상 정보를 찾을 수 없음")
	}
	snippet := result.Items[0].Snippet
	thumb := snippet.Thumbnails.High.URL
	if thumb == "" {
		thumb = snippet.Thumbnails.Medium.URL
	}
	if thumb == "" {
		thumb = snippet.Thumbnails.Default.URL
	}
	return VideoMeta{
		Title:     snippet.Title,
		Channel:   snippet.ChannelTitle,
		Thumbnail: thumb,
	}, nil
}
