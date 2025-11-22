package main

import (
	"log"
	"os"
	"time"
)

func main() {
	writer, err := NewGlamourWriter()
	if err != nil {
		log.Fatal(err)
	}

	ollamaClient, err := NewOllamaClient("qwen2.5-coder:7b", 120*time.Second, writer)
	if err != nil {
		log.Fatal(err)
	}

	prompt := os.Args[1]

	err = ollamaClient.GenerateWithContext("You are a code generator. Output ONLY code with no explanations.", prompt)
	if err != nil {
		log.Fatal(err)
	}
}
