package flomo

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/log"
	"github.com/lincaiyong/uniapi/utils"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Memo struct {
	Slug      string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	Tags      []string
}

var gToken string

func Init(token string) {
	gToken = token
}

func UpdatedMemo(slug string, latestUpdatedAt time.Time) ([]*Memo, error) {
	ret := make([]*Memo, 0)
	limit := 200
	memos, err := updatedMemoByPage(latestUpdatedAt, slug, limit)
	if err != nil {
		return nil, err
	}
	ret = append(ret, memos...)
	for len(memos) == limit {
		lastMemo := memos[len(memos)-1]
		memos, err = updatedMemoByPage(lastMemo.UpdatedAt, lastMemo.Slug, limit)
		if err != nil {
			return nil, err
		}
		ret = append(ret, memos...)
	}
	return ret, nil
}

func updatedMemoByPage(latestUpdatedAt time.Time, slug string, limit int) ([]*Memo, error) {
	baseURL := "https://flomoapp.com/api/v1/memo/updated/"

	n := map[string]string{
		"limit":             strconv.Itoa(limit),
		"latest_updated_at": strconv.Itoa(int(latestUpdatedAt.UTC().Unix())),
		"latest_slug":       slug,
		"tz":                "8:0",
		"timestamp":         strconv.Itoa(int(time.Now().Unix())),
		"api_key":           "flomo_web",
		"app_version":       "4.0",
		"platform":          "web",
		"webp":              "1",
	}
	sign := getSign(n)
	params := url.Values{}
	params.Set("limit", n["limit"])
	params.Set("latest_updated_at", n["latest_updated_at"])
	params.Set("latest_slug", n["latest_slug"])
	params.Set("tz", n["tz"])
	params.Set("timestamp", n["timestamp"])
	params.Add("api_key", n["api_key"])
	params.Add("app_version", n["app_version"])
	params.Add("platform", n["platform"])
	params.Add("webp", n["webp"])
	params.Add("sign", sign)
	fullURL := baseURL + "?" + params.Encode()
	log.InfoLog(fullURL)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", gToken))

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var respData struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    []struct {
			Content   string   `json:"content"`
			CreatorId int      `json:"creator_id"`
			Source    string   `json:"source"`
			Tags      []string `json:"tags"`
			Pin       int      `json:"pin"`
			CreatedAt string   `json:"created_at"`
			UpdatedAt string   `json:"updated_at"`
			DeletedAt string   `json:"deleted_at"`
			Slug      string   `json:"slug"`
			LinkCount int      `json:"link_count"`
			Files     []any    `json:"files"`
		} `json:"data"`
	}
	err = json.Unmarshal(body, &respData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if respData.Code != 0 {
		return nil, fmt.Errorf("get response with error: code=%d, message=%s", respData.Code, respData.Message)
	}
	ret := make([]*Memo, 0, len(respData.Data))
	for _, data := range respData.Data {
		createdAt, _ := time.ParseInLocation(time.DateTime, data.CreatedAt, time.Local)
		updatedAt, _ := time.ParseInLocation(time.DateTime, data.UpdatedAt, time.Local)
		ret = append(ret, &Memo{
			Slug:      data.Slug,
			Content:   data.Content,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			Tags:      data.Tags,
		})
	}
	return ret, nil
}

func getSign(e map[string]string) string {
	keys := make([]string, 0, len(e))
	for k := range e {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	items := make([]string, 0, len(e))
	for _, key := range keys {
		value := e[key]
		if value != "" {
			items = append(items, fmt.Sprintf("%s=%s", key, value))
		}
	}
	result := fmt.Sprintf("%sdbbc3dd73364b4084c3a69346e0ce2b2", strings.Join(items, "&"))
	result = fmt.Sprintf("%x", md5.Sum([]byte(result)))
	return result
}
