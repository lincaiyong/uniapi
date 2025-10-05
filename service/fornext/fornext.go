package fornext

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/uniapi/utils"
	"net/http"
	"strings"
	"time"
)

var gSpaceId, gModelName, gModelId, gPromptKey, gPromptPlatformSession string

func Init(spaceId, modelName, modelId, promptKey, session string) {
	gSpaceId = spaceId
	gModelName = modelName
	gModelId = modelId
	gPromptKey = promptKey
	gPromptPlatformSession = session
}

type Body struct {
	SpaceId string      `json:"space_id"`
	Prompt  BodyPrompt  `json:"prompt"`
	Message BodyMessage `json:"message"`
}

type BodyPrompt struct {
	ModelConfig BodyPromptModelConfig `json:"model_config"`
	PromptKey   string                `json:"prompt_key"`
}

type BodyPromptModelConfig struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type BodyMessage struct {
	MessageType int    `json:"message_type"`
	Content     string `json:"content"`
}

func buildBody(question string) string {
	body := Body{
		SpaceId: gSpaceId,
		Prompt: BodyPrompt{
			ModelConfig: BodyPromptModelConfig{
				Name: gModelName,
				Id:   gModelId,
			},
			PromptKey: gPromptKey,
		},
		Message: BodyMessage{
			MessageType: 2,
			Content:     question,
		},
	}
	b, _ := json.Marshal(body)
	return string(b)
}

func ChatCompletion(q string, f func(string)) (string, error) {
	url := "https://fornax.bytedance.net/api/devops/prompt_platform/v1/prompt/streaming_send_message"
	body := buildBody(q)
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("fail to create request: %w", err)
	}
	req.Header.Add("Cookie", "prompt-platform-session="+gPromptPlatformSession)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Agw-Js-Conv", "str")

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}}
	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return "", fmt.Errorf("fail to do client request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	scanner := bufio.NewScanner(resp.Body)
	var sb strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			data := strings.Trim(line[5:], " ")
			var chunk struct {
				Item struct {
					Content string `json:"content"`
				} `json:"item"`
			}
			err = json.Unmarshal([]byte(data), &chunk)
			if err != nil {
				return "", fmt.Errorf("fail to unmarshal chunk: %w", err)
			}
			f(chunk.Item.Content)
			sb.WriteString(chunk.Item.Content)
		} else if line != "" {
			return "", fmt.Errorf("get fornext response with unexpected line: %s", line)
		}
	}
	err = scanner.Err()
	if err != nil {
		return "", fmt.Errorf("fail to scan response body: %w", err)
	}
	return sb.String(), nil
}
