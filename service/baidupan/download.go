package baidupan

import (
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/log"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func downloadByLink(url string) ([]byte, error) {
	log.InfoLog("download link: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %w", err)
	}
	req.Header.Set("Cookie", cookieValue())
	client := &http.Client{}
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

func getTemplateVariable() (sign1, sign2 string, timestamp int, err error) {
	params := url.Values{}
	params.Add("fields", `["sign1","timestamp","sign3"]`)
	fullUrl := fmt.Sprintf("https://pan.baidu.com/api/gettemplatevariable?%s", params.Encode())
	var req *http.Request
	req, err = http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		err = fmt.Errorf("fail to create request: %w", err)
		return
	}
	req.Header.Set("Cookie", cookieValue())
	client := &http.Client{}
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

func getSignTimestamp() (string, int, error) {
	sign1, sign3, timestamp, err := getTemplateVariable()
	if err != nil {
		return "", 0, err
	}
	sign := panSign2(sign1, sign3)
	return sign, timestamp, nil
}

func downloadByFileId(fileId int64) ([]byte, error) {
	log.InfoLog("download file by id: %d", fileId)
	sign, timestamp, err := getSignTimestamp()
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("fidlist", fmt.Sprintf(`[%d]`, fileId))
	params.Add("sign", sign)
	params.Add("timestamp", strconv.Itoa(timestamp))
	fullUrl := fmt.Sprintf("https://pan.baidu.com/api/download?%s", params.Encode())
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("fail to create request: %v", err)
	}
	req.Header.Set("Cookie", cookieValue())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fail to do request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response code: %d, %s", resp.StatusCode, resp.Status)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("fail to read response body: %v", err)
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
		return nil, fmt.Errorf("fail to unmarshal response body: %v", err)
	}
	if respJson.Errno != 0 {
		return nil, fmt.Errorf("unexpected response errno: %d", respJson.Errno)
	}
	if len(respJson.DownloadLink) != 1 {
		return nil, fmt.Errorf("unexpected dlink array size: %d", len(respJson.DownloadLink))
	}
	link := respJson.DownloadLink[0].Link
	link = strings.ReplaceAll(link, `\/`, "/")
	return downloadByLink(link)
}
