package monica

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lincaiyong/uniapi/utils"
	"net/http"
	"strings"
)

type Body struct {
	TaskUid        string   `json:"task_uid"` // task:<task>
	BotUid         string   `json:"bot_uid"`  // monica
	Data           BodyData `json:"data"`
	Language       string   `json:"language"`         // auto
	Locale         string   `json:"locale"`           // zh_CN
	TaskType       string   `json:"task_type"`        // chat_with_custom_bot
	ToolData       struct{} `json:"tool_data"`        // {}
	AiRespLanguage string   `json:"ai_resp_language"` // Chinese (Simplified)
}

type BodyData struct {
	ConversationId      string         `json:"conversation_id"`        // conv:<conv>
	PreGeneratedReplyId string         `json:"pre_generated_reply_id"` // msg:<msg3>
	PreParentItemId     string         `json:"pre_parent_item_id"`     // msg:<msg2>
	Origin              string         `json:"origin"`                 // https://monica.im/home/chat/Monica/monica
	OriginPageTitle     string         `json:"origin_page_title"`      // New Chat
	TriggerBy           string         `json:"trigger_by"`             // auto
	UseModel            string         `json:"use_model"`              // <model>
	IsIncognito         bool           `json:"is_incognito"`           // false
	UseNewMemory        bool           `json:"use_new_memory"`         // true
	UseMemorySuggestion bool           `json:"use_memory_suggestion"`  // true
	Items               []BodyDataItem `json:"items"`
}

type BodyDataItem struct {
	ItemId         string           `json:"item_id"`                  // msg:<msg1> / msg:<msg2>
	ConversationId string           `json:"conversation_id"`          // conv:<conv> / conv:<conv>
	ItemType       string           `json:"item_type"`                // reply / question
	Summary        string           `json:"summary,omitempty"`        // __RENDER_BOT_WELCOME_MSG__ / nil
	ParentItemId   string           `json:"parent_item_id,omitempty"` // nil / msg:<msg1>
	Data           BodyDataItemData `json:"data"`
}

type BodyDataItemData struct {
	Type         string  `json:"type"`                    // text / text
	Content      string  `json:"content"`                 // __RENDER_BOT_WELCOME_MSG__ / <q>
	QuoteContent *string `json:"quote_content,omitempty"` // nil / ""
	MaxToken     *int    `json:"max_token,omitempty"`     // nil / 0
	IsIncognito  *bool   `json:"is_incognito,omitempty"`  // nil / false
}

func buildBody(model, question string) string {
	taskId := fmt.Sprintf("task:%s", uuid.New().String())
	convId := fmt.Sprintf("conv:%s", uuid.New().String())
	msg1Id := fmt.Sprintf("msg:<%s>", uuid.New().String())
	msg2Id := fmt.Sprintf("msg:<%s>", uuid.New().String())
	msg3Id := fmt.Sprintf("msg:<%s>", uuid.New().String())
	body := Body{
		TaskUid: taskId,
		BotUid:  "monica",
		Data: BodyData{
			ConversationId:      convId,
			PreGeneratedReplyId: msg3Id,
			PreParentItemId:     msg2Id,
			Origin:              "https://monica.im/home/chat/Monica/monica",
			OriginPageTitle:     "New Chat",
			TriggerBy:           "auto",
			UseModel:            model,
			IsIncognito:         false,
			UseNewMemory:        true,
			UseMemorySuggestion: true,
			Items: []BodyDataItem{
				{
					ItemId:         msg1Id,
					ConversationId: convId,
					ItemType:       "reply",
					Summary:        "__RENDER_BOT_WELCOME_MSG__",
					Data: BodyDataItemData{
						Type:    "text",
						Content: "__RENDER_BOT_WELCOME_MSG__",
					},
				},
				{
					ItemId:         msg2Id,
					ConversationId: convId,
					ItemType:       "question",
					ParentItemId:   msg1Id,
					Data: BodyDataItemData{
						Type:         "text",
						Content:      question,
						QuoteContent: new(string),
						MaxToken:     new(int),
						IsIncognito:  new(bool),
					},
				},
			},
		},
		Language:       "auto",
		Locale:         "zh_CN",
		TaskType:       "chat_with_custom_bot",
		ToolData:       struct{}{},
		AiRespLanguage: "Chinese (Simplified)",
	}
	b, _ := utils.MarshalIndentNoEscape(body, "", "    ")
	return string(b)
}

var gSessionId string

func Init(sessionId string) {
	gSessionId = sessionId
}

func ChatCompletion(ctx context.Context, model, q string, f func(string)) (string, error) {
	if gSessionId == "" {
		return "", fmt.Errorf("gSessionId is empty, call Init() first")
	}

	url := "https://api.monica.im/api/custom_bot/chat"
	body := buildBody(model, q)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("fail to create request: %w", err)
	}
	req.Header.Add("Cookie", "session_id="+gSessionId)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}}
	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return "", fmt.Errorf("fail to do client request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get http response with error: %d, %s", resp.StatusCode, resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	var sb strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data:") {
			data := strings.Trim(line[5:], " ")
			var chunk struct {
				Text  string `json:"text"`
				Error any    `json:"error"`
			}
			err = json.Unmarshal([]byte(data), &chunk)
			if err != nil {
				return "", fmt.Errorf("fail to unmarshal json: %w", err)
			}
			if chunk.Error != nil {
				return "", fmt.Errorf("read chunk error: %v", chunk.Error)
			}
			f(chunk.Text)
			sb.WriteString(chunk.Text)
		} else if line != "" {
			var respBody struct {
				Code int    `json:"code"`
				Msg  string `json:"msg"`
			}
			err = json.Unmarshal([]byte(line), &respBody)
			if err != nil {
				return "", fmt.Errorf("fail to unmarshal json: %w", err)
			}
			return "", fmt.Errorf("get monica response with error: %d, %s", respBody.Code, respBody.Msg)
		}
	}
	err = scanner.Err()
	if err != nil {
		return "", fmt.Errorf("fail to scan body: %w", err)
	}
	return sb.String(), nil
}
