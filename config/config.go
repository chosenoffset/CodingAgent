package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Index struct {
		Directories []string `yaml:"directories"`
		Extensions  []string `yaml:"extensions"`
		Excludes    []string `yaml:"exclude"`
	} `yaml:"index"`

	LLM struct {
		Model        string  `yaml:"model"`
		Temperature  float64 `yaml:"temperature"`
		Timeout      string  `yaml:"timeout"`
		SystemPrompt string  `yaml:"system_prompt"`
	} `yaml:"llm"`

	VectorDB struct {
		URL        string `yaml:"url"`
		Collection string `yaml:"collection"`
	} `yaml:"vector_db"`

	Ollama struct {
		URL            string `yaml:"url"`
		EmbeddingModel string `yaml:"embedding_model"`
	} `yaml:"ollama"`
}

func LoadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
