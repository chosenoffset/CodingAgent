package parser

import (
	"CodingCompanion/vector"
	"context"
	"log"
	"path/filepath"

	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

type GoParser struct {
	db vector.CodeVectorDB
}

func NewGoParser() *GoParser {
	vecDb, err := vector.NewCodeVectorDB()
	if err != nil {
		log.Fatal(err)
	}

	return &GoParser{
		db: *vecDb,
	}
}

func (gp *GoParser) ParseDirectory(dirPath string) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		snippets, err := gp.Parse(path)
		if err != nil {
			return err
		}

		for _, snippet := range snippets {
			gp.db.AddCode(
				snippet.ID,
				snippet.Code,
				snippet.Language,
				snippet.FilePath,
			)
		}

		return nil
	})

	return err
}

func (gp *GoParser) Parse(filepath string) ([]*vector.CodeSnippet, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var snippets []*vector.CodeSnippet

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		startPos := fset.Position(funcDecl.Pos())
		endPos := fset.Position(funcDecl.End())

		code := string(content[startPos.Offset:endPos.Offset])

		var docString string
		if funcDecl.Doc != nil {
			docString = funcDecl.Doc.Text()
		}

		snippet := &vector.CodeSnippet{
			ID:           fmt.Sprintf("%s:%d", filepath, startPos.Line),
			FilePath:     filepath,
			Language:     "go",
			FunctionName: funcDecl.Name.Name,
			Code:         code,
			StartLine:    startPos.Line,
			EndLine:      endPos.Line,
			DocString:    strings.TrimSpace(docString),
		}

		snippets = append(snippets, snippet)
		return true
	})

	return snippets, nil
}

func (gp *GoParser) Query(question string, matchesToReturn int) ([]vector.SearchResult, error) {
	return gp.db.Search(context.Background(), question, matchesToReturn)
}

func (gp *GoParser) BuildPromptWithContext(question string, results []vector.SearchResult) string {
	// If no relevant code found, just ask the question
	if len(results) == 0 {
		return question
	}

	// Build prompt with your code examples
	var prompt strings.Builder

	prompt.WriteString("You are a coding assistant. The user has asked:\n\n")
	prompt.WriteString(question)
	prompt.WriteString("\n\n")

	prompt.WriteString("Here are relevant examples from the user's own codebase:\n\n")

	for i, result := range results {
		prompt.WriteString(fmt.Sprintf("--- Example %d ---\n", i+1))
		prompt.WriteString(result.Code) // The code snippet
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("Based on how the user has solved similar problems in their own code, ")
	prompt.WriteString("provide a solution that matches their coding style and patterns.\n")

	return prompt.String()
}
