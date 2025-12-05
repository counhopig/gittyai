package config

import (
	"fmt"

	"github.com/counhopig/gittyai/agent"
	"github.com/counhopig/gittyai/crew"
	"github.com/counhopig/gittyai/llm"
	"github.com/counhopig/gittyai/memory"
	"github.com/counhopig/gittyai/task"
)

// Builder helps construct a crew from a configuration
type Builder struct {
	project *Project
	agents  []*agent.Agent
	tasks   []*task.Task
}

// NewBuilder creates a new configuration builder
func NewBuilder(project *Project) *Builder {
	return &Builder{
		project: project,
		agents:  make([]*agent.Agent, 0),
		tasks:   make([]*task.Task, 0),
	}
}

// BuildLLM creates an LLM provider from configuration
func BuildLLM(cfg LLMConfig) (llm.LLM, error) {
	switch cfg.Provider {
	case "openai":
		return llm.NewOpenAI(llm.Config{
			APIKey:      cfg.APIKey,
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   cfg.MaxTokens,
		})
	case "anthropic":
		if cfg.Model == "" {
			cfg.Model = "claude-3-haiku-20240307" // Set a reasonable default
		}
		return llm.NewAnthropic(llm.Config{
			APIKey:      cfg.APIKey,
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   cfg.MaxTokens,
		})
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.Provider)
	}
}

// BuildAgents creates agents from configuration
func (b *Builder) BuildAgents() error {
	llmProvider, err := BuildLLM(b.project.LLM)
	if err != nil {
		return fmt.Errorf("failed to build LLM: %w", err)
	}

	mem := memory.New()

	for _, agentCfg := range b.project.Agents {
		ag := agent.New(agent.Config{
			Name:      agentCfg.Name,
			Role:      agentCfg.Role,
			Goal:      agentCfg.Goal,
			Backstory: agentCfg.Backstory,
			Verbose:   agentCfg.Verbose,
			MaxIter:   agentCfg.MaxIter,
			MaxRPM:    agentCfg.MaxRPM,
			LLM:       llmProvider, // Each agent uses the global LLM
			Memory:    mem,
		})
		b.agents = append(b.agents, ag)
	}

	return nil
}

// BuildTasks creates tasks from configuration
func (b *Builder) BuildTasks() error {
	if len(b.agents) == 0 {
		return fmt.Errorf("agents must be built before tasks")
	}

	// Create agent map for easy lookup
	agentMap := make(map[string]*agent.Agent)
	for _, ag := range b.agents {
		agentMap[ag.Name] = ag
	}

	for _, taskCfg := range b.project.Tasks {
		ag, exists := agentMap[taskCfg.Agent]
		if !exists {
			return fmt.Errorf("task '%s' references non-existent agent: %s", taskCfg.Description, taskCfg.Agent)
		}

		tsk := task.New(task.Config{
			Description:    taskCfg.Description,
			ExpectedOutput: taskCfg.ExpectedOutput,
			Agent:          ag,
			Context:        taskCfg.Context,
		})

		b.tasks = append(b.tasks, tsk)
	}

	return nil
}

// GetAgents returns all built agents
func (b *Builder) GetAgents() []*agent.Agent {
	return b.agents
}

// GetTasks returns all built tasks
func (b *Builder) GetTasks() []*task.Task {
	return b.tasks
}

// Build constructs the crew from configuration
func (b *Builder) Build() (*crew.Crew, error) {
	if err := b.BuildAgents(); err != nil {
		return nil, err
	}

	if err := b.BuildTasks(); err != nil {
		return nil, err
	}

	var process crew.Process
	switch b.project.Execution.Process {
	case "parallel":
		process = crew.Parallel
	default:
		process = crew.Sequential
	}

	return crew.New(crew.Config{
		Agents:  b.agents,
		Tasks:   b.tasks,
		Process: process,
	}), nil
}

// BuildFromConfig is a convenience function to build a crew directly from a config file
func BuildFromConfig(configPath string) (*crew.Crew, error) {
	project, err := LoadYAML(configPath)
	if err != nil {
		return nil, err
	}

	builder := NewBuilder(project)
	return builder.Build()
}
