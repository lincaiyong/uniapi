package utils

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

func DoRequestWithResp(ctx context.Context, method string, url string, header map[string]string, body []byte) (*http.Response, error) {
	var b io.Reader
	if body != nil {
		b = bytes.NewBuffer(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, b)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	return ClientDoRetry(client, req, 3)
}

func DoRequest(ctx context.Context, method string, url string, header map[string]string, body []byte) ([]byte, error) {
	resp, err := DoRequestWithResp(ctx, method, url, header, body)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return b, nil
}
