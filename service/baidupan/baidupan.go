package baidupan

import (
	"errors"
	"fmt"
	"path"
)

var gBdUss string
var gSToken string

func Init(bduss, stoken string) {
	gBdUss = bduss
	gSToken = stoken
}

func cookieValue() string {
	if gBdUss == "" || gSToken == "" {
		return ""
	}
	return fmt.Sprintf("BDUSS=%s; STOKEN=%s", gBdUss, gSToken)
}

var fileNotFoundError = errors.New("file not found")

func getFileId(filePath string) (fileId int64, err error) {
	dirPath := path.Dir(filePath)
	items, err := listDir(dirPath)
	if err != nil {
		return 0, fmt.Errorf("fail to list dir, %w", err)
	}
	for _, item := range items {
		if item.Path == filePath {
			return item.FsId, nil
		}
	}
	return 0, fileNotFoundError
}

func Download(filePath string) ([]byte, error) {
	if cookieValue() == "" {
		return nil, fmt.Errorf("cookie is empty, should call Init() first")
	}
	fileId, err := getFileId(filePath)
	if err != nil {
		return nil, err
	}
	return downloadByFileId(fileId)
}

func Upload(filePath string, content []byte) error {
	if cookieValue() == "" {
		return fmt.Errorf("cookie is empty, should call Init() first")
	}
	fileId, err := getFileId(filePath)
	if err != nil && !errors.Is(err, fileNotFoundError) {
		return err
	}
	
	return nil
}
