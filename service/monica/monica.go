package monica

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

//	const bodyTemplate = `{
//	   "task_uid": "task:<task>",
//	   "bot_uid": "monica",
//	   "data": {
//	       "conversation_id": "conv:<conv>",
//	       "items": [
//	           {
//	               "item_id": "msg:<msg1>",
//	               "conversation_id": "conv:<conv>",
//	               "item_type": "reply",
//	               "summary": "__RENDER_BOT_WELCOME_MSG__",
//	               "data": {
//	                   "type": "text",
//	                   "content": "__RENDER_BOT_WELCOME_MSG__"
//	               }
//	           },
//	           {
//	               "conversation_id": "conv:<conv>",
//	               "item_id": "msg:<msg2>",
//	               "item_type": "question",
//	               "parent_item_id": "msg:<msg1>",
//	               "data": {
//	                   "type": "text",
//	                   "content": <q>,
//	                   "quote_content": "",
//	                   "max_token": 0,
//	                   "is_incognito": false
//	               }
//	           }
//	       ],
//	       "pre_generated_reply_id": "msg:<msg3>",
//	       "pre_parent_item_id": "msg:<msg2>",
//	       "origin": "https://monica.im/home/chat/Monica/monica",
//	       "origin_page_title": "New Chat",
//	       "trigger_by": "auto",
//	       "use_model": "<model>",
//	       "is_incognito": false,
//	       "use_new_memory": true,
//	       "use_memory_suggestion": true
//	   },
//	   "language": "auto",
//	   "locale": "zh_CN",
//	   "task_type": "chat_with_custom_bot",
//	   "tool_data": {},
//	   "ai_resp_language": "Chinese (Simplified)"
//	}`

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
	ConversationId      string `json:"conversation_id"`        // conv:<conv>
	PreGeneratedReplyId string `json:"pre_generated_reply_id"` // msg:<msg3>
	PreParentItemId     string `json:"pre_parent_item_id"`     // msg:<msg2>
	Origin              string `json:"origin"`                 // https://monica.im/home/chat/Monica/monica
	OriginPageTitle     string `json:"origin_page_title"`      // New Chat
	TriggerBy           string `json:"trigger_by"`             // auto
	UseModel            string `json:"use_model"`              // <model>
	IsIncognito         bool   `json:"is_incognito"`           // false
	UseNewMemory        bool   `json:"use_new_memory"`         // true
	UseMemorySuggestion bool   `json:"use_memory_suggestion"`  // true
	Items               []BodyDataItem
}

type BodyDataItem struct {
	ItemId         string           `json:"item_id"`           // msg:<msg1> / msg:<msg2>
	ConversationId string           `json:"conversation_id"`   // conv:<conv> / conv:<conv>
	ItemType       string           `json:"item_type"`         // reply / question
	Summary        string           `json:"summary,omitempty"` // __RENDER_BOT_WELCOME_MSG__ / ""
	Data           BodyDataItemData `json:"data"`
}

type BodyDataItemData struct {
	Type         string `json:"text"`          // text / text
	Content      string `json:"content"`       // __RENDER_BOT_WELCOME_MSG__ / <q>
	QuoteContent string `json:"quote_content"` // nil / ""
	MaxToken     int    `json:"max_token"`     // nil / 0
	IsIncognito  bool   `json:"is_incognito"`  // nil / false
}

func buildBody(model, question string) string {
	taskId := fmt.Sprintf("task:%s", uuid.New().String())
	convId := fmt.Sprintf("conv:%s", uuid.New().String())
	msg1Id := fmt.Sprintf("msg:%s", uuid.New().String())
	msg2Id := fmt.Sprintf("msg:%s", uuid.New().String())
	msg3Id := fmt.Sprintf("msg:%s", uuid.New().String())
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
						Type:         "text",
						Content:      "__RENDER_BOT_WELCOME_MSG__",
						QuoteContent: "",
						MaxToken:     0,
						IsIncognito:  false,
					},
				},
				{
					ItemId:         msg2Id,
					ConversationId: convId,
					ItemType:       "question",
					Summary:        "",
					Data: BodyDataItemData{
						Type:         "text",
						Content:      question,
						QuoteContent: "",
						MaxToken:     0,
						IsIncognito:  false,
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
	b, _ := json.MarshalIndent(body, "", "  ")
	return string(b)
}

var gModel, gSessionId string

func Init(model, sessionId string) {
	gModel = model
	gSessionId = sessionId
}

func ChatCompletion(q string, f func(string)) (string, error) {
	if gSessionId == "" {
		return "", fmt.Errorf("gSessionId is empty, call Init() first")
	}

	url := "https://api.monica.im/api/custom_bot/chat"
	body := buildBody(gModel, q)
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("fail to create request: %v", err)
	}
	req.Header.Add("Cookie", "session_id="+gSessionId)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	client := &http.Client{}
	var resp *http.Response
	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
	}
	if err != nil {
		return "", fmt.Errorf("fail to do client request: %v", err)
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
				Text string `json:"text"`
			}
			err = json.Unmarshal([]byte(data), &chunk)
			if err != nil {
				return "", fmt.Errorf("fail to unmarshal json: %v", err)
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
				return "", fmt.Errorf("fail to unmarshal json: %v", err)
			}
			return "", fmt.Errorf("get monica response with error: %d, %s", respBody.Code, respBody.Msg)
		}
	}
	err = scanner.Err()
	if err != nil {
		return "", fmt.Errorf("fail to scan body: %v", err)
	}
	return sb.String(), nil
}
