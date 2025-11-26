package main

import (
	"CodingCompanion/ai"
	"CodingCompanion/formatter"
	parser2 "CodingCompanion/parser"
	"fmt"
	"log"
	"time"
)

func main() {
	goParser := parser2.NewGoParser()
	err := goParser.ParseDirectory("/mnt/c/Development/Projects/Outpost9")
	if err != nil {
		log.Fatal(err)
	}

	writer, err := formatter.NewGlamourWriter()
	if err != nil {
		log.Fatal(err)
	}

	ollamaClient, err := ai.NewOllamaClient("qwen2.5-coder:7b", 120*time.Second, writer)
	if err != nil {
		log.Fatal(err)
	}

	results, err := goParser.Query("How do I setup a scene in Ebitengine?", 3)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n=== Search Results ===\n")
	fmt.Printf("Found %d code snippets:\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. Distance: %.3f\n", i+1, result.Distance)
		fmt.Printf("   Preview: %s...\n", result.Code[:min(100, len(result.Code))])
	}
	fmt.Printf("======================\n\n")

	prompt := goParser.BuildPromptWithContext("How do I setup a scene in Ebitengine?", results)

	fmt.Printf("Prompt: %s\n", prompt)

	err = ollamaClient.GenerateWithContext("You are a code generator. Output ONLY code with no explanations.", prompt)
	if err != nil {
		log.Fatal(err)
	}
}
