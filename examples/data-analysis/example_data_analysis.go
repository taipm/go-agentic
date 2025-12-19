package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/taipm/go-agentic"
)

// createDataAnalysisCrew creates a complete data analysis crew
func createDataAnalysisCrew() *agentic.Crew {
	tools := createDataAnalysisTools()

	researcher := &agentic.Agent{
		ID:          "researcher",
		Name:        "Data Researcher",
		Role:        "Data collection and exploration specialist",
		Backstory:   "You are expert at collecting and exploring datasets. Analyze the available data and identify relevant patterns.",
		Model:       "gpt-4o",
		Tools:       tools,
		Temperature: 0.7,
		IsTerminal:  false,
	}

	analyst := &agentic.Agent{
		ID:          "analyst",
		Name:        "Data Analyst",
		Role:        "Statistical analysis expert",
		Backstory:   "You are expert at statistical analysis and data interpretation. Perform deep analysis and identify correlations.",
		Model:       "gpt-4o",
		Tools:       tools,
		Temperature: 0.7,
		IsTerminal:  false,
	}

	synthesizer := &agentic.Agent{
		ID:          "synthesizer",
		Name:        "Report Synthesizer",
		Role:        "Insights and recommendations creator",
		Backstory:   "You specialize in synthesizing analysis results into actionable insights and clear recommendations.",
		Model:       "gpt-4o",
		Tools:       []*agentic.Tool{},
		Temperature: 0.7,
		IsTerminal:  true,
	}

	crew := &agentic.Crew{
		Agents:      []*agentic.Agent{researcher, analyst, synthesizer},
		MaxRounds:   10,
		MaxHandoffs: 5,
	}

	return crew
}

// createDataAnalysisTools creates all data analysis tools
func createDataAnalysisTools() []*agentic.Tool {
	return []*agentic.Tool{
		{
			Name:        "AnalyzeDataset",
			Description: "Analyze a dataset and return summary statistics",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"dataset_name": map[string]interface{}{
						"type":        "string",
						"description": "Name of the dataset to analyze",
					},
				},
				"required": []string{"dataset_name"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				datasetName, ok := args["dataset_name"].(string)
				if !ok {
					return "", fmt.Errorf("invalid dataset_name parameter")
				}
				return fmt.Sprintf("Dataset Analysis: %s\n- Total Records: 10,250\n- Columns: 15\n- Missing Values: 2.3%%\n- Data Types: 8 numeric, 7 categorical", datasetName), nil
			},
		},
		{
			Name:        "GenerateReport",
			Description: "Generate a detailed analysis report",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"analysis_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of analysis report to generate",
					},
				},
				"required": []string{"analysis_type"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				analysisType, ok := args["analysis_type"].(string)
				if !ok {
					return "", fmt.Errorf("invalid analysis_type parameter")
				}
				return fmt.Sprintf("Report Generated: %s Analysis\n- Executive Summary: Key findings identified\n- Methodology: Statistical analysis performed\n- Confidence Level: 95%%\n- Pages: 25", analysisType), nil
			},
		},
		{
			Name:        "IdentifyTrends",
			Description: "Identify trends in the dataset",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"time_period": map[string]interface{}{
						"type":        "string",
						"description": "Time period to analyze trends for",
					},
				},
				"required": []string{"time_period"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				timePeriod, ok := args["time_period"].(string)
				if !ok {
					return "", fmt.Errorf("invalid time_period parameter")
				}
				return fmt.Sprintf("Trend Analysis: %s\n- Upward Trend: +8.5%%\n- Seasonality: Strong Q4 effect\n- Volatility: Moderate\n- Forecast: Continued growth expected", timePeriod), nil
			},
		},
		{
			Name:        "CompareSets",
			Description: "Compare two datasets or subsets",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"set_a": map[string]interface{}{
						"type":        "string",
						"description": "First dataset name",
					},
					"set_b": map[string]interface{}{
						"type":        "string",
						"description": "Second dataset name",
					},
				},
				"required": []string{"set_a", "set_b"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				setA, _ := args["set_a"].(string)
				setB, _ := args["set_b"].(string)
				return fmt.Sprintf("Comparison: %s vs %s\n- Mean Difference: 15.3%%\n- Variance Difference: 8.7%%\n- Statistical Significance: p < 0.05\n- Practical Significance: Moderate", setA, setB), nil
			},
		},
		{
			Name:        "CalculateCorrelation",
			Description: "Calculate correlation between variables",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"variable_1": map[string]interface{}{
						"type":        "string",
						"description": "First variable name",
					},
					"variable_2": map[string]interface{}{
						"type":        "string",
						"description": "Second variable name",
					},
				},
				"required": []string{"variable_1", "variable_2"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				var1, _ := args["variable_1"].(string)
				var2, _ := args["variable_2"].(string)
				correlation := rand.Float64()*2 - 1
				return fmt.Sprintf("Correlation: %s & %s\n- Pearson r: %.3f\n- p-value: 0.001\n- Relationship: Strong positive\n- Causation: Cannot be determined", var1, var2, correlation), nil
			},
		},
		{
			Name:        "DetectAnomalies",
			Description: "Detect anomalies and outliers in the data",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sensitivity": map[string]interface{}{
						"type":        "string",
						"description": "Detection sensitivity (low, medium, high)",
					},
				},
				"required": []string{"sensitivity"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				sensitivity, ok := args["sensitivity"].(string)
				if !ok {
					return "", fmt.Errorf("invalid sensitivity parameter")
				}
				return fmt.Sprintf("Anomaly Detection: %s sensitivity\n- Anomalies Detected: 43\n- Percentage of Data: 0.42%%\n- Top Anomaly: 5.2 std devs away\n- Recommendation: Investigate isolated cases", sensitivity), nil
			},
		},
		{
			Name:        "ProjectTrend",
			Description: "Project future trends based on analysis",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"projection_months": map[string]interface{}{
						"type":        "string",
						"description": "Number of months to project",
					},
				},
				"required": []string{"projection_months"},
			},
			Handler: func(ctx context.Context, args map[string]interface{}) (string, error) {
				months, ok := args["projection_months"].(string)
				if !ok {
					return "", fmt.Errorf("invalid projection_months parameter")
				}
				return fmt.Sprintf("Trend Projection: Next %s months\n- Expected Growth: 12-15%%\n- Confidence Interval: 95%%\n- Risk Factors: Market volatility\n- Recommendation: Monitor quarterly", months), nil
			},
		},
	}
}
