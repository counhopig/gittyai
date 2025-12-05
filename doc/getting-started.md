# Getting Started with Gitty

This guide will help you get up and running with Gitty in 5 minutes.

## Installation

### Prerequisites

- Go 1.21 or higher
- API key from OpenAI or Anthropic

### Install the Package

```bash
go get github.com/counhopig/gittyai
```

## Your First Agent

Let's create a simple research agent that can answer questions.

### 1. Create a New Project

```bash
mkdir my-gitty-project
cd my-gitty-project
go mod init my-agent
```

### 2. Create the Main File

Create `main.go`:

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
        APIKey: "your-api-key-here", // Replace with your API key
        Model:  "gpt-4o-mini",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create an agent
    agent := agent.New(agent.Config{
        Name: "researcher",
        Role: "Research Analyst",
        Goal: "Find and analyze information",
        Backstory: "Expert at finding relevant information quickly.",
        LLM: llm,
    })

    // Create a task
    t := task.New(task.Config{
        Description: "What are the benefits of using Go for AI agents?",
        Agent: agent,
    })

    // Run the agent
    crew := crew.New(crew.Config{
        Agents: []*agent.Agent{agent},
        Tasks:  []*task.Task{t},
        Process: crew.Sequential,
    })

    results, err := crew.Kickoff(context.Background())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(crew.FormatResults(results))
}
```

### 3. Run Your Agent

```bash
export OPENAI_API_KEY="your-api-key-here"
go run main.go
```

You should see your agent's response!

## Next Steps

- Learn about [Multi-Agent Teams](multi-agent-teams.md)
- Explore [Task Orchestration](task-orchestration.md)
- Read the [API Reference](../api/)
- Check out [Examples](./examples//)
