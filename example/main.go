package main

import (
	"fmt"
	"github.com/lincaiyong/uniapi/service/fornext"
	"github.com/lincaiyong/uniapi/service/monica"
	"os"
)

func monicaExample() {
	monica.Init(monica.ModelClaude4Sonnet, os.Getenv("MONICA_SESSION_ID"))
	_, err := monica.ChatCompletion("hi", func(s string) {
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
	_, err := fornext.ChatCompletion("hi", func(s string) {
		fmt.Print(s)
	})
	if err != nil {
		fmt.Printf("fail to completion: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()
}

func main() {
	os.Args = []string{"x", "monica"}
	os.Args[1] = "fornext"
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
	}
}
