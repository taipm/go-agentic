# Research Assistant Example

A complete example of using go-agentic for research and information gathering.

## Overview

This example demonstrates a multi-agent system for research:
- **Research Investigator**: Investigates topics and gathers information
- **Research Analyst**: Analyzes findings and identifies patterns
- **Research Documentor**: Documents findings in clear reports

## Features

- Academic database searching
- Citation network analysis
- Theory comparison
- Key author identification
- Research gap identification
- Methodology analysis
- Finding synthesis
- Field development tracking

## Quick Start

### 1. Setup

```bash
cd examples/research-assistant
cp ../.env.example .env
# Edit .env and add your OPENAI_API_KEY
```

### 2. Run

```bash
go run main.go
```

### 3. Try It

Example research queries:
- "Research machine learning in healthcare"
- "Find key authors in AI safety"
- "What are research gaps in deep learning?"
- "Compare supervised vs unsupervised learning"
- "Track evolution of neural networks"
- "Analyze methodology trends in NLP"
- "Identify anomalies in citation patterns"

## Available Tools

- **SearchAcademicDatabases** - Search papers (PubMed, IEEE, arXiv, etc)
- **ExtractPaperMetadata** - Get paper details and citations
- **AnalyzeCitationNetwork** - Understand citation relationships
- **CompareTheories** - Compare theoretical frameworks
- **FindKeyAuthors** - Identify influential researchers
- **IdentifyResearchGaps** - Find unexplored areas
- **AnalyzeMethodology** - Understand research approaches
- **SynthesizeFindings** - Combine multiple studies
- **TrackFieldDevelopment** - Historical evolution
- **EvaluateSourceCredibility** - Assess source reliability

## Architecture

```
Research Assistant Crew
├── Investigator (search & gather with tools)
├── Analyst (analyze & synthesize with tools)
└── Documentor (report generation)
    ├── Academic Search
    ├── Citation Analysis
    ├── Theory Analysis
    ├── Author Analysis
    ├── Methodology Analysis
    ├── Gap Analysis
    └── Field Tracking
```

## Workflow Example

1. **Investigator** searches academic databases for papers
2. **Analyst** analyzes citations, theories, and methodologies
3. **Documentor** synthesizes findings into comprehensive reports

## Customization

### Adding Research Domains

```go
func customResearchTool(ctx context.Context, args map[string]interface{}) (string, error) {
    // Your research logic
    return "research findings", nil
}
```

### Connecting to Real Databases

Integrate with:
- PubMed API
- IEEE Xplore
- arXiv
- Semantic Scholar
- CrossRef

### Custom Analysis Metrics

Add domain-specific analysis:
- Citation impact analysis
- Researcher collaboration networks
- Research trend prediction
- Funding source analysis

## Integration Examples

### Connect to Real Academic APIs

```go
// Example: Query PubMed
func searchPubMed(query string) ([]Paper, error) {
    // Implementation using PubMed API
}

// Use in research tool
```

### Generate Research Reports

Automatically generate:
- Literature reviews
- Research summaries
- Citation analysis reports
- Methodology comparisons

## Advanced Features

### Citation Network Analysis

Understand how papers reference each other and identify influential works.

### Theory Comparison

Systematically compare different theoretical frameworks in your field.

### Research Gap Analysis

Identify promising areas for future research and unexplored questions.

## Learn More

- See parent directory README for go-agentic fundamentals
- Review tool implementations for research patterns
- Check agent configurations for customization
- Explore academic APIs for data integration

## Use Cases

- Literature review automation
- Research trend analysis
- Methodology benchmarking
- Citation impact analysis
- Researcher networking
- Grant opportunity identification
