# Gitty - Go AI Agent Framework

[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/doc/devel/release)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![GoDoc](https://pkg.go.dev/badge/github.com/counhopig/gittyai)](https://pkg.go.dev/github.com/counhopig/gittyai)

Gitty is a lightweight, easy-to-use AI Agent framework written in Go. Inspired by CrewAI, it allows you to quickly build multi-agent applications for automating tasks, research, content creation, and more.

## Features

- 🚀 **Simple & Intuitive**: Get started in minutes with minimal setup
- 🤖 **Multi-Agent Orchestration**: Coordinate multiple agents with different roles
- ⚡ **True Concurrency**: Built on Go's goroutines for parallel execution
- 🎛️ **Flexible LLM Support**: Integrate with OpenAI, Anthropic, and more
- 💾 **Memory System**: Short-term and long-term memory for agents
- 🔧 **Extensible**: Easy to add custom tools and integrations
- 📄 **YAML Configuration**: Define agents and tasks in simple configuration files

## Installation

```bash
go get github.com/counhopig/gittyai
```

### Prerequisites

- Go 1.21 or higher
- API key for OpenAI or Anthropic (or other supported LLM provider)

## Quick Start

### 1. Basic Programmatic API

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/counhopig/gittyai/agent"
    "github.com/counhopig/gittyai/crew"
    "github.com/counhopig/gittyai/llm"
    "github.com/counhopig/gittyai/task"
)

func main() {
    // Create LLM provider
    llm, err := llm.NewOpenAI(llm.Config{
        APIKey: "your-api-key",
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

    // Create and run the crew
    crew := crew.New(crew.Config{
        Agents:  []*agent.Agent{researcher},
        Tasks:   []*task.Task{researchTask},
        Process: crew.Sequential,
    })

    results, err := crew.Kickoff(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(crew.FormatResults(results))
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
  provider: openai
  api_key: ${OPENAI_API_KEY}
  model: gpt-4o-mini
  temperature: 0.7
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
    crew, err := config.BuildFromConfig("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    results, err := crew.Kickoff(context.Background())
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

### Crew

A Crew orchestrates multiple agents to execute tasks:

```go
crew := crew.New(crew.Config{
    Agents:  []*agent.Agent{researcher, writer},
    Tasks:   []*task.Task{research, write},
    Process: crew.Sequential, // or Parallel, Hierarchical
})

results, err := crew.Kickoff(ctx)
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
crew := crew.New(crew.Config{
    Agents:  []*agent.Agent{researcher, writer, reviewer},
    Tasks:   []*task.Task{research, write, review},
    Process: crew.Sequential,
})

results, _ := crew.Kickoff(ctx)
```

### Using Different LLM Providers

```go
// OpenAI
openAI := llm.NewOpenAI(llm.Config{
    APIKey: "your-openai-key",
    Model:  "gpt-4o",
})

// Anthropic Claude
anthropic := llm.NewAnthropic(llm.Config{
    APIKey: "your-anthropic-key",
    Model:  "claude-3-sonnet-20240229",
})

// Use with agent
agent := agent.New(agent.Config{
    LLM: anthropic,
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

## Project Structure

```
gittyai/
├── agent/          # Agent definitions
├── crew/           # Multi-agent orchestration
├── task/           # Task management
├── llm/            # LLM provider abstractions
├── memory/         # Memory systems
├── tools/          # Tool integrations
├── config/         # Configuration parsing
└── examples/       # Example projects
    ├── simple.yaml      # Basic configuration example
    ├── api_example.go   # Programmatic API example
    └── config_example.go # Config-driven example
```

## Configuration Reference

### Agent Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique agent identifier |
| `role` | string | Yes | Agent's role in the team |
| `goal` | string | Yes | What the agent aims to accomplish |
| `backstory` | string | Yes | Agent's persona and background |
| `verbose` | boolean | No | Enable detailed logging (default: false) |
| `max_iter` | integer | No | Maximum iterations (default: 25) |
| `max_rpm` | integer | No | Max requests per minute (default: 10) |
| `tools` | array | No | List of tool names enabled for this agent |

### Task Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | Yes | Task description |
| `expected_output` | string | No | Expected result format |
| `agent` | string | Yes | Agent name to assign |
| `context` | array | No | Previous tasks to reference |

### LLM Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `provider` | string | Yes | LLM provider (openai, anthropic) |
| `api_key` | string | Yes | API key for the provider |
| `model` | string | No | Model name (defaults vary by provider) |
| `temperature` | float | No | Generation randomness (0.0-2.0) |
| `max_tokens` | integer | No | Maximum response tokens |

### Execution Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `process` | string | No | Execution mode: sequential, parallel, hierarchical |

## Environment Variables

- `OPENAI_API_KEY`: Your OpenAI API key
- `ANTHROPIC_API_KEY`: Your Anthropic API key

## Examples

See the `examples/` directory for complete working examples:

- **api_example.go**: Programmatic API usage
- **config_example.go**: YAML configuration usage
- **simple.yaml**: Example configuration file

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

- [x] Core agent/task/crew abstractions
- [x] OpenAI and Anthropic integration
- [x] YAML configuration support
- [x] Sequential and parallel execution
- [ ] Concurrent streaming support
- [ ] Advanced tool system
- [ ] RAG-based long-term memory
- [ ] Vector database integration
- [ ] Web UI dashboard
- [ ] CLI tools
- [ ] More LLM providers

## Acknowledgments

- Inspired by [CrewAI](https://github.com/crewAIInc/crewAI)
- Built with [go-openai](https://github.com/sashabaranov/go-openai)

## Support

For questions and support, please open an issue on the GitHub repository.

---

Made with ❤️ and Go
