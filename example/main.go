package main

import (
	"fmt"
	"github.com/lincaiyong/uniapi/service/edgetts"
	"github.com/lincaiyong/uniapi/service/fornext"
	"github.com/lincaiyong/uniapi/service/monica"
	"os"
)

func monicaExample() {
	monica.Init(monica.ModelClaude4Sonnet, os.Getenv("MONICA_SESSION_ID"))
	_, err := monica.ChatCompletion("131加412", func(s string) {
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

func main() {
	os.Args = []string{"x", "monica"}
	//os.Args[1] = "fornext"
	os.Args[1] = "edgetts"
	if len(os.Args) < 2 {
		fmt.Println("Usage: main <service>")
		os.Exit(1)
	}
	service := os.Args[1]
	switch service {
	case "monica":
		monicaExample()
	case "fornext":
		fornextExample()
	case "edgetts":
		edgettsExample()
	}
}
