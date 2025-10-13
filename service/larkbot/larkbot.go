package larkbot

import (
	"context"
	"encoding/json"
	"fmt"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

// https://open.feishu.cn/document/uAjLw4CM/ukTMukTMukTM/server-side-sdk/golang-sdk-guide/preparations

var gAppId, gAppSecret, gReceiveId string

func Init(appId, appSecret, receiveId string) {
	gAppId = appId
	gAppSecret = appSecret
	gReceiveId = receiveId
}

func Send(ctx context.Context, msg string) error {
	if gAppId == "" || gAppSecret == "" || gReceiveId == "" {
		return fmt.Errorf("app id or app secret or receive id is empty, call Init() first")
	}
	client := lark.NewClient(gAppId, gAppSecret)
	b, _ := json.Marshal(map[string]string{
		"text": msg,
	})
	content := string(b)
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(`open_id`).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(gReceiveId).
			MsgType(`text`).
			Content(content).
			Build()).
		Build()
	resp, err := client.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("fail to create message: %v", err)
	}
	if !resp.Success() {
		return fmt.Errorf("unexpected response: %s", larkcore.Prettify(resp.CodeError))
	}
	return nil
}

func SendTo(ctx context.Context, msg, chatId string) error {
	if gAppId == "" || gAppSecret == "" {
		return fmt.Errorf("app id or app secret or receive id is empty, call Init() first")
	}
	client := lark.NewClient(gAppId, gAppSecret)
	b, _ := json.Marshal(map[string]string{
		"text": msg,
	})
	content := string(b)
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType(`chat_id`).
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(chatId).
			MsgType(`text`).
			Content(content).
			Build()).
		Build()
	resp, err := client.Im.V1.Message.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("fail to create message: %v", err)
	}
	if !resp.Success() {
		return fmt.Errorf("unexpected response: %s", larkcore.Prettify(resp.CodeError))
	}
	return nil
}
