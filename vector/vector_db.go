package vector

import (
	"context"
	"log"

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
