package baidupan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/log"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func downloadByLink(ctx context.Context, url string) ([]byte, error) {
	log.InfoLog("download file by link: %s", url)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
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
		return nil, fmt.Errorf("fail to download url, status: %d", resp.StatusCode)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to io read response body: %w", err)
	}
	return bs, nil
}

func getTemplateVariable(ctx context.Context) (sign1, sign2 string, timestamp int, err error) {
	params := url.Values{}
	params.Add("fields", `["sign1","timestamp","sign3"]`)
	fullUrl := fmt.Sprintf("https://pan.baidu.com/api/gettemplatevariable?%s", params.Encode())
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, "GET", fullUrl, nil)
	if err != nil {
		err = fmt.Errorf("fail to create request: %w", err)
		return
	}
	req.Header.Set("Cookie", cookieValue())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("fail to do request: %w", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("unexpected response code: %d, %s", resp.StatusCode, resp.Status)
		return
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("fail to read response body: %w", err)
		return
	}
	var respJson struct {
		Errno  int `json:"errno"`
		Result struct {
			Sign1     string `json:"sign1"`
			Timestamp int    `json:"timestamp"`
			Sign3     string `json:"sign3"`
		} `json:"result"`
	}
	err = json.Unmarshal(bs, &respJson)
	if err != nil {
		err = fmt.Errorf("fail to unmarshal response body: %w", err)
		return
	}
	if respJson.Errno != 0 {
		err = fmt.Errorf("unexpected response errno: %d", respJson.Errno)
		return
	}
	return respJson.Result.Sign1, respJson.Result.Sign3, respJson.Result.Timestamp, nil
}

func getSignTimestamp(ctx context.Context) (string, int, error) {
	sign1, sign3, timestamp, err := getTemplateVariable(ctx)
	if err != nil {
		return "", 0, err
	}
	sign := panSign2(sign1, sign3)
	return sign, timestamp, nil
}

func getDownloadLink(ctx context.Context, fileId int64) (string, error) {
	log.InfoLog("get download link: %d", fileId)
	sign, timestamp, err := getSignTimestamp(ctx)
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Add("fidlist", fmt.Sprintf(`[%d]`, fileId))
	params.Add("sign", sign)
	params.Add("timestamp", strconv.Itoa(timestamp))
	fullUrl := fmt.Sprintf("https://pan.baidu.com/api/download?%s", params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", fullUrl, nil)
	if err != nil {
		return "", fmt.Errorf("fail to create request: %v", err)
	}
	req.Header.Set("Cookie", cookieValue())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fail to do request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response code: %d, %s", resp.StatusCode, resp.Status)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fail to read response body: %v", err)
	}
	var respJson struct {
		Errno        int `json:"errno"`
		DownloadLink []struct {
			FsID string `json:"fs_id"`
			Link string `json:"dlink"`
		} `json:"dlink"`
	}
	err = json.Unmarshal(bs, &respJson)
	if err != nil {
		return "", fmt.Errorf("fail to unmarshal response body: %v", err)
	}
	if respJson.Errno != 0 {
		return "", fmt.Errorf("unexpected response errno: %d", respJson.Errno)
	}
	if len(respJson.DownloadLink) != 1 {
		return "", fmt.Errorf("unexpected dlink array size: %d", len(respJson.DownloadLink))
	}
	link := respJson.DownloadLink[0].Link
	link = strings.ReplaceAll(link, `\/`, "/")
	return link, nil
}
