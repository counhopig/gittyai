package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/counhopig/gittyai/agent"
	"github.com/counhopig/gittyai/llm"
	"github.com/counhopig/gittyai/orchestrator"
	"github.com/counhopig/gittyai/task"
)

func main() {
	// Example 1: Programmatic API usage
	// This example creates agents and tasks programmatically

	// Create LLM provider
	llmProvider, err := llm.NewOpenAI(llm.Config{
		APIKey:      getEnv("OPENAI_API_KEY", "your-api-key"),
		Model:       "gpt-4o-mini",
		Temperature: 0.7,
		MaxTokens:   2000,
	})
	if err != nil {
		log.Fatalf("Failed to create LLM provider: %v", err)
	}

	// Create agents
	researcher := agent.New(agent.Config{
		Name:      "researcher",
		Role:      "Research Analyst",
		Goal:      "Gather comprehensive information and provide insights",
		Backstory: "Expert research analyst with 10+ years of experience",
		Verbose:   true,
		LLM:       llmProvider,
	})

	writer := agent.New(agent.Config{
		Name:      "writer",
		Role:      "Content Writer",
		Goal:      "Create engaging and informative content",
		Backstory: "Professional writer skilled in tech content",
		Verbose:   true,
		LLM:       llmProvider,
	})

	// Create tasks
	researchTask := task.New(task.Config{
		Description:    "Research the impact of AI on modern software development",
		ExpectedOutput: "A detailed analysis of AI's impact on software development",
		Agent:          researcher,
	})

	writeTask := task.New(task.Config{
		Description:    "Write a blog post based on the research findings",
		ExpectedOutput: "An engaging blog post suitable for publication",
		Agent:          writer,
	})

	// Create orchestrator and execute
	researchOrch := orchestrator.New(orchestrator.Config{
		Agents:  []*agent.Agent{researcher},
		Tasks:   []*task.Task{researchTask},
		Process: orchestrator.Sequential,
	})

	fmt.Println("=== Running Simple Research Agent ===")
	ctx := context.Background()
	results, err := researchOrch.Kickoff(ctx)
	if err != nil {
		log.Fatalf("Execution failed: %v", err)
	}

	fmt.Println(orchestrator.FormatResults(results))

	// Example 2: Multi-agent workflow
	fmt.Println("\n=== Running Multi-Agent Workflow ===")

	fullOrch := orchestrator.New(orchestrator.Config{
		Agents:  []*agent.Agent{researcher, writer},
		Tasks:   []*task.Task{researchTask, writeTask},
		Process: orchestrator.Sequential,
	})

	results, err = fullOrch.Kickoff(ctx)
	if err != nil {
		log.Fatalf("Multi-agent execution failed: %v", err)
	}

	fmt.Println(orchestrator.FormatResults(results))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
