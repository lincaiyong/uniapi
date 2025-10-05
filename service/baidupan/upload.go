package baidupan

import (
	"bytes"
	"clitool/util"
	"clitool/util/cli"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func uploadSuperFile(savePath, uploadId string, bs []byte) {
	cli.Info("upload super file: %s", uploadId)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", "blob")
	if err != nil {
		cli.Fatal("fail to create form file: %v", err)
	}
	_, err = io.WriteString(fileWriter, string(bs))
	if err != nil {
		cli.Fatal("fail to write form file: %v", err)
	}
	err = writer.Close()
	if err != nil {
		cli.Fatal("fail to close form file: %v", err)
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
		cli.Fatal("fail to create request: %v", err)
	}
	req.Header.Set("Cookie", cookieValue())
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		cli.Fatal("fail to do request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		cli.Fatal("fail to do request, status: %d", resp.StatusCode)
	}
}

type preCreateResponse struct {
	UploadId string `json:"uploadId"`
	Errno    int    `json:"errno"`
}

func uploadPreCreate(savePath, md5 string) string {
	cli.Info("pre create: %s, %s", savePath, md5)
	fullUrl := fmt.Sprintf("https://pan.baidu.com/api/precreate")
	formData := url.Values{}
	formData.Set("path", savePath)
	formData.Set("autoinit", "1")
	formData.Set("block_list", fmt.Sprintf(`["%s"]`, md5))
	req, err := http.NewRequest("POST", fullUrl, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		cli.Fatal("fail to create request: %v", err)
	}
	req.Header.Set("Cookie", cookieValue())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		cli.Fatal("fail to do request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		cli.Fatal("unexpected response code: %d, %s", resp.StatusCode, resp.Status)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		cli.Fatal("fail to read response body: %v", err)
	}
	var respJson preCreateResponse
	err = json.Unmarshal(bs, &respJson)
	if err != nil {
		cli.Fatal("fail to unmarshal response body: %v", err)
	}
	if respJson.Errno != 0 {
		cli.Fatal("unexpected response errno: %d", respJson.Errno)
	}
	return respJson.UploadId
}

func uploadCreate(savePath, uploadId, md5 string, size int) {
	/*
		POST /api/create?isdir=0&app_id=250528&channel=chunlei&web=1&clienttype=0 HTTP/1.1
		Host: pan.baidu.com
		Cookie: BDUSS=2RTS1pZQjQwYjBiMFZJRms0OVZqdUp4ZzluQzZQaUJEWkVNNzlYYzlrYUNnS1JvSVFBQUFBJCQAAAAAAAAAAAEAAACObyE0WUlOR0hVQVNIVVhJQV8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAILzfGiC83xoY; STOKEN=2dfcfa40665f98d3c1580ed6edc45f46d7581d4386e81b9b3725db97390cf8ab
		Content-Type: application/x-www-form-urlencoded
		Content-Length: 197

		path=%2Ftest4.txt&size=12&uploadid=N1-MTYzLjEyNS43Ny4xNTU6MTc1MzAyODUxNjo1MDQyNDgyOTM4MTgzMTY2NA%3d%3d&block_list=%5B%226f5902ac237024bdd0c176cb93063dc4%22%5D
	*/
	cli.Info("pre create: %s, %s", savePath, md5)
	fullUrl := fmt.Sprintf("https://pan.baidu.com/api/create?isdir=0&app_id=250528&channel=chunlei&web=1&clienttype=0")
	formData := url.Values{}
	formData.Set("path", savePath)
	formData.Set("size", strconv.Itoa(size))
	formData.Set("uploadid", uploadId)
	formData.Set("block_list", fmt.Sprintf(`["%s"]`, md5))
	req, err := http.NewRequest("POST", fullUrl, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		cli.Fatal("fail to create request: %v", err)
	}
	req.Header.Set("Cookie", cookieValue())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		cli.Fatal("fail to do request: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		cli.Fatal("unexpected response code: %d, %s", resp.StatusCode, resp.Status)
	}
	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		cli.Fatal("fail to read response body: %v", err)
	}
	var respJson preCreateResponse
	err = json.Unmarshal(bs, &respJson)
	if err != nil {
		cli.Fatal("fail to unmarshal response body: %v", err)
	}
	if respJson.Errno != 0 {
		cli.Fatal("unexpected response errno: %d", respJson.Errno)
	}
}

func Upload(filePath, fileName string) string {
	cli.Info("upload file: %s -> %s", filePath, fileName)
	bs, err := os.ReadFile(filePath)
	if err != nil {
		cli.Fatal("fail to read file: %v", err)
	}
	size := len(bs)
	md5 := util.Md5(bs)
	savePath := fmt.Sprintf("/goodfun/%s", fileName)
	uploadId := uploadPreCreate(savePath, md5)
	uploadSuperFile(savePath, uploadId, bs)
	uploadCreate(savePath, uploadId, md5, size)
	return ""
}
