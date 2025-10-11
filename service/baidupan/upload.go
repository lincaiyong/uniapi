package baidupan

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/log"
	"github.com/lincaiyong/uniapi/utils"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
)

func uploadSuperFile(savePath, uploadId string, bs []byte) error {
	log.InfoLog("upload super file: %s", uploadId)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "blob")
	if err != nil {
		return fmt.Errorf("fail to create form file: %w", err)
	}
	_, err = io.WriteString(fileWriter, string(bs))
	if err != nil {
		return fmt.Errorf("fail to write form file: %w", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("fail to close form file: %w", err)
	}
	params := url.Values{}
	params.Set("method", "upload")
	params.Set("app_id", "250528")
	params.Set("channel", "chunlei")
	params.Set("web", "1")
	params.Set("clienttype", "0")
	params.Set("path", savePath)
	params.Set("uploadid", uploadId)
	params.Set("uploadsign", "0")
	params.Set("partseq", "0")
	fullUrl := fmt.Sprintf("https://c5.pcs.baidu.com/rest/2.0/pcs/superfile2?%s", params.Encode())
	req, err := http.NewRequest("POST", fullUrl, body)
	if err != nil {
		return fmt.Errorf("fail to create request: %v", err)
	}
	req.Header.Set("Cookie", cookieValue())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return fmt.Errorf("fail to do request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to do request, status: %d", resp.StatusCode)
	}
	return nil
}

func uploadPreCreate(savePath, md5 string) (string, error) {
	log.InfoLog("pre create: %s, %s", savePath, md5)
	fullUrl := fmt.Sprintf("https://pan.baidu.com/api/precreate")
	formData := url.Values{}
	formData.Set("path", savePath)
	formData.Set("autoinit", "1")
	formData.Set("block_list", fmt.Sprintf(`["%s"]`, md5))
	req, err := http.NewRequest("POST", fullUrl, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("fail to create request: %v", err)
	}
	req.Header.Set("Cookie", cookieValue())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fail to do request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected response code: %d, %s", resp.StatusCode, resp.Status)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fail to read response body: %v", err)
	}
	var respJson struct {
		UploadId string `json:"uploadId"`
		Errno    int    `json:"errno"`
	}
	err = json.Unmarshal(bs, &respJson)
	if err != nil {
		return "", fmt.Errorf("fail to unmarshal response body: %v", err)
	}
	if respJson.Errno != 0 {
		return "", fmt.Errorf("unexpected response errno: %d", respJson.Errno)
	}
	return respJson.UploadId, nil
}

func uploadCreate(savePath, uploadId, md5 string, size int) error {
	log.InfoLog("upload create: %s, %s", savePath, md5)
	fullUrl := fmt.Sprintf("https://pan.baidu.com/api/create?isdir=0&app_id=250528&channel=chunlei&web=1&clienttype=0")
	formData := url.Values{}
	formData.Set("path", savePath)
	formData.Set("size", strconv.Itoa(size))
	formData.Set("uploadid", uploadId)
	formData.Set("block_list", fmt.Sprintf(`["%s"]`, md5))
	req, err := http.NewRequest("POST", fullUrl, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return fmt.Errorf("fail to create request: %v", err)
	}
	req.Header.Set("Cookie", cookieValue())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fail to do request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected response code: %d, %s", resp.StatusCode, resp.Status)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("fail to read response body: %v", err)
	}
	var respJson struct {
		UploadId string `json:"uploadId"`
		Errno    int    `json:"errno"`
	}
	err = json.Unmarshal(bs, &respJson)
	if err != nil {
		return fmt.Errorf("fail to unmarshal response body: %v", err)
	}
	if respJson.Errno != 0 {
		return fmt.Errorf("unexpected response errno: %d", respJson.Errno)
	}
	return nil
}

func calcMd5(b []byte) string {
	hash := md5.New()
	hash.Write(b)
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum[:])
}
