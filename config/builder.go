package config

import (
	"github.com/counhopig/gittyai/agent"
	"github.com/counhopig/gittyai/errors"
	"github.com/counhopig/gittyai/llm"
	"github.com/counhopig/gittyai/memory"
	"github.com/counhopig/gittyai/orchestrator"
	"github.com/counhopig/gittyai/task"
)

// Builder helps construct an orchestrator from a configuration
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
	case ProviderOpenAI:
		return llm.NewOpenAI(llm.Config{
			APIKey:      cfg.APIKey,
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   cfg.MaxTokens,
		})
	case ProviderAnthropic:
		if cfg.Model == "" {
			cfg.Model = "claude-3-haiku-20240307" // Set a reasonable default
		}
		return llm.NewAnthropic(llm.Config{
			APIKey:      cfg.APIKey,
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			MaxTokens:   cfg.MaxTokens,
		})
	case ProviderAzureOpenAI:
		return buildAzureOpenAI(cfg)
	case ProviderOllama, ProviderLMStudio, ProviderGroq, ProviderTogether, ProviderDeepseek, ProviderOpenrouter, ProviderOpenAILike:
		// Handle OpenAI-like providers
		baseURL := cfg.BaseURL
		model := cfg.Model

		// Set default base URLs and models for known providers if not specified
		if baseURL == "" {
			switch cfg.Provider {
			case ProviderOllama:
				baseURL = "http://localhost:11434/v1"
				if model == "" {
					model = "llama3.2"
				}
			case ProviderLMStudio:
				baseURL = "http://localhost:1234/v1"
				if model == "" {
					model = "local-model"
				}
			case ProviderGroq:
				baseURL = "https://api.groq.com/openai/v1"
				if model == "" {
					model = "llama-3.1-70b-versatile"
				}
			case ProviderTogether:
				baseURL = "https://api.together.xyz/v1"
				if model == "" {
					model = "meta-llama/Llama-3-70b-chat-hf"
				}
			case ProviderDeepseek:
				baseURL = "https://api.deepseek.com/v1"
				if model == "" {
					model = "deepseek-chat"
				}
			case ProviderOpenrouter:
				baseURL = "https://openrouter.ai/api/v1"
				if model == "" {
					model = "openai/gpt-4o-mini"
				}
			}
		}

		return llm.NewOpenAILike(llm.OpenAILikeConfig{
			BaseURL:      baseURL,
			APIKey:       cfg.APIKey,
			Model:        model,
			Temperature:  cfg.Temperature,
			MaxTokens:    cfg.MaxTokens,
			Headers:      cfg.Headers,
			SystemPrompt: cfg.SystemPrompt,
		})
	default:
		return nil, errors.UnsupportedType(cfg.Provider).WithContext("provider", cfg.Provider)
	}
}

// buildAzureOpenAI creates an Azure OpenAI LLM provider from configuration
func buildAzureOpenAI(cfg LLMConfig) (llm.LLM, error) {
	if cfg.Endpoint == "" {
		return nil, errors.RequiredField("endpoint")
	}
	if cfg.DeploymentName == "" {
		return nil, errors.RequiredField("deployment_name")
	}

	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = "2024-02-15-preview"
	}

	return llm.NewAzureOpenAI(llm.AzureOpenAIConfig{
		Endpoint:       cfg.Endpoint,
		APIKey:         cfg.APIKey,
		DeploymentName: cfg.DeploymentName,
		APIVersion:     apiVersion,
		Temperature:    cfg.Temperature,
		MaxTokens:      cfg.MaxTokens,
	})
}

// BuildAgents creates agents from configuration
func (b *Builder) BuildAgents() error {
	llmProvider, err := BuildLLM(b.project.LLM)
	if err != nil {
		return errors.Wrap(errors.ErrInvalidConfig, "failed to build LLM", err).WithContext("provider", b.project.LLM.Provider)
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
		return errors.InvalidConfig("build_order", "agents must be built before tasks")
	}

	// Create agent map for easy lookup
	agentMap := make(map[string]*agent.Agent)
	for _, ag := range b.agents {
		agentMap[ag.Name] = ag
	}

	for _, taskCfg := range b.project.Tasks {
		ag, exists := agentMap[taskCfg.Agent]
		if !exists {
			return errors.Configf("task '%s' references non-existent agent: %s", taskCfg.Description, taskCfg.Agent)
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

// Build constructs the orchestrator from configuration
func (b *Builder) Build() (*orchestrator.Orchestrator, error) {
	if err := b.BuildAgents(); err != nil {
		return nil, err
	}

	if err := b.BuildTasks(); err != nil {
		return nil, err
	}

	var process orchestrator.Process
	switch b.project.Execution.Process {
	case "parallel":
		process = orchestrator.Parallel
	default:
		process = orchestrator.Sequential
	}

	return orchestrator.New(orchestrator.Config{
		Agents:  b.agents,
		Tasks:   b.tasks,
		Process: process,
	}), nil
}

// BuildFromConfig is a convenience function to build an orchestrator directly from a config file
func BuildFromConfig(configPath string) (*orchestrator.Orchestrator, error) {
	project, err := LoadYAML(configPath)
	if err != nil {
		return nil, err
	}

	builder := NewBuilder(project)
	return builder.Build()
}
