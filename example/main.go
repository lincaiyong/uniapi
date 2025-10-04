package main

import (
	"fmt"
	"github.com/lincaiyong/uniapi/service/monica"
	"os"
)

func monicaExample() {
	monica.SetSessionId(os.Getenv("MONICA_SESSION_ID"))
	_, err := monica.ChatCompletion(monica.ModelGPT41Nano, "hi", func(s string) {
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
	if len(os.Args) < 2 {
		fmt.Println("Usage: main <service>")
		os.Exit(1)
	}
	service := os.Args[1]
	switch service {
	case "monica":
		monicaExample()
	}
}
