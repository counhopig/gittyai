# Gitty - Go AI Agent Framework

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/doc/devel/release)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/counhopig/gittyai)](https://pkg.go.dev/github.com/counhopig/gittyai)

Gitty is a lightweight, easy-to-use AI Agent framework written in Go. Inspired by CrewAI, it allows you to quickly build multi-agent applications for automating tasks, research, content creation, and more.

## Features

- 🚀 **Simple & Intuitive**: Get started in minutes with minimal setup
- 🤖 **Multi-Agent Orchestration**: Coordinate multiple agents with different roles
- ⚡ **True Concurrency**: Built on Go's goroutines for parallel execution
- 🎛️ **Flexible LLM Support**: Integrate with OpenAI, Anthropic, Azure OpenAI, Ollama, Groq, Deepseek, OpenRouter, Together AI, LM Studio, and any OpenAI-compatible API
- 💾 **Memory System**: Short-term and long-term memory for agents
- 🔧 **Extensible**: Easy to add custom tools and integrations
- 📄 **YAML Configuration**: Define agents and tasks in simple configuration files
- 🌐 **OpenAI-Compatible API**: Support for any provider with OpenAI-compatible API (Azure, Ollama, LM Studio, vLLM, LocalAI, Together AI, OpenRouter, etc.)
- 🔌 **Plugin System**: Easy to add custom LLM providers and tools
- 🛡️ **Structured Error Handling**: Comprehensive error types with context, severity levels, and retry logic

## Installation

```bash
go get github.com/counhopig/gittyai
```

### Prerequisites

- Go 1.21 or higher

## Quick Start

### 1. Basic Programmatic API

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/counhopig/gittyai/agent"
    "github.com/counhopig/gittyai/llm"
    "github.com/counhopig/gittyai/orchestrator"
    "github.com/counhopig/gittyai/task"
)

func main() {
    // Create LLM provider (configure as needed)
    llm, err := llm.NewOpenAI(llm.Config{
        APIKey: "your-openai-api-key",  // Configure your API key
        Model:  "gpt-4o-mini",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create an agent
    researcher := agent.New(agent.Config{
        Name:      "researcher",
        Role:      "Research Analyst",
        Goal:      "Gather comprehensive information",
        Backstory: "Expert in finding relevant information",
        LLM:       llm,
    })

    // Create a task
    researchTask := task.New(task.Config{
        Description:    "Research AI Agent frameworks",
        ExpectedOutput: "A detailed summary",
        Agent:          researcher,
    })

    // Create and run the orchestrator
    orch := orchestrator.New(orchestrator.Config{
        Agents:  []*agent.Agent{researcher},
        Tasks:   []*task.Task{researchTask},
        Process: orchestrator.Sequential,
    })

    results, err := orch.Kickoff(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(orchestrator.FormatResults(results))
}
```

### 2. YAML Configuration

Create a `config.yaml` file:

```yaml
project: research-project
version: 1.0

agents:
  - name: researcher
    role: Research Analyst
    goal: Gather comprehensive information from various sources
    backstory: You are an expert research analyst with extensive experience
    verbose: true

tasks:
  - description: Research the latest developments in AI
    expected_output: A comprehensive summary
    agent: researcher

execution:
  process: sequential

llm:
  provider: openai-like  # Supports: openai, anthropic, azure-openai, ollama, groq, deepseek, openrouter, together, lmstudio, openai-like
  api_key: "your-api-key-here"  # Configure as needed
  base_url: "https://api.openai.com/v1"  # Required for openai-like providers
  model: gpt-4o-mini
  temperature: 0.7
  max_tokens: 2000
  # Optional: Custom headers for OpenAI-compatible APIs
  headers:
    "X-Custom-Header": "custom-value"
  # Optional: System prompt for OpenAI-like providers
  system_prompt: "You are a helpful AI assistant"
```

Then use it in your code:

```go
package main

import (
    "context"
    "log"

    "github.com/counhopig/gittyai/config"
)

func main() {
    orch, err := config.BuildFromConfig("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    results, err := orch.Kickoff(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    // process results...
}
```

## Core Concepts

### Agent

An Agent is the core unit that performs tasks. It has:

- **Identity**: Name, role, goal, and backstory
- **Behavior**: Verbose mode, iteration limits, rate limits
- **Memory**: Storage for past interactions
- **LLM**: The language model provider for decision making

```go
agent := agent.New(agent.Config{
    Name:      "writer",
    Role:      "Content Writer",
    Goal:      "Create engaging blog posts",
    Backstory: "Experienced tech writer",
    LLM:       llmProvider,
})
```

### Task

A Task is a unit of work assigned to an agent:

```go
task := task.New(task.Config{
    Description:    "Write a blog post about AI",
    ExpectedOutput: "Well-structured blog post",
    Agent:          writer,
})
```

### Orchestrator

An Orchestrator coordinates multiple agents to execute tasks:

```go
orch := orchestrator.New(orchestrator.Config{
    Agents:  []*agent.Agent{researcher, writer},
    Tasks:   []*task.Task{research, write},
    Process: orchestrator.Sequential, // or Parallel, Hierarchical
})

results, err := orch.Kickoff(ctx)
```

### Process Types

- **Sequential**: Tasks executed one after another
- **Parallel**: Tasks executed concurrently using goroutines
- **Hierarchical**: Tasks delegated by a manager agent (future enhancement)

## Advanced Usage

### Multi-Agent Workflow

```go
// Create different agents
researcher := agent.New(agent.Config{ /* ... */ })
writer := agent.New(agent.Config{ /* ... */ })
reviewer := agent.New(agent.Config{ /* ... */ })

// Create workflow
research := task.New(task.Config{
    Description: "Research trend X",
    Agent:       researcher,
})

write := task.New(task.Config{
    Description: "Write report based on research",
    Agent:       writer,
})

review := task.New(task.Config{
    Description: "Review and improve the report",
    Agent:       reviewer,
})

// Execute workflow
orch := orchestrator.New(orchestrator.Config{
    Agents:  []*agent.Agent{researcher, writer, reviewer},
    Tasks:   []*task.Task{research, write, review},
    Process: orchestrator.Sequential,
})

results, _ := orch.Kickoff(ctx)
```

### Using Different LLM Providers

```go
// OpenAI
openAI := llm.NewOpenAI(llm.Config{
    APIKey: "your-openai-api-key",  // Configure as needed
    Model:  "gpt-4o",
})

// Anthropic Claude
anthropic := llm.NewAnthropic(llm.Config{
    APIKey: "your-anthropic-api-key",  // Configure as needed
    Model:  "claude-3-sonnet-20240229",
})

// Azure OpenAI
azureOpenAI := llm.NewAzureOpenAI(llm.AzureOpenAIConfig{
    Endpoint:       "https://your-resource.openai.azure.com",
    APIKey:         "your-azure-api-key",  // Configure as needed
    DeploymentName: "gpt-4o",
    APIVersion:     "2024-02-15-preview",
})

// Ollama (local LLM)
ollama := llm.NewOpenAILike(llm.OpenAILikeConfig{
    BaseURL: "http://localhost:11434/v1",  // Configure your Ollama server
    Model:   "llama3.2",
    // APIKey is optional for local Ollama
})

// Groq (fast inference)
groq := llm.NewOpenAILike(llm.OpenAILikeConfig{
    BaseURL: "https://api.groq.com/openai/v1",
    APIKey:  "your-groq-api-key",  // Configure as needed
    Model:   "llama-3.1-70b-versatile",
})

// Deepseek
deepseek := llm.NewOpenAILike(llm.OpenAILikeConfig{
    BaseURL: "https://api.deepseek.com/v1",
    APIKey:  "your-deepseek-api-key",  // Configure as needed
    Model:   "deepseek-chat",
})

// OpenRouter (access to multiple models)
openrouter := llm.NewOpenRouter("your-openrouter-api-key", "openai/gpt-4o-mini")

// Together AI
together := llm.NewTogether("your-together-api-key", "meta-llama/Llama-3-70b-chat-hf")

// LM Studio (local)
lmstudio := llm.NewLMStudio("local-model", "http://localhost:1234/v1")

// Any OpenAI-compatible API
custom := llm.NewOpenAILike(llm.OpenAILikeConfig{
    BaseURL:      "https://your-custom-api.com/v1",
    APIKey:       "your-custom-api-key",  // Configure as needed
    Model:        "your-model",
    SystemPrompt: "You are a helpful assistant",  // Optional system prompt
    Headers: map[string]string{
        "X-Custom-Header": "value",  // Custom headers if needed
        "HTTP-Referer":    "https://your-app.com",  // Required by some providers
    },
})

// Use with agent
agent := agent.New(agent.Config{
    LLM: ollama,  // or any other provider
    // ...
})
```

### Adding Tools

```go
twitterTool := tools.NewBaseTool(
    "twitter_search",
    "Search Twitter for recent posts",
    map[string]interface{}{
        "query": "search query",
        "count": 10,
    },
)

// Override the execute method
type TwitterTool struct {
    tools.BaseTool
}

func (t *TwitterTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // Implement tool logic
    return "Twitter search results", nil
}

// Register tool
registry := tools.NewRegistry()
registry.Register(&TwitterTool{*twitterTool})
```

### Custom Memory

```go
// Use the base memory
mem := memory.New()

// Or implement your own

myMemory := &MyMemory{}
agent := agent.New(agent.Config{
    Memory: myMemory,
    // ...
})
```

### Error Handling

GittyAI provides structured error handling with rich context:

```go
import "github.com/counhopig/gittyai/errors"

// Check error types
if err != nil {
    // Check if it's a structured error
    if gittyErr, ok := err.(*errors.Error); ok {
        // Access error details
        fmt.Printf("Error Code: %s\n", gittyErr.Code)
        fmt.Printf("Severity: %s\n", gittyErr.Severity)
        fmt.Printf("Context: %v\n", gittyErr.Context)
        fmt.Printf("Retryable: %v\n", gittyErr.Retryable)

        // Check specific error types
        if errors.IsRetryable(err) {
            // Retry the operation
        }

        if errors.HasCode(err, errors.ErrInvalidConfig) {
            // Handle configuration error
        }
    }
}

// Common error types include:
// - ErrRequiredField: Missing required configuration
// - ErrInvalidConfig: Invalid configuration values
// - ErrAPICall: Failed API calls (automatically marked as retryable)
// - ErrNetworkUnavail: Network issues (automatically marked as retryable)
// - ErrUnsupported: Unsupported features or providers
// - ErrInternal: Internal system errors
```

## Project Structure

```
gittyai/
├── agent/          # Agent definitions and management
├── orchestrator/   # Multi-agent orchestration
├── task/           # Task management
├── llm/            # LLM provider abstractions (OpenAI, Anthropic, OpenAI-compatible)
├── memory/         # Memory systems
├── tools/          # Tool integrations
├── config/         # Configuration parsing (YAML, builder)
├── errors/         # Structured error handling with rich context
└── examples/       # Example projects
    ├── simple.yaml      # Basic configuration example
    ├── advanced.yaml    # Advanced multi-provider configuration example
    ├── api_example.go   # Programmatic API example
    └── config_example.go # Config-driven example
```

## Configuration Reference

### Agent Configuration

| Field       | Type    | Required | Description                               |
| ----------- | ------- | -------- | ----------------------------------------- |
| `name`      | string  | Yes      | Unique agent identifier                   |
| `role`      | string  | Yes      | Agent's role in the team                  |
| `goal`      | string  | Yes      | What the agent aims to accomplish         |
| `backstory` | string  | Yes      | Agent's persona and background            |
| `verbose`   | boolean | No       | Enable detailed logging (default: false)  |
| `max_iter`  | integer | No       | Maximum iterations (default: 25)          |
| `max_rpm`   | integer | No       | Max requests per minute (default: 10)     |
| `tools`     | array   | No       | List of tool names enabled for this agent |

### Task Configuration

| Field             | Type   | Required | Description                 |
| ----------------- | ------ | -------- | --------------------------- |
| `description`     | string | Yes      | Task description            |
| `expected_output` | string | No       | Expected result format      |
| `agent`           | string | Yes      | Agent name to assign        |
| `context`         | array  | No       | Previous tasks to reference |

### LLM Configuration

| Field            | Type    | Required | Description                                      |
| ---------------- | ------- | -------- | ------------------------------------------------ |
| `provider`       | string  | Yes      | LLM provider (openai, anthropic, azure-openai, ollama, groq, deepseek, openrouter, together, lmstudio, openai-like) |
| `api_key`        | string  | Yes*     | API key for the provider (*optional for local providers) |
| `base_url`       | string  | No       | API endpoint URL (required for openai-like providers) |
| `model`          | string  | No       | Model name (defaults vary by provider)           |
| `temperature`    | float   | No       | Generation randomness (0.0-2.0)                  |
| `max_tokens`     | integer | No       | Maximum response tokens                          |
| `system_prompt`  | string  | No       | System prompt for OpenAI-like providers          |
| `headers`        | object  | No       | Custom HTTP headers (openai-like only)           |
| `endpoint`       | string  | No       | Azure OpenAI endpoint (azure-openai only)        |
| `deployment_name`| string  | No       | Azure OpenAI deployment name (azure-openai only) |
| `api_version`    | string  | No       | Azure OpenAI API version (azure-openai only)     |

### Execution Configuration

| Field     | Type   | Required | Description                                        |
| --------- | ------ | -------- | -------------------------------------------------- |
| `process` | string | No       | Execution mode: sequential, parallel, hierarchical |

## Examples

See the `examples/` directory for complete working examples:

- **simple.yaml**: Basic configuration example with OpenAI
- **advanced.yaml**: Advanced multi-provider configuration showcasing various LLM providers (Ollama, Groq, Deepseek, OpenRouter, Together AI, Azure OpenAI, Anthropic, and more)
- **api_example.go**: Programmatic API usage
- **config_example.go**: Config-driven example

## Development

### Running Tests

```bash
go test ./...
```

### Building Examples

```bash
# Build API example
go build -o examples/api_example examples/api_example.go

# Run with config
go run examples/config_example.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Roadmap

- [x] Core agent/task/orchestrator abstractions
- [x] OpenAI and Anthropic integration
- [x] YAML configuration support
- [x] Sequential and parallel execution
- [x] OpenAI-compatible API support (Ollama, Groq, Deepseek, etc.)
- [ ] Concurrent streaming support
- [ ] Advanced tool system
- [ ] RAG-based long-term memory
- [ ] Vector database integration
- [ ] Web UI dashboard
- [ ] More LLM providers (Google Gemini, Cohere, etc.)
- [ ] Plugin system for custom integrations

## Acknowledgments

- Inspired by [CrewAI](https://github.com/crewAIInc/crewAI)
- Thanks to the open-source Go community for libraries and tools

## Support

For questions and support, please open an issue on the GitHub repository.