package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
)

type FormatWriter interface {
	Write(string) error
	Flush() error
}

type GlamourWriter struct {
	buffer   strings.Builder
	renderer *glamour.TermRenderer
	printed  int
}

func NewGlamourWriter() (*GlamourWriter, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, err
	}

	return &GlamourWriter{
		renderer: r,
		printed:  0,
	}, nil
}

func (sgw *GlamourWriter) Write(text string) error {
	fmt.Print(text)
	sgw.buffer.WriteString(text)
	sgw.printed++

	return nil
}

func (sgw *GlamourWriter) Flush() error {
	content := sgw.buffer.String()

	if content == "" {
		return nil
	}

	if sgw.printed > 0 {
		fmt.Print("\r\033[K") // Clear line

		for i := 0; i < strings.Count(content, "\n"); i++ {
			fmt.Print("\033[F\033[K") // Move up and clear
		}
	}

	out, err := sgw.renderer.Render(content)
	if err != nil {
		return err
	}

	fmt.Print(out)
	sgw.buffer.Reset()
	sgw.printed = 0
	return nil
}
