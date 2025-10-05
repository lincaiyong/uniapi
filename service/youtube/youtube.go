package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lincaiyong/uniapi/utils"
	"io"
	"net/http"
)

type SubtitleText struct {
	StartTime float64
	EndTime   float64
	Text      string
}

func requestCaptionTrackBaseUrl(client *http.Client, videoID string) (string, error) {
	reqUrl := "https://www.youtube.com/youtubei/v1/player?prettyPrint=false"
	data, _ := json.Marshal(PlayerRequest{
		Context: PlayerRequestContext{
			Client: PlayerRequestContextClient{
				ClientName:    "WEB",
				ClientVersion: "2.20250925.01.00",
			},
		},
		VideoID: videoID,
	})
	req, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Safari/605.1.15")

	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return "", fmt.Errorf("fail to do client request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	var playerResp PlayerResponse
	if err = json.Unmarshal(body, &playerResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}
	for _, track := range playerResp.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks {
		if track.LanguageCode == "en" && track.Kind == "asr" {
			return track.BaseURL, nil
		}
	}
	return "", errors.New("no captions found for this video")
}

func requestTimedText(client *http.Client, trackBaseUrl string) (*Caption, error) {
	reqUrl := trackBaseUrl + "&fmt=json3"
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.5 Safari/605.1.15")

	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return nil, fmt.Errorf("fail to do client request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read subtitle response: %w", err)
	}

	var caption Caption
	if err = json.Unmarshal(body, &caption); err != nil {
		return nil, fmt.Errorf("failed to unmarshal subtitle response: %w", err)
	}
	return &caption, nil
}

func DownloadAutoCaption(videoID string) (*Caption, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	trackBaseUrl, err := requestCaptionTrackBaseUrl(client, videoID)
	if err != nil {
		return nil, err
	}
	caption, err := requestTimedText(client, trackBaseUrl)
	if err != nil {
		return nil, err
	}
	return caption, nil
}
