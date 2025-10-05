package baidupan

import (
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/uniapi/utils"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func deleteFile(filePath string) error {
	taskId, err := createDeleteFileTask(filePath)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}

	for i := 0; i < 10; i++ {
		var done bool
		done, err = queryDeleteTaskStatus(taskId)
		if err != nil {
			return fmt.Errorf("query task failed: %w", err)
		}
		if done {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("delete task timeout")
}

func createDeleteFileTask(filePath string) (int64, error) {
	reqUrl := fmt.Sprintf("https://pan.baidu.com/api/filemanager?async=2&onnest=fail&opera=delete&bdstoken=%s&newVerify=1&clienttype=0&app_id=250528&web=1", gSToken)
	data := url.Values{}
	data.Set("filelist", fmt.Sprintf(`["%s"]`, filePath))
	req, err := http.NewRequest("POST", reqUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return 0, fmt.Errorf("fail to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", cookieValue())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return 0, fmt.Errorf("fail to do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("fail to read response body: %w", err)
	}

	var deleteResp struct {
		Errno     int   `json:"errno"`
		RequestID int64 `json:"request_id"`
		TaskID    int64 `json:"taskid"`
	}
	err = json.Unmarshal(body, &deleteResp)
	if err != nil {
		return 0, fmt.Errorf("fail to unmarshal response body: %w", err)
	}

	if deleteResp.Errno != 0 {
		return 0, fmt.Errorf("delete failed with errno: %d", deleteResp.Errno)
	}

	return deleteResp.TaskID, nil
}

func queryDeleteTaskStatus(taskID int64) (bool, error) {
	reqUrl := fmt.Sprintf("https://pan.baidu.com/share/taskquery?taskid=%d", taskID)
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return false, fmt.Errorf("fail to create request: %w", err)
	}
	req.Header.Set("Cookie", cookieValue())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := utils.ClientDoRetry(client, req, 3)
	if err != nil {
		return false, fmt.Errorf("fail to do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("fail to read response body: %w", err)
	}

	var taskResp struct {
		Errno     int    `json:"errno"`
		RequestID int64  `json:"request_id"`
		TaskErrno int    `json:"task_errno"`
		Status    string `json:"status"`
		Total     int    `json:"total"`
	}
	err = json.Unmarshal(body, &taskResp)
	if err != nil {
		return false, fmt.Errorf("fail to unmarshal response body: %w", err)
	}

	if taskResp.Errno != 0 {
		return false, fmt.Errorf("task query failed with errno: %d", taskResp.Errno)
	}

	return taskResp.Status == "success", nil
}
