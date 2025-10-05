package googletrans

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lincaiyong/uniapi/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func TranslateEnToZh(text string) (string, error) {
	ret, err := translate(text, "en", "zh")
	if err != nil {
		return "", err
	}
	return ret, nil
}

func TranslateZhToEn(text string) (string, error) {
	ret, err := translate(text, "zh", "en")
	if err != nil {
		return "", err
	}
	return ret, nil
}

func translate(text, sourceLang, targetLang string) (string, error) {
	if text == "" {
		return "", errors.New("empty text")
	}

	baseURL := "https://translate.googleapis.com/translate_a/single"

	params := url.Values{}
	params.Set("client", "gtx")
	params.Set("sl", sourceLang)
	params.Set("tl", targetLang)
	params.Set("hl", targetLang)
	params.Set("dt", "at")
	params.Add("dt", "bd")
	params.Add("dt", "ex")
	params.Add("dt", "ld")
	params.Add("dt", "md")
	params.Add("dt", "qca")
	params.Add("dt", "rw")
	params.Add("dt", "rm")
	params.Add("dt", "ss")
	params.Add("dt", "t")
	params.Set("ie", "UTF-8")
	params.Set("oe", "UTF-8")
	params.Set("otf", "1")
	params.Set("ssel", "0")
	params.Set("tsel", "0")
	params.Set("tk", "xxxx")
	params.Set("q", text)
	fullURL := baseURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	translatedText, err := parseGoogleTranslateResponse(body)
	if err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return translatedText, nil
}

func parseGoogleTranslateResponse(body []byte) (string, error) {
	var result []any
	err := json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	if len(result) == 0 {
		return "", fmt.Errorf("empty response")
	}
	firstElement, ok := result[0].([]any)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	var translatedParts []string
	for _, part := range firstElement {
		if partArray, ok := part.([]any); ok && len(partArray) > 0 {
			if translatedText, ok := partArray[0].(string); ok {
				translatedParts = append(translatedParts, translatedText)
			}
		}
	}
	if len(translatedParts) == 0 {
		return "", fmt.Errorf("no translation found in response")
	}

	return strings.Join(translatedParts, ""), nil
}
