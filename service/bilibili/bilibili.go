package bilibili

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/log"
	"github.com/lincaiyong/uniapi/utils"
	"net/http"
)

/*
GET /x/relation/followings?vmid=109372877 HTTP/2
Host: api.bilibili.com
Cookie: SESSDATA=71b6fef9%2C1776694329%2C9225a%2Aa2CjDNVpNYD54IOCWHBZdCKveo3_pw8CzLxIEA30IvnukLBjs6ZuzeC6dvXjna7A9V7YwSVjRKRVVJRDluM0NKejdoUVlfNHNkbFZwTlpLenpoYV9KajNuVURMYWgtQTItdkRwTFFiMVRleGZXeXhMSUlid2V3cl90RF91aWxMR21yV1NoR3pXeWV3IIEC;
*/

var sessData, biliJCT, vmId string

func Init(sessData_, biliJCT_, vmId_ string) {
	sessData = sessData_
	biliJCT = biliJCT_
	vmId = vmId_
}

func doRequest(ctx context.Context, method string, url string, body []byte) ([]byte, error) {
	headers := map[string]string{
		"Cookie": fmt.Sprintf("SESSDATA=%s", sessData),
	}
	if body != nil {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
	}
	return utils.DoRequest(ctx, method, url, headers, body)
}

type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type FollowingResponse struct {
	BaseResponse
	Data struct {
		List  []*FollowingInfo `json:"list"`
		Total int              `json:"total"`
	} `json:"data"`
}
type FollowingInfo struct {
	Mid   int    `json:"mid"`
	Uname string `json:"uname"`
	Face  string `json:"face"`
	Sign  string `json:"sign"`
}

func getFollowingsByPage(ctx context.Context, pageNo int) ([]*FollowingInfo, int, error) {
	reqUrl := fmt.Sprintf("https://api.bilibili.com/x/relation/followings?vmid=%s&pn=%d&ps=24", vmId, pageNo)
	b, err := doRequest(ctx, "GET", reqUrl, nil)
	if err != nil {
		log.ErrorLog("fail to get followings: %v", err)
		return nil, 0, err
	}
	var resp FollowingResponse
	err = json.Unmarshal(b, &resp)
	if err != nil {
		log.ErrorLog("fail to unmarshal response data: %v", err)
		return nil, 0, err
	}
	if resp.Code != 0 {
		log.ErrorLog("fail to get followings: %d, %s", resp.Code, resp.Message)
		return nil, 0, fmt.Errorf("get response with error: %d, %s", resp.Code, resp.Message)
	}
	return resp.Data.List, resp.Data.Total, nil
}

func GetAllFollowings(ctx context.Context) ([]*FollowingInfo, error) {
	pageNo := 1
	followings, total, err := getFollowingsByPage(ctx, pageNo)
	if err != nil {
		return nil, err
	}
	ret := followings
	for len(ret) < total {
		pageNo++
		followings, _, err = getFollowingsByPage(ctx, pageNo)
		if err != nil {
			return nil, err
		}
		ret = append(ret, followings...)
	}
	return ret, nil
}

func CancelFollow(ctx context.Context, fid string) error {
	reqUrl := "https://api.bilibili.com/x/relation/modify"
	s := fmt.Sprintf(`fid=%s&act=2&csrf=%s`, fid, biliJCT)
	b, err := doRequest(ctx, http.MethodPost, reqUrl, []byte(s))
	if err != nil {
		return err
	}
	var resp BaseResponse
	err = json.Unmarshal(b, &resp)
	if err != nil {
		log.ErrorLog("fail to unmarshal response data: %v", err)
		return err
	}
	if resp.Code != 0 {
		log.ErrorLog("fail to cancel followings: %d, %s", resp.Code, resp.Message)
		return fmt.Errorf("fail to cancel followings: %d, %s", resp.Code, resp.Message)
	}
	return nil
}
