package utils

import (
	"net/http"
)

func Retry(count int, f func() error) error {
	var err error
	for i := 0; i < count; i++ {
		err = f()
		if err == nil {
			return nil
		}
	}
	return err
}

func ClientDoRetry(client *http.Client, req *http.Request, retryCount int) (*http.Response, error) {
	var resp *http.Response
	err := Retry(retryCount, func() error {
		var err error
		resp, err = client.Do(req)
		if err != nil && resp != nil {
			_ = resp.Body.Close()
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
