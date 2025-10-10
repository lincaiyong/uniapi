package object

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

var gObjectServerUrl string
var gObjectServerToken string

func Init(objectServerUrl, objectServerToken string) {
	gObjectServerUrl = objectServerUrl
	gObjectServerToken = objectServerToken
}

func Put(data []byte) (string, error) {
	if gObjectServerUrl == "" || gObjectServerToken == "" {
		return "", fmt.Errorf("object server url or token is empty")
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "file")
	if err != nil {
		return "", err
	}
	_, err = part.Write(data)
	if err != nil {
		return "", err
	}
	err = writer.Close()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/object/put?token=%s", gObjectServerUrl, gObjectServerToken)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%d, %s", resp.StatusCode, string(b))
	}
	return string(b), nil
}

func Get(sha1 string) ([]byte, error) {
	if gObjectServerUrl == "" || gObjectServerToken == "" {
		return nil, fmt.Errorf("object server url or token is empty")
	}
	url := fmt.Sprintf("%s/object/get/%s?token=%s", gObjectServerUrl, sha1, gObjectServerToken)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%d, %s", resp.StatusCode, resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}
