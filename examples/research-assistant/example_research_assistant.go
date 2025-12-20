package main

import (
	"context"
	"fmt"

	"github.com/taipm/go-agentic"
)

// createResearchAssistantCrew creates a complete research assistant crew
func createResearchAssistantCrew() *agentic.Crew {
	tools := createResearchAssistantTools()

	investigator := &agentic.Agent{
		ID:          "investigator",
		Name:        "Research Investigator",
		Role:        "Literature search and discovery specialist",
		Backstory:   "You are expert at finding and retrieving relevant academic papers and research materials. Use databases to discover relevant work.",
		Model:       "gpt-4o-mini",
		Tools:       tools,
		Temperature: 0.7,
		IsTerminal:  false,
	}

	analyzer := &agentic.Agent{
		ID:          "analyzer",
		Name:        "Research Analyzer",
		Role:        "Academic analysis expert",
		Backstory:   "You are expert at analyzing research papers, identifying methodologies, and extracting key insights.",
		Model:       "gpt-4o-mini",
		Tools:       tools,
		Temperature: 0.7,
		IsTerminal:  false,
	}

	documentor := &agentic.Agent{
		ID:          "documentor",
		Name:        "Research Documentor",
		Role:        "Research synthesis and documentation expert",
		Backstory:   "You specialize in synthesizing research findings into coherent research summaries and recommendations.",
		Model:       "gpt-4o-mini",
		Tools:       []*agentic.Tool{},
		Temperature: 0.7,
		IsTerminal:  true,
	}

	crew := &agentic.Crew{
		Agents:      []*agentic.Agent{investigator, analyzer, documentor},
		MaxRounds:   10,
		MaxHandoffs: 5,
	}

	return crew
}

// createResearchAssistantTools creates all research assistant tools
func createResearchAssistantTools() []*agentic.Tool {
	return []*agentic.Tool{
		{
			Name:        "SearchAcademicDatabases",
			Description: "Search academic databases like PubMed, arXiv, Google Scholar",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for academic papers",
					},
					"database": map[string]interface{}{
						"type":        "string",
						"description": "Database to search (pubmed, arxiv, scholar)",
					},
				},
				"required": []string{"query"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				query, ok := args["query"].(string)
				if !ok {
					return "", fmt.Errorf("invalid query parameter")
				}
				database, _ := args["database"].(string)
				if database == "" {
					database = "google-scholar"
				}
				return fmt.Sprintf("Search Results from %s:\n- Result 1: Paper on %s (2023)\n- Result 2: Study comparing approaches (2022)\n- Result 3: Recent meta-analysis (2024)\n- Total Results: 1,247", database, query), nil
			},
		},
		{
			Name:        "ExtractPaperMetadata",
			Description: "Extract metadata and abstract from an academic paper",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"paper_id": map[string]interface{}{
						"type":        "string",
						"description": "Paper identifier (DOI, arXiv ID, or PubMed ID)",
					},
				},
				"required": []string{"paper_id"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				paperID, ok := args["paper_id"].(string)
				if !ok {
					return "", fmt.Errorf("invalid paper_id parameter")
				}
				return fmt.Sprintf("Paper Metadata: %s\n- Title: Advanced Methods in Research\n- Authors: Smith, J., Johnson, A., et al.\n- Published: 2023\n- Citations: 42\n- Abstract: This paper presents novel methodologies...", paperID), nil
			},
		},
		{
			Name:        "AnalyzeCitationNetwork",
			Description: "Analyze citation networks and paper relationships",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"paper_id": map[string]interface{}{
						"type":        "string",
						"description": "Central paper ID",
					},
				},
				"required": []string{"paper_id"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				paperID, ok := args["paper_id"].(string)
				if !ok {
					return "", fmt.Errorf("invalid paper_id parameter")
				}
				return fmt.Sprintf("Citation Network Analysis for %s:\n- Papers Citing This: 127\n- Papers Cited: 84\n- Influential Papers: 8\n- Citation Trajectory: Rising\n- Related Fields: 5", paperID), nil
			},
		},
		{
			Name:        "CompareTheories",
			Description: "Compare different theoretical approaches in the literature",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"theory_1": map[string]interface{}{
						"type":        "string",
						"description": "First theory name",
					},
					"theory_2": map[string]interface{}{
						"type":        "string",
						"description": "Second theory name",
					},
				},
				"required": []string{"theory_1", "theory_2"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				theory1, _ := args["theory_1"].(string)
				theory2, _ := args["theory_2"].(string)
				return fmt.Sprintf("Theory Comparison: %s vs %s\n- Similarities: Both address core mechanisms\n- Key Differences: Different causal pathways\n- Supporting Evidence: 12 papers for theory 1, 15 for theory 2\n- Consensus: Partial consensus in field", theory1, theory2), nil
			},
		},
		{
			Name:        "FindKeyAuthors",
			Description: "Identify key authors and their contributions in a field",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"field": map[string]interface{}{
						"type":        "string",
						"description": "Research field or topic",
					},
				},
				"required": []string{"field"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				field, ok := args["field"].(string)
				if !ok {
					return "", fmt.Errorf("invalid field parameter")
				}
				return fmt.Sprintf("Key Authors in %s:\n- Author 1: 156 publications, h-index 48\n- Author 2: 127 publications, h-index 42\n- Author 3: 98 publications, h-index 38\n- Emerging: Author 4 with 23 recent papers", field), nil
			},
		},
		{
			Name:        "IdentifyResearchGaps",
			Description: "Identify gaps and future directions in research",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"topic": map[string]interface{}{
						"type":        "string",
						"description": "Research topic to analyze",
					},
				},
				"required": []string{"topic"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				topic, ok := args["topic"].(string)
				if !ok {
					return "", fmt.Errorf("invalid topic parameter")
				}
				return fmt.Sprintf("Research Gaps in %s:\n- Gap 1: Limited longitudinal studies\n- Gap 2: Few cross-cultural analyses\n- Gap 3: Integration with new methods\n- Opportunity: Machine learning applications\n- Priority: High-impact emerging directions", topic), nil
			},
		},
		{
			Name:        "AnalyzeMethodology",
			Description: "Analyze and compare research methodologies",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"methodology_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of methodology (qualitative, quantitative, mixed)",
					},
				},
				"required": []string{"methodology_type"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				methodType, ok := args["methodology_type"].(string)
				if !ok {
					return "", fmt.Errorf("invalid methodology_type parameter")
				}
				return fmt.Sprintf("Methodology Analysis: %s\n- Papers Using This: 87\n- Common Approaches: 5 major patterns\n- Strengths: High rigor, clear results\n- Limitations: Generalization concerns\n- Evolution: Increasing adoption over time", methodType), nil
			},
		},
		{
			Name:        "SynthesizeFindings",
			Description: "Synthesize findings from multiple papers",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"topic": map[string]interface{}{
						"type":        "string",
						"description": "Topic for synthesis",
					},
					"num_papers": map[string]interface{}{
						"type":        "string",
						"description": "Number of papers to synthesize",
					},
				},
				"required": []string{"topic"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				topic, _ := args["topic"].(string)
				numPapers, _ := args["num_papers"].(string)
				return fmt.Sprintf("Synthesis of Findings for %s (%s papers):\n- Common Themes: 4 major themes identified\n- Level of Agreement: 78%% consensus\n- Key Evidence: 45 robust findings\n- Conflicting Results: 3 areas of disagreement\n- Strength of Evidence: Moderate to Strong", topic, numPapers), nil
			},
		},
		{
			Name:        "TrackFieldDevelopment",
			Description: "Track the evolution and development of a research field",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"field": map[string]interface{}{
						"type":        "string",
						"description": "Research field to track",
					},
					"time_period": map[string]interface{}{
						"type":        "string",
						"description": "Time period to analyze (e.g., '2010-2024')",
					},
				},
				"required": []string{"field"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				field, _ := args["field"].(string)
				timePeriod, _ := args["time_period"].(string)
				return fmt.Sprintf("Field Development: %s (%s)\n- Publication Growth: 12%% annual increase\n- Paradigm Shifts: 2 major transitions\n- Emerging Topics: AI-enabled approaches\n- Leading Institutions: 5 main research centers\n- Future Directions: Interdisciplinary integration", field, timePeriod), nil
			},
		},
	}
}
