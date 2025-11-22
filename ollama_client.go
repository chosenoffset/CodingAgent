package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ollama/ollama/api"
)

type OllamaClient struct {
	model      string
	client     *api.Client
	randomness float64
	timeout    time.Duration
	writer     FormatWriter
}

func NewOllamaClient(model string, timeout time.Duration, writer FormatWriter) (*OllamaClient, error) {
	client := &OllamaClient{model: model, timeout: timeout, writer: writer}
	var err error

	client.client, err = api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("unable to create client")
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err = client.client.Show(ctx, &api.ShowRequest{Model: model})
	if err != nil {
		return nil, fmt.Errorf("model %s not found: %w (try: ollama pull %s)", model, err, model)
	}

	return client, nil
}

func (client *OllamaClient) SetRandomness(randomness float64) {
	client.randomness = randomness
}

func (client *OllamaClient) GenerateWithContext(systemPrompt, userPrompt string) error {
	finalPrompt := fmt.Sprintf("System: %s\n\nUser: %s\n\nAssistant:", systemPrompt, userPrompt)
	return client.Generate(finalPrompt)
}

func (client *OllamaClient) Generate(prompt string) error {
	request := &api.GenerateRequest{
		Model:  client.model,
		Prompt: prompt,
		Options: map[string]interface{}{
			"randomness":  client.randomness,
			"num_predict": 2048,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.timeout)
	defer cancel()

	responseFunc := func(response api.GenerateResponse) error {
		return client.writer.Write(response.Response)
	}
	err := client.client.Generate(ctx, request, responseFunc)

	if err != nil {
		return err
	}

	return client.writer.Flush()
}
