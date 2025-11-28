package main

import (
	"CodingCompanion/ai"
	config2 "CodingCompanion/config"
	"CodingCompanion/formatter"
	parser2 "CodingCompanion/parser"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	config, err := config2.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var buildIndex bool
	var query string

	flag.BoolVar(&buildIndex, "index", false, "Index source directories")
	flag.StringVar(&query, "query", "Explain to the user that if they want to ask a question, they have to provide the text of the question", "Query string")
	flag.Parse()

	if !isOllamaRunning() {
		err = startOllama()
		if err != nil {
			log.Fatal(err)
		}
	}

	if !isChromaRunning() {
		err = startChroma()
		if err != nil {
			log.Fatal(err)
		}
	}

	goParser := parser2.NewGoParser()

	if buildIndex {
		indexLocalProjects(goParser, config.Index.Directories)
	} else {
		writer, err := formatter.NewGlamourWriter()
		if err != nil {
			log.Fatal(err)
		}

		ollamaClient, err := ai.NewOllamaClient(config.LLM.Model, 120*time.Second, writer)
		if err != nil {
			log.Fatal(err)
		}

		results, err := goParser.Query(query, 3)
		if err != nil {
			log.Fatal(err)
		}

		for _, result := range results {
			fmt.Println("Result distance:", result.Distance)
		}

		prompt := goParser.BuildPromptWithContext(query, results)

		err = ollamaClient.GenerateWithContext(config.LLM.SystemPrompt, prompt)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func isOllamaRunning() bool {
	resp, err := http.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == 200
}

func isChromaRunning() bool {
	resp, err := http.Get("http://localhost:8000/api/v2/heartbeat")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == 200
}

func startOllama() error {
	fmt.Println("Starting Ollama...")
	cmd := exec.Command("ollama", "serve")
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Wait up to 10 seconds for Ollama to be ready
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		if isOllamaRunning() {
			fmt.Println("✓ Ollama started")
			return nil
		}
	}

	return fmt.Errorf("ollama started but didn't become ready in time")
}

func startChroma() error {
	fmt.Println("Starting ChromaDB...")
	home, _ := os.UserHomeDir()
	dataPath := filepath.Join(home, ".coding-assistant", "chroma_data")
	_ = os.MkdirAll(dataPath, 0755)

	cmd := exec.Command("chroma", "run", "--path", dataPath, "--port", "8000")
	err := cmd.Start()
	if err != nil {
		return err
	}

	for i := 0; i < 20; i++ {
		time.Sleep(1 * time.Second)
		if isChromaRunning() {
			fmt.Println("✓ ChromaDB started")
			return nil
		}
	}

	return fmt.Errorf("chromadb started but didn't become ready in time")
}

func indexLocalProjects(goParser *parser2.GoParser, directories []string) {
	for _, directory := range directories {
		fmt.Printf("Processing: %s\n", directory)
		err := goParser.ParseDirectory(directory)
		if err != nil {
			fmt.Println(err)
		}
	}
}
