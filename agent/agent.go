package agent

import (
	"context"
	"fmt"

	"github.com/counhopig/gittyai/errors"
	"github.com/counhopig/gittyai/llm"
	"github.com/counhopig/gittyai/memory"
)

// Agent represents an AI agent with specific capabilities and behavior
type Agent struct {
	// Identity
	Name      string
	Role      string
	Goal      string
	Backstory string

	// Behavior
	Verbose bool
	MaxIter int
	MaxRPM  int

	// Memory
	Memory memory.Memory

	// LLM Provider
	LLM llm.LLM
}

// Config represents the configuration for creating an Agent
type Config struct {
	Name      string
	Role      string
	Goal      string
	Backstory string
	Verbose   bool
	MaxIter   int
	MaxRPM    int
	LLM       llm.LLM
	Memory    memory.Memory
}

// New creates a new Agent
func New(cfg Config) *Agent {
	maxIter := cfg.MaxIter
	if maxIter <= 0 {
		maxIter = 25
	}

	maxRPM := cfg.MaxRPM
	if maxRPM <= 0 {
		maxRPM = 10
	}

	return &Agent{
		Name:      cfg.Name,
		Role:      cfg.Role,
		Goal:      cfg.Goal,
		Backstory: cfg.Backstory,
		Verbose:   cfg.Verbose,
		MaxIter:   maxIter,
		MaxRPM:    maxRPM,
		LLM:       cfg.LLM,
		Memory:    cfg.Memory,
	}
}

// Execute processes a task and returns the result
func (a *Agent) Execute(ctx context.Context, taskDescription string) (string, error) {
	if a.LLM == nil {
		return "", errors.MissingConfig("LLM provider").WithContext("agent", a.Name)
	}

	// Build the prompt
	prompt := a.buildPrompt(taskDescription)

	// Call LLM
	resp, err := a.LLM.Generate(ctx, prompt)
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "failed to execute task", err).WithContext("agent", a.Name).WithContext("task_length", len(taskDescription))
	}

	// Store in memory
	if a.Memory != nil {
		_ = a.Memory.Store(ctx, memory.Record{
			AgentName: a.Name,
			Content:   fmt.Sprintf("Task: %s\nResult: %s", taskDescription, resp),
		})
	}

	return resp, nil
}

// buildPrompt constructs the prompt for the agent
func (a *Agent) buildPrompt(task string) string {
	return fmt.Sprintf(
		`You are %s.
Your role is: %s
Your goal is: %s
Your backstory: %s

Task: %s

Please complete the task and provide a clear, detailed response.`,
		a.Name,
		a.Role,
		a.Goal,
		a.Backstory,
		task,
	)
}

// String returns a string representation of the agent
func (a *Agent) String() string {
	return fmt.Sprintf("Agent{Name: %s, Role: %s, Goal: %s}", a.Name, a.Role, a.Goal)
}
