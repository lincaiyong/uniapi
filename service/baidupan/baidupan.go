package baidupan

import (
	"errors"
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
	log.InfoLog("download file: %s", filePath)
	if cookieValue() == "" {
		return nil, fmt.Errorf("cookie is empty, should call Init() first")
	}
	fileId, err := getFileId(filePath)
	if err != nil {
		return nil, err
	}
	link, err := getDownloadLink(fileId)
	if err != nil {
		return nil, err
	}
	return downloadByLink(link)
}

func Upload(filePath string, content []byte) error {
	log.InfoLog("upload file: %s", filePath)
	if cookieValue() == "" {
		return fmt.Errorf("cookie is empty, should call Init() first")
	}
	_, err := getFileId(filePath)
	if err == nil || !errors.Is(err, fileNotFoundError) {
		log.InfoLog("upload file: %s already exists, delete it", filePath)
		err = deleteFile(filePath)
		if err != nil {
			return err
		}
	}

	size := len(content)
	hash := calcMd5(content)
	uploadId, err := uploadPreCreate(filePath, hash)
	if err != nil {
		return err
	}
	err = uploadSuperFile(filePath, uploadId, content)
	if err != nil {
		return err
	}
	err = uploadCreate(filePath, uploadId, hash, size)
	return err
}
