package playwright

import (
	"github.com/lincaiyong/log"
	"regexp"
	"testing"
)

func Test1(t *testing.T) {
	result, err := GetHeader("https://v.flomoapp.com/mine", regexp.MustCompile(`^https://flomoapp.com/api/v1/tag/updated/.+`), "authorization")
	if err != nil {
		t.Fatal(err)
	}
	log.InfoLog(result)
}
