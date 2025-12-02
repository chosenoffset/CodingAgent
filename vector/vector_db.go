package vector

import (
	"context"
	"fmt"
	"log"
	"strings"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
)

type CodeVectorDB struct {
	client     chroma.Client
	collection chroma.Collection
}

type SearchResult struct {
	Code     string
	Distance float64
}

func NewCodeVectorDB() (*CodeVectorDB, error) {
	// Create client
	client, err := chroma.NewHTTPClient()
	if err != nil {
		return nil, err
	}

	// Create collection - ChromaDB handles embeddings automatically
	collection, err := client.GetOrCreateCollection(
		context.Background(),
		"code_snippets",
		chroma.WithCollectionMetadataCreate(
			chroma.NewMetadata(
				chroma.NewStringAttribute("type", "code"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return &CodeVectorDB{
		client:     client,
		collection: collection,
	}, nil
}

// AddSnippet stores a full CodeSnippet with contextual information encoded
// into both the text and the metadata, so it can be reconstructed later.
func (db *CodeVectorDB) AddSnippet(snippet *CodeSnippet) error {
	// Build a header that encodes context directly into the stored text.
	// This makes it easy to reconstruct file, function, and line numbers
	// from the search result without relying on Chroma-specific metadata APIs.
	var headerBuilder strings.Builder
	headerBuilder.WriteString(fmt.Sprintf("// file: %s\n", snippet.FilePath))
	if snippet.FunctionName != "" {
		headerBuilder.WriteString(fmt.Sprintf("// function: %s\n", snippet.FunctionName))
	}
	if snippet.StartLine > 0 || snippet.EndLine > 0 {
		headerBuilder.WriteString(fmt.Sprintf("// lines: %d-%d\n", snippet.StartLine, snippet.EndLine))
	}
	if snippet.DocString != "" {
		// Keep docstring compact; trim extra whitespace.
		headerBuilder.WriteString("/* doc: ")
		headerBuilder.WriteString(strings.TrimSpace(snippet.DocString))
		headerBuilder.WriteString(" */\n")
	}

	text := headerBuilder.String() + snippet.Code

	return db.collection.Add(
		context.Background(),
		chroma.WithIDs(chroma.DocumentID(snippet.ID)),
		chroma.WithTexts(text),
		chroma.WithMetadatas(
			chroma.NewDocumentMetadata(
				chroma.NewStringAttribute("language", snippet.Language),
				chroma.NewStringAttribute("file", snippet.FilePath),
				chroma.NewStringAttribute("function_name", snippet.FunctionName),
				chroma.NewStringAttribute("start_line", fmt.Sprintf("%d", snippet.StartLine)),
				chroma.NewStringAttribute("end_line", fmt.Sprintf("%d", snippet.EndLine)),
			),
		),
	)
}

func (db *CodeVectorDB) AddCode(id string, code string, language string, filepath string) error {
	// ChromaDB automatically generates embeddings for the text
	return db.collection.Add(
		context.Background(),
		chroma.WithIDs(chroma.DocumentID(id)),
		chroma.WithTexts(code), // Just pass text, ChromaDB embeds it
		chroma.WithMetadatas(
			chroma.NewDocumentMetadata(
				chroma.NewStringAttribute("language", language),
				chroma.NewStringAttribute("file", filepath),
			),
		),
	)
}

func (db *CodeVectorDB) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// ChromaDB automatically embeds the query text
	results, err := db.collection.Query(
		ctx,
		chroma.WithQueryTexts(query), // Just pass text, ChromaDB embeds it
		chroma.WithNResults(limit),
	)
	if err != nil {
		log.Fatalf("Error querying: %s\n", err)
	}

	var searchResults []SearchResult
	for _, group := range results.GetDocumentsGroups() {
		for i, doc := range group {
			searchResults = append(searchResults, SearchResult{
				Code:     doc.ContentString(),
				Distance: float64(results.GetDistancesGroups()[0][i]),
			})
		}
	}

	return searchResults, nil
}

func (db *CodeVectorDB) Close() error {
	return db.client.Close()
}
