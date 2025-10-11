package baidupan

import (
	"crypto/sha1"
	"encoding/hex"
	"path"
)

func sha1Of(b []byte) string {
	hash := sha1.New()
	hash.Write(b)
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum[:])
}

func pathOf(sha1 string) string {
	return path.Join("/object", sha1[:2], sha1[2:4], sha1[4:6], sha1[6:8], sha1[8:])
}

func PutObject(data []byte) (string, error) {
	hash := sha1Of(data)
	filePath := pathOf(hash)
	_, err := GetFileId(filePath)
	if err != nil {
		err = Upload(filePath, data)
		if err != nil {
			return "", err
		}
	}
	return hash, nil
}

func GetObject(hash string) ([]byte, error) {
	filePath := pathOf(hash)
	b, err := Download(filePath)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func HasObject(hash string) bool {
	filePath := pathOf(hash)
	_, err := GetFileId(filePath)
	return err == nil
}
