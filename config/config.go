package config

// Project represents the complete configuration for a project
type Project struct {
	Project   string            `yaml:"project"`
	Version   string            `yaml:"version"`
	Agents    []AgentConfig     `yaml:"agents"`
	Tasks     []TaskConfig      `yaml:"tasks"`
	Execution ExecutionConfig   `yaml:"execution"`
	LLM       LLMConfig         `yaml:"llm"`
	Settings  map[string]interface{} `yaml:"settings,omitempty"`
}

// AgentConfig represents an agent configuration
type AgentConfig struct {
	Name      string   `yaml:"name"`
	Role      string   `yaml:"role"`
	Goal      string   `yaml:"goal"`
	Backstory string   `yaml:"backstory"`
	Verbose   bool     `yaml:"verbose,omitempty"`
	MaxIter   int      `yaml:"max_iter,omitempty"`
	MaxRPM    int      `yaml:"max_rpm,omitempty"`
	Tools     []string `yaml:"tools,omitempty"`
}

// TaskConfig represents a task configuration
type TaskConfig struct {
	Description    string   `yaml:"description"`
	ExpectedOutput string   `yaml:"expected_output,omitempty"`
	Agent          string   `yaml:"agent"`
	Context        []string `yaml:"context,omitempty"`
}

// ExecutionConfig controls how tasks are executed
type ExecutionConfig struct {
	Process string `yaml:"process"` // "sequential", "parallel", "hierarchical"
}

// LLMConfig holds the LLM provider configuration
type LLMConfig struct {
	Provider    string                 `yaml:"provider"`
	APIKey      string                 `yaml:"api_key"`
	Model       string                 `yaml:"model"`
	Temperature float32                `yaml:"temperature,omitempty"`
	MaxTokens   int                    `yaml:"max_tokens,omitempty"`
	Extra       map[string]interface{} `yaml:",inline"`
}

// DefaultProject returns a minimal default project
func DefaultProject() *Project {
	return &Project{
		Project: "my-agent-project",
		Version: "1.0",
		Execution: ExecutionConfig{
			Process: "sequential",
		},
		LLM: LLMConfig{
			Provider:    "openai",
			Model:       "gpt-4o",
			Temperature: 0.7,
		},
	}
}