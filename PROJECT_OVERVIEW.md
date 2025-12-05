# Gitty Framework - Project Overview

## 🎯 Project Structure

```
gittyai/
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── LICENSE                     # MIT License
├── .gitignore                  # Git ignore patterns
│
├── README.md                   # Main project documentation
├── PROJECT_OVERVIEW.md         # This file
│
├── agent/                      # Agent module
│   └── agent.go               # Core agent implementation
│
├── crew/                       # Crew/workflow orchestration
│   └── crew.go                # Multi-agent coordination
│
├── task/                       # Task module
│   └── task.go                # Task definitions
│
├── llm/                        # LLM abstractions
│   ├── llm.go                 # LLM interface
│   ├── openai.go              # OpenAI implementation
│   └── anthropic.go           # Anthropic implementation
│
├── memory/                     # Memory systems
│   └── memory.go              # Short-term memory
│
├── tools/                      # Tool framework
│   └── tools.go               # Tool interface and registry
│
├── config/                     # Configuration system
│   ├── config.go              # Configuration structures
│   ├── yaml.go                # YAML parser
│   └── builder.go             # Builder pattern for instantiating objects
│
├── examples/                   # Example applications
│   ├── api_example.go         # Programmatic API usage example
│   ├── config_example.go      # Configuration-driven usage
│   └── simple.yaml            # Sample YAML configuration
│
└── doc/                       # Documentation
    ├── getting-started.md     # Quick start guide
    ├── api/                   # API reference (directory)
    ├── examples/              # Extended examples (directory)
    └── guides/                # User guides (directory)

Total: ~13 Go files, 1 config file, full documentation
```

## 🚀 How Users Import and Use This Framework

### 1. Install the Package

```bash
go get github.com/counhopig/gittyai
```

### 2. Basic Usage Pattern

```go
import (
    "github.com/counhopig/gittyai/agent"
    "github.com/counhopig/gittyai/crew"
    "github.com/counhopig/gittyai/llm"
    "github.com/counhopig/gittyai/task"
)

// Create LLM
llm, _ := llm.NewOpenAI(llm.Config{
    APIKey: "your-key",
    Model:  "gpt-4o-mini",
})

// Create Agent
researcher := agent.New(agent.Config{
    Name: "researcher",
    Role: "Research Analyst",
    Goal: "Find information",
    LLM:  llm,
})

// Create Task
t := task.New(task.Config{
    Description: "Research topic X",
    Agent:       researcher,
})

// Create Crew and Execute
c := crew.New(crew.Config{
    Agents:  []*agent.Agent{researcher},
    Tasks:   []*task.Task{t},
    Process: crew.Sequential,
})

results, _ := c.Kickoff(context.Background())
```

### 3. Configuration-Driven Usage

```go
import "github.com/counhopig/gittyai/config"

// Build from YAML
crew, _ := config.BuildFromConfig("config.yaml")
results, _ := crew.Kickoff(context.Background())
```

## 📦 Framework Components

### 1. Core Modules

**Agent** (`agent/`) - Autonomous entity that performs tasks
- Configuration-driven creation
- LLM integration for reasoning
- Memory support (short-term)
- Flexible behavior settings

**Task** (`task/`) - Unit of work to be completed
- Assigned to specific agents
- Can reference previous tasks for context
- Expected output specification

**Crew** (`crew/`) - Multi-agent orchestration
- Sequential, Parallel, and Hierarchical execution
- Synchronized task management
- Result aggregation

### 2. LLM Integration

**Supported Providers**:
- ✅ OpenAI (GPT-4, GPT-3.5, etc.)
- ✅ Anthropic (Claude 3)

**Interface**: Clean abstraction for easy provider addition

```go
type LLM interface {
    Generate(ctx context.Context, prompt string) (string, error)
}
```

### 3. Memory System

**Current**: In-memory short-term storage
**Planned**: Vector-based RAG for long-term memory

### 4. Tools Framework

**Interface**: Extensible tool system
```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
    Args() map[string]interface{}
}
```

### 5. Configuration System

**File Support**: YAML configuration
**Builder Pattern**: Automatic object instantiation from config

## 🎨 Framework Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      USER APPLICATION                       │
└─────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    FRAMEWORK API LAYER                      │
├─────────────────────────────────────────────────────────────┤
│  ✓ Programmatic API    ✓ YAML Config     ✓ Builder Pattern│
└─────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    CORE ORCHESTRATION                       │
├─────────────────────────────────────────────────────────────┤
│  Crew (Sequential/Parallel/Hierarchical)                    │
│  Task Management & Execution                                │
│  Result Aggregation                                         │
└─────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    AGENT EXECUTION                          │
├─────────────────────────────────────────────────────────────┤
│  Agent → LLM.Generate()                                     │
│  Agent → Memory.Store/Retrieve                              │
│  Agent → Tool.Execute() (future)                           │
└─────────────────────────────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE                           │
├─────────────────────────────────────────────────────────────┤
│  LLM Providers (OpenAI, Anthropic)                        │
│  Memory Systems (Short-term, Long-term)                    │
│  Tool Registry & Management                               │
└─────────────────────────────────────────────────────────────┘
```

## 🌟 Key Features

### 1. Simplicity
- Minimal setup required
- Sensible defaults
- Clean, idiomatic Go

### 2. Flexibility
- Multiple LLM providers
- Customizable agents
- Extensible tool system

### 3. Performance
- True concurrency via goroutines
- Minimal overhead
- Efficient resource usage

### 4. Developer Experience
- Type-safe APIs
- Clear error messages
- Comprehensive examples

## 📝 Configuration Example

```yaml
project: my-ai-agent
version: 1.0

agents:
  - name: researcher
    role: Research Analyst
    goal: Find accurate information
    backstory: Expert researcher with 10+ years experience
    verbose: true
    max_iter: 15

tasks:
  - description: Research latest AI trends
    expected_output: Comprehensive report
    agent: researcher

execution:
  process: sequential

llm:
  provider: openai
  api_key: ${OPENAI_API_KEY}
  model: gpt-4o-mini
  temperature: 0.7
```

## 🔧 Extending the Framework

### Adding a New LLM Provider

```go
package llm

type MyProvider struct {
    config Config
}

func NewMyProvider(cfg Config) (*MyProvider, error) {
    return &MyProvider{config: cfg}, nil
}

func (m *MyProvider) Generate(ctx context.Context, prompt string) (string, error) {
    // Implementation
    return "response", nil
}
```

### Adding a Custom Tool

```go
type MyTool struct {
    tools.BaseTool
}

func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
    // Implementation
    return "result", nil
}
```

## ✅ Framework Status

### Implemented
- ✅ Core agent/task/crew abstractions
- ✅ OpenAI and Anthropic integration
- ✅ Sequential and parallel execution
- ✅ YAML configuration support
- ✅ In-memory storage
- ✅ Basic tool framework

### TODO
- ⏳ Stream processing
- ⏳ Vector-based long-term memory
- ⏳ RAG system
- ⏳ More built-in tools
- ⏳ Web UI
- ⏳ Advanced error handling
- ⏳ Observability (metrics, tracing)

## 🎯 Perfect For

- ✨ Startups building AI agents quickly
- 🚀 Prototypes and MVPs
- 📊 Automated research workflows
- 📝 Content generation pipelines
- 🔍 Multi-step analysis tasks
- 🤖 Customer service automation

## 📚 Getting Started

See [README.md](../README.md) for full documentation and [doc/getting-started.md](doc/getting-started.md) for a quick start guide.

## 🎉 Summary

This framework provides:
1. **Clear Abstractions**: Easy-to-understand components
2. **Go Best Practices**: Idiomatic, performant code
3. **Extensibility**: Easy to add providers and tools
4. **Production-Ready**: Error handling, configuration, examples
5. **Developer-Friendly**: Good docs, examples, and API design

Users can import the framework and build sophisticated AI agent systems with minimal code!
