# Data Analysis Example

A complete example of using go-agentic for data analysis and insights generation.

## Overview

This example demonstrates a multi-agent system for data analysis:
- **Research Analyst**: Gathers and organizes relevant data
- **Data Analyst**: Performs statistical analysis and identifies patterns
- **Insights Synthesizer**: Generates actionable recommendations

## Features

- Dataset analysis and summarization
- Report generation with visualizations
- Trend identification
- Dataset comparison
- Correlation calculation
- Anomaly detection
- Trend projection

## Quick Start

### 1. Setup

```bash
cd examples/data-analysis
cp ../.env.example .env
# Edit .env and add your OPENAI_API_KEY
```

### 2. Run

```bash
go run main.go
```

### 3. Try It

Example queries:
- "Analyze Q4 sales data"
- "Compare performance between regions"
- "Identify trends in customer retention"
- "Detect anomalies in our metrics"
- "Project next quarter's growth"
- "Find correlations between variables"

## Available Tools

- **AnalyzeDataset** - Summary statistics and distributions
- **GenerateReport** - Comprehensive analysis reports
- **IdentifyTrends** - Time-series trend analysis
- **CompareSets** - Dataset comparison
- **CalculateCorrelation** - Variable correlation analysis
- **DetectAnomalies** - Outlier and anomaly detection
- **ProjectTrend** - Future value projections

## Architecture

```
Data Analysis Crew
├── Researcher (data gathering)
├── Analyst (statistical analysis with tools)
└── Synthesizer (insights and recommendations)
    ├── Dataset Analysis
    ├── Trend Analysis
    ├── Correlation Analysis
    ├── Anomaly Detection
    └── Forecasting
```

## Workflow Example

1. **Researcher** gathers information about dataset
2. **Analyst** performs statistical analysis using tools
3. **Synthesizer** creates actionable insights and recommendations

## Customization

### Adding Custom Analysis Tools

```go
func customAnalysisTool(ctx context.Context, args map[string]interface{}) (string, error) {
    // Your analysis logic
    return "analysis result", nil
}
```

### Modifying Report Format

Edit the report generation templates in the tool handlers.

### Adding Domain-Specific Analysis

Create specialized agents for your industry:
- Financial analysis
- Healthcare metrics
- Marketing analytics
- Operations research

## Integration

Connect to real data sources:

```go
// Example: Connect to database
func loadDataFromDB(query string) ([]string, error) {
    // Load data from your database
}

// Then use in analysis tool
```

## Performance Considerations

- For large datasets, consider batching analysis
- Use caching for frequently analyzed data
- Run complex analyses asynchronously

## Learn More

- See parent directory README for go-agentic fundamentals
- Review tool implementations for analysis patterns
- Check agent roles for customization ideas
