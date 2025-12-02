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
			if err := gp.db.AddSnippet(snippet); err != nil {
				return err
			}
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

	var prompt strings.Builder

	prompt.WriteString("REFERENCE EXAMPLES FROM MY CODEBASE:\n")
	prompt.WriteString("(Use these to understand my existing projects and coding style.\n")
	prompt.WriteString("When applicable, base your answer on these files/functions, but improve\n")
	prompt.WriteString("structure and follow current best practices.)\n\n")

	for i, result := range results {
		filePath, funcName, startLine, endLine, codeBody := extractContextFromResult(result.Code)

		prompt.WriteString(fmt.Sprintf("--- Example %d ---\n", i+1))
		if filePath != "" {
			prompt.WriteString(fmt.Sprintf("File: %s\n", filePath))
		}
		if funcName != "" {
			prompt.WriteString(fmt.Sprintf("Function: %s\n", funcName))
		}
		if startLine > 0 && endLine > 0 {
			prompt.WriteString(fmt.Sprintf("Lines: %d-%d\n", startLine, endLine))
		}
		if filePath != "" || funcName != "" || (startLine > 0 && endLine > 0) {
			prompt.WriteString("\n")
		}

		// The actual code (without the synthetic header comments)
		prompt.WriteString(codeBody)
		prompt.WriteString("\n\n")
	}

	prompt.WriteString("QUESTION:\n")
	prompt.WriteString(question)
	prompt.WriteString("\n\n")

	prompt.WriteString("When answering:\n")
	prompt.WriteString("- Identify which of the above files/functions in my codebase you are basing the answer on.\n")
	prompt.WriteString("- Use my existing code as a starting point, but improve structure, readability, and best practices.\n")
	prompt.WriteString("- Keep naming and error-handling patterns consistent with the examples.\n")
	prompt.WriteString("- Provide a complete, runnable code example in your answer.\n")

	return prompt.String()
}

// extractContextFromResult parses the synthetic header we stored in the vector DB
// and returns file path, function name, line numbers, and the code body without headers.
func extractContextFromResult(code string) (filePath, funcName string, startLine, endLine int, body string) {
	lines := strings.Split(code, "\n")

	var headerLines int
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "// file:") {
			filePath = strings.TrimSpace(strings.TrimPrefix(trim, "// file:"))
			headerLines = i + 1
		} else if strings.HasPrefix(trim, "// function:") {
			funcName = strings.TrimSpace(strings.TrimPrefix(trim, "// function:"))
			headerLines = i + 1
		} else if strings.HasPrefix(trim, "// lines:") {
			rangeStr := strings.TrimSpace(strings.TrimPrefix(trim, "// lines:"))
			parts := strings.Split(rangeStr, "-")
			if len(parts) == 2 {
				fmt.Sscanf(parts[0], "%d", &startLine)
				fmt.Sscanf(parts[1], "%d", &endLine)
			}
			headerLines = i + 1
		} else if strings.HasPrefix(trim, "/* doc:") {
			// Treat doc as part of the header; skip it from the body.
			headerLines = i + 1
		} else {
			// Stop once we hit a non-header line.
			break
		}
	}

	if headerLines > 0 && headerLines < len(lines) {
		body = strings.Join(lines[headerLines:], "\n")
	} else {
		body = code
	}

	return
}
