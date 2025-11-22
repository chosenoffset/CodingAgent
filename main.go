package main

import (
	"fmt"
	"log"
	"time"
)

func main() {
	ollamaClient, err := NewOllamaClient("", "qwen2.5-coder:7b", 120*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	response, err := ollamaClient.GenerateWithContext("You are a code generator. Output ONLY code with no explanations.", "Write an HTTP handler in Go")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response)
}
