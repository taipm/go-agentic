package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	agenticcore "github.com/taipm/go-agentic/core"
	agentictools "github.com/taipm/go-agentic/core/tools"
	"github.com/sashabaranov/go-openai"
)

// CreateQdrantTools creates all Qdrant-related tools for the agents
func CreateQdrantTools(qc *QdrantClient, openaiKey string) []*agenticcore.Tool {
	return []*agenticcore.Tool{
		createGenerateEmbeddingTool(openaiKey),
		createSearchCollectionTool(qc),
		createListCollectionsTool(qc),
		createGetCollectionInfoTool(qc),
	}
}

// generateEmbeddingHandler generates embeddings using OpenAI
func generateEmbeddingHandler(openaiKey string) func(context.Context, map[string]interface{}) (string, error) {
	return func(ctx context.Context, args map[string]interface{}) (string, error) {
		text, ok := args["text"].(string)
		if !ok {
			return "", fmt.Errorf("text parameter required and must be a string")
		}

		if text == "" {
			return "", fmt.Errorf("text parameter cannot be empty")
		}

		// Create OpenAI client
		client := openai.NewClient(openaiKey)

		// Generate embedding using text-embedding-3-large
		resp, err := client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
			Model: openai.LargeEmbedding3,
			Input: []string{text},
		})
		if err != nil {
			return "", fmt.Errorf("failed to generate embedding: %w", err)
		}

		if len(resp.Data) == 0 {
			return "", fmt.Errorf("no embedding returned from OpenAI")
		}

		// Convert embedding to JSON
		embedding := resp.Data[0].Embedding
		jsonBytes, err := json.Marshal(embedding)
		if err != nil {
			return "", fmt.Errorf("failed to marshal embedding: %w", err)
		}

		// Return summary + JSON (agents need JSON but logs should be clean)
		return fmt.Sprintf("✅ Embedding generated (%d dimensions)\n%s", len(embedding), string(jsonBytes)), nil
	}
}

// createGenerateEmbeddingTool creates the embedding generation tool
func createGenerateEmbeddingTool(openaiKey string) *agenticcore.Tool {
	return &agenticcore.Tool{
		Name:        "GenerateEmbedding",
		Description: "Generate a vector embedding from Vietnamese text using OpenAI text-embedding-3-large model",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Vietnamese text to convert into an embedding vector (3072 dimensions)",
				},
			},
			"required": []string{"text"},
		},
		Handler: generateEmbeddingHandler(openaiKey),
	}
}

// searchCollectionHandler performs vector similarity search using Qdrant REST API
func searchCollectionHandler(qc *QdrantClient) func(context.Context, map[string]interface{}) (string, error) {
	return func(ctx context.Context, args map[string]interface{}) (string, error) {
		pe := agentictools.NewParameterExtractor(args).WithTool("SearchCollection")
		collectionName := pe.RequireString("collection_name")
		queryVectorJSON := pe.RequireString("query_vector")
		limit := pe.OptionalInt("limit", 5)

		if err := pe.Errors(); err != nil {
			return "", err
		}

		// Parse query vector (JSON array)
		var queryVector []float32
		err := json.Unmarshal([]byte(queryVectorJSON), &queryVector)
		if err != nil {
			return "", fmt.Errorf("failed to parse query_vector: %w", err)
		}

		// Call Qdrant REST API for real search results
		searchResults, err := qc.SearchCollection(ctx, collectionName, queryVector, limit)
		if err != nil {
			// Log error but don't fail - return informative message
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("⚠️ Search in %s returned no results or encountered an error:\n", collectionName))
			sb.WriteString(fmt.Sprintf("Error: %v\n\n", err))
			sb.WriteString("This may indicate:\n")
			sb.WriteString("- Collection does not exist or is empty\n")
			sb.WriteString("- Connection issue with Qdrant server\n")
			sb.WriteString("- No documents matched the query vector\n")
			return sb.String(), nil
		}

		// Format results
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("🔍 Search Results from %s (Found %d results):\n\n", collectionName, len(searchResults)))

		if len(searchResults) == 0 {
			sb.WriteString("No results found for this query.\n")
		} else {
			for i, result := range searchResults {
				score, _ := result["score"].(float32)
				sb.WriteString(fmt.Sprintf("%d. [Score: %.2f]\n", i+1, score))

				// Format payload fields
				for key, value := range result {
					if key != "id" && key != "score" {
						sb.WriteString(fmt.Sprintf("   %s: %v\n", key, value))
					}
				}
				sb.WriteString(fmt.Sprintf("   ID: %v\n\n", result["id"]))
			}
		}

		return sb.String(), nil
	}
}

// createSearchCollectionTool creates the collection search tool
func createSearchCollectionTool(qc *QdrantClient) *agenticcore.Tool {
	return &agenticcore.Tool{
		Name:        "SearchCollection",
		Description: "Search a Qdrant collection using vector similarity. Use GenerateEmbedding first to get the query vector.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection to search: askat_regulations, askat_helpdesk, or askat_incidents",
				},
				"query_vector": map[string]interface{}{
					"type":        "string",
					"description": "JSON array of floats representing the query embedding vector (3072 dimensions)",
				},
				"limit": map[string]interface{}{
					"type":        "string",
					"description": "Maximum number of results to return (default: 5, max: 100)",
				},
			},
			"required": []string{"collection_name", "query_vector"},
		},
		Handler: searchCollectionHandler(qc),
	}
}

// listCollectionsHandler lists all available collections
func listCollectionsHandler(qc *QdrantClient) func(context.Context, map[string]interface{}) (string, error) {
	return func(ctx context.Context, args map[string]interface{}) (string, error) {
		collections, err := qc.ListCollections(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to list collections: %w", err)
		}

		var sb strings.Builder
		sb.WriteString("📚 Available Collections:\n\n")

		for i, name := range collections {
			// Get collection info for status
			info, _ := qc.GetCollectionInfo(ctx, name)
			var status string
			if info != nil {
				status = fmt.Sprintf(" (%d points)", info["points_count"])
			}
			sb.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, name, status))
		}

		return sb.String(), nil
	}
}

// createListCollectionsTool creates the list collections tool
func createListCollectionsTool(qc *QdrantClient) *agenticcore.Tool {
	return &agenticcore.Tool{
		Name:        "ListCollections",
		Description: "List all available Qdrant collections and their basic information",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []string{},
		},
		Handler: listCollectionsHandler(qc),
	}
}

// getCollectionInfoHandler gets detailed info about a collection
func getCollectionInfoHandler(qc *QdrantClient) func(context.Context, map[string]interface{}) (string, error) {
	return func(ctx context.Context, args map[string]interface{}) (string, error) {
		pe := agentictools.NewParameterExtractor(args).WithTool("GetCollectionInfo")
		collectionName := pe.RequireString("collection_name")

		if err := pe.Errors(); err != nil {
			return "", err
		}

		info, err := qc.GetCollectionInfo(ctx, collectionName)
		if err != nil {
			return "", fmt.Errorf("failed to get collection info: %w", err)
		}

		infoBytes, err := json.MarshalIndent(info, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to format collection info: %w", err)
		}

		return fmt.Sprintf("📊 Collection Info for %s:\n\n%s", collectionName, string(infoBytes)), nil
	}
}

// createGetCollectionInfoTool creates the get collection info tool
func createGetCollectionInfoTool(qc *QdrantClient) *agenticcore.Tool {
	return &agenticcore.Tool{
		Name:        "GetCollectionInfo",
		Description: "Get detailed information about a specific Qdrant collection",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection to get information about",
				},
			},
			"required": []string{"collection_name"},
		},
		Handler: getCollectionInfoHandler(qc),
	}
}
