package baidupan

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ListDirItem struct {
	FsId int64  `json:"fs_id"`
	Path string `json:"path"`
	Md5  string `json:"md5"`
	Size int64  `json:"size"`
}

func listDir(dir string) ([]*ListDirItem, error) {
	if !strings.HasPrefix(dir, "/") {
		return nil, fmt.Errorf("invalid file path: %s, should start with \"/\"", dir)
	}
	dirQuoted := url.QueryEscape(dir)
	panUrl := fmt.Sprintf("https://pan.baidu.com/api/list?app_id=250528&dir=%s&page=1&num=100", dirQuoted)
	req, err := http.NewRequest("GET", panUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %w", err)
	}
	req.Header.Set("Cookie", cookieValue())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fail to download url: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fail to list dir, status: %d", resp.StatusCode)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to io read response body: %w", err)
	}
	var result struct {
		Errno int            `json:"errno"`
		List  []*ListDirItem `json:"list"`
	}
	if err = json.Unmarshal(bs, &result); err != nil {
		return nil, fmt.Errorf("fail to unmarshal response body: %w", err)
	}
	if result.Errno != 0 {
		return nil, fmt.Errorf("fail to list dir, code: %d", result.Errno)
	}
	return result.List, nil
}
