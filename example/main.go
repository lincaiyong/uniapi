package main

import (
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/uniapi/service/baidupan"
	"github.com/lincaiyong/uniapi/service/edgetts"
	"github.com/lincaiyong/uniapi/service/fornext"
	"github.com/lincaiyong/uniapi/service/monica"
	"github.com/lincaiyong/uniapi/service/youtube"
	"os"
)

func monicaExample() {
	monica.Init(os.Getenv("MONICA_SESSION_ID"))
	_, err := monica.ChatCompletion(monica.ModelGPT4oMini, "131加412，春眠不觉晓，", func(s string) {
		fmt.Print(s)
	})
	if err != nil {
		fmt.Printf("fail to completion: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()
}

func fornextExample() {
	fornext.Init(os.Getenv("FORNEXT_SPACE_ID"), os.Getenv("FORNEXT_MODEL_NAME"), os.Getenv("FORNEXT_MODEL_ID"),
		os.Getenv("FORNEXT_PROMPT_KEY"), os.Getenv("FORNEXT_PROMPT_PLATFORM_SESSION"))
	_, err := fornext.ChatCompletion("3+4=", func(s string) {
		fmt.Print(s)
	})
	if err != nil {
		fmt.Printf("fail to completion: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()
}

func edgettsExample() {
	b, err := edgetts.EdgeTTS("你好，春眠不觉晓")
	if err != nil {
		fmt.Printf("fail to run edgetts: %v\n", err)
		os.Exit(1)
	}
	err = os.WriteFile("output.wav", b, 0644)
	if err != nil {
		fmt.Printf("fail to write edgetts: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("saved to output.wav")
}

func youtubeExample() {
	caption, err := youtube.DownloadAutoCaption("vStJoetOxJg")
	if err != nil {
		fmt.Printf("fail to download caption: %v\n", err)
		os.Exit(1)
	}
	b, _ := json.MarshalIndent(caption, "", "  ")
	fmt.Println(string(b))
}

func baidupanExample() {
	baidupan.Init(os.Getenv("BAIDU_PAN_BDUSS"), os.Getenv("BAIDU_PAN_STOKEN"))
	b, err := baidupan.Download("/goodfun/test.txt")
	if err != nil {
		fmt.Printf("fail to download baidupan: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(b))
}

func main() {
	if len(os.Args) < 2 {
		os.Args = []string{"x", "monica"}
		//os.Args[1] = "fornext"
		//os.Args[1] = "edgetts"
		//os.Args[1] = "youtube"
		os.Args[1] = "baidupan"
	}
	service := os.Args[1]
	switch service {
	case "monica":
		monicaExample()
	case "fornext":
		fornextExample()
	case "edgetts":
		edgettsExample()
	case "youtube":
		youtubeExample()
	case "baidupan":
		baidupanExample()
	}
}
