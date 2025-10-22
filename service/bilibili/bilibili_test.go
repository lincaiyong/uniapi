package bilibili

import (
	"context"
	"github.com/lincaiyong/log"
	"strconv"
	"strings"
	"testing"
)

func init() {
	Init(
		"",
		"",
		"",
	)
}

func TestGetFollowings(t *testing.T) {
	all, err := GetAllFollowings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range all {
		log.InfoLog("%d: %s, %d, %s", i+1, v.Uname, v.Mid, strings.ReplaceAll(v.Sign, "\n", " "))
	}
}

func TestCancelFollow(t *testing.T) {
	err := CancelFollow(context.Background(), "614417008")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCancelAll(t *testing.T) {
	all, err := GetAllFollowings(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i, v := range all {
		log.InfoLog("%d: %s, %d, %s", i+1, v.Uname, v.Mid, strings.ReplaceAll(v.Sign, "\n", " "))
		err = CancelFollow(context.Background(), strconv.Itoa(v.Mid))
		if err != nil {
			t.Fatal(err)
		}
	}
}
