module github.com/taipm/go-agentic/examples/00-hello-crew-tools

go 1.25.2

require (
	github.com/joho/godotenv v1.5.1
	github.com/taipm/go-agentic/core v1.0.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/openai/openai-go/v3 v3.14.0 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	golang.org/x/sync v0.19.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/taipm/go-agentic/core => ../../core
