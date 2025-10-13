package baidupan

import (
	"context"
	"fmt"
	"github.com/lincaiyong/log"
	"path"
)

var gBdUss string
var gSToken string

func Init(bdUss, stoken string) {
	gBdUss = bdUss
	gSToken = stoken
}

func cookieValue() string {
	if gBdUss == "" || gSToken == "" {
		return ""
	}
	return fmt.Sprintf("BDUSS=%s; STOKEN=%s", gBdUss, gSToken)
}

func GetFileId(ctx context.Context, filePath string) (fileId int64, err error) {
	dirPath := path.Dir(filePath)
	items, err := listDir(ctx, dirPath)
	if err != nil {
		return 0, fmt.Errorf("fail to list dir, %w", err)
	}
	for _, item := range items {
		if item.Path == filePath {
			return item.FsId, nil
		}
	}
	return 0, fmt.Errorf("file not exists")
}

func Download(ctx context.Context, filePath string) ([]byte, error) {
	log.InfoLog("download file: %s", filePath)
	if cookieValue() == "" {
		return nil, fmt.Errorf("cookie is empty, should call Init() first")
	}
	fileId, err := GetFileId(ctx, filePath)
	if err != nil {
		return nil, err
	}
	link, err := getDownloadLink(ctx, fileId)
	if err != nil {
		return nil, err
	}
	return downloadByLink(ctx, link)
}

func Upload(ctx context.Context, filePath string, content []byte) error {
	log.InfoLog("upload file: %s", filePath)
	if cookieValue() == "" {
		return fmt.Errorf("cookie is empty, should call Init() first")
	}
	_, err := GetFileId(ctx, filePath)
	if err == nil {
		log.InfoLog("upload file: %s already exists, delete it", filePath)
		err = deleteFile(ctx, filePath)
		if err != nil {
			return err
		}
	}

	size := len(content)
	hash := calcMd5(content)
	uploadId, err := uploadPreCreate(ctx, filePath, hash)
	if err != nil {
		return err
	}
	err = uploadSuperFile(ctx, filePath, uploadId, content)
	if err != nil {
		return err
	}
	err = uploadCreate(ctx, filePath, uploadId, hash, size)
	return err
}
