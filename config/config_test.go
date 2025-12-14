package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/counhopig/gittyai/errors"
)

func TestDefaultProject(t *testing.T) {
	project := DefaultProject()

	if project.Project != "my-agent-project" {
		t.Errorf("DefaultProject().Project = %v, want %v", project.Project, "my-agent-project")
	}

	if project.Version != "1.0" {
		t.Errorf("DefaultProject().Version = %v, want %v", project.Version, "1.0")
	}

	if project.LLM.Provider != ProviderOpenAI {
		t.Errorf("DefaultProject().LLM.Provider = %v, want %v", project.LLM.Provider, ProviderOpenAI)
	}

	if project.LLM.Model != "gpt-4o" {
		t.Errorf("DefaultProject().LLM.Model = %v, want %v", project.LLM.Model, "gpt-4o")
	}

	if project.LLM.Temperature != 0.7 {
		t.Errorf("DefaultProject().LLM.Temperature = %v, want %v", project.LLM.Temperature, 0.7)
	}
}

func TestProject_Validate(t *testing.T) {
	tests := []struct {
		name    string
		project *Project
		wantErr bool
		errCode string
	}{
		{
			name: "valid project",
			project: &Project{
				Project: "test-project",
				Agents: []AgentConfig{
					{Name: "agent1", Role: "role1", Goal: "goal1"},
				},
				Tasks: []TaskConfig{
					{Description: "task1", Agent: "agent1"},
				},
				LLM: LLMConfig{
					Provider: ProviderOpenAI,
					Model:    "gpt-4o",
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			project: &Project{
				Project: "",
				Agents: []AgentConfig{
					{Name: "agent1", Role: "role1", Goal: "goal1"},
				},
				Tasks: []TaskConfig{
					{Description: "task1", Agent: "agent1"},
				},
				LLM: LLMConfig{
					Provider: ProviderOpenAI,
					Model:    "gpt-4o",
				},
			},
			wantErr: true,
			errCode: errors.CategoryValidation,
		},
		{
			name: "no agents",
			project: &Project{
				Project: "test-project",
				Agents:  []AgentConfig{},
				Tasks: []TaskConfig{
					{Description: "task1", Agent: "agent1"},
				},
				LLM: LLMConfig{
					Provider: ProviderOpenAI,
					Model:    "gpt-4o",
				},
			},
			wantErr: true,
			errCode: errors.CategoryValidation,
		},
		{
			name: "no tasks",
			project: &Project{
				Project: "test-project",
				Agents: []AgentConfig{
					{Name: "agent1", Role: "role1", Goal: "goal1"},
				},
				Tasks: []TaskConfig{},
				LLM: LLMConfig{
					Provider: ProviderOpenAI,
					Model:    "gpt-4o",
				},
			},
			wantErr: true,
			errCode: errors.CategoryValidation,
		},
		{
			name: "missing LLM provider",
			project: &Project{
				Project: "test-project",
				Agents: []AgentConfig{
					{Name: "agent1", Role: "role1", Goal: "goal1"},
				},
				Tasks: []TaskConfig{
					{Description: "task1", Agent: "agent1"},
				},
				LLM: LLMConfig{
					Provider: "",
					Model:    "gpt-4o",
				},
			},
			wantErr: true,
			errCode: errors.CategoryValidation,
		},
		{
			name: "duplicate agent names",
			project: &Project{
				Project: "test-project",
				Agents: []AgentConfig{
					{Name: "agent1", Role: "role1", Goal: "goal1"},
					{Name: "agent1", Role: "role2", Goal: "goal2"},
				},
				Tasks: []TaskConfig{
					{Description: "task1", Agent: "agent1"},
				},
				LLM: LLMConfig{
					Provider: ProviderOpenAI,
					Model:    "gpt-4o",
				},
			},
			wantErr: true,
			errCode: errors.CategoryValidation,
		},
		{
			name: "task references non-existent agent",
			project: &Project{
				Project: "test-project",
				Agents: []AgentConfig{
					{Name: "agent1", Role: "role1", Goal: "goal1"},
				},
				Tasks: []TaskConfig{
					{Description: "task1", Agent: "agent2"},
				},
				LLM: LLMConfig{
					Provider: ProviderOpenAI,
					Model:    "gpt-4o",
				},
			},
			wantErr: true,
			errCode: errors.CategoryValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
				} else if errObj, ok := err.(*errors.Error); ok {
					if errObj.Code.Category != tt.errCode {
						t.Errorf("Validate() error code = %v, want %v", errObj.Code.Category, tt.errCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestLoadYAML(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `
project: test-project
version: "1.0"
agents:
  - name: researcher
    role: Research Analyst
    goal: Gather information
tasks:
  - description: Research AI trends
    agent: researcher
llm:
  provider: openai
  model: gpt-4o
  temperature: 0.7
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	project, err := LoadYAML(tmpFile)
	if err != nil {
		t.Fatalf("LoadYAML() error = %v", err)
	}

	if project.Project != "test-project" {
		t.Errorf("LoadYAML().Project = %v, want %v", project.Project, "test-project")
	}

	if len(project.Agents) != 1 {
		t.Errorf("LoadYAML() agents count = %v, want %v", len(project.Agents), 1)
	}

	if len(project.Tasks) != 1 {
		t.Errorf("LoadYAML() tasks count = %v, want %v", len(project.Tasks), 1)
	}

	if project.LLM.Provider != ProviderOpenAI {
		t.Errorf("LoadYAML().LLM.Provider = %v, want %v", project.LLM.Provider, ProviderOpenAI)
	}
}

func TestLoadYAML_InvalidFile(t *testing.T) {
	_, err := LoadYAML("/nonexistent/file.yaml")
	if err == nil {
		t.Errorf("LoadYAML() expected error for non-existent file")
	}
}

func TestSaveYAML(t *testing.T) {
	project := &Project{
		Project: "save-test",
		Version: "1.0",
		Agents: []AgentConfig{
			{Name: "test-agent", Role: "Tester", Goal: "Test things"},
		},
		Tasks: []TaskConfig{
			{Description: "Test task", Agent: "test-agent"},
		},
		LLM: LLMConfig{
			Provider:    ProviderOpenAI,
			Model:       "gpt-4o",
			Temperature: 0.7,
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "output.yaml")

	err := SaveYAML(project, tmpFile)
	if err != nil {
		t.Fatalf("SaveYAML() error = %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Errorf("SaveYAML() file was not created")
	}

	// Load it back to verify
	loadedProject, err := LoadYAML(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load saved YAML: %v", err)
	}

	if loadedProject.Project != project.Project {
		t.Errorf("Saved and loaded project name mismatch")
	}
}

func TestProviderConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"OpenAI", ProviderOpenAI, "openai"},
		{"Anthropic", ProviderAnthropic, "anthropic"},
		{"Ollama", ProviderOllama, "ollama"},
		{"LMStudio", ProviderLMStudio, "lmstudio"},
		{"AzureOpenAI", ProviderAzureOpenAI, "azure-openai"},
		{"Groq", ProviderGroq, "groq"},
		{"Together", ProviderTogether, "together"},
		{"Deepseek", ProviderDeepseek, "deepseek"},
		{"Openrouter", ProviderOpenrouter, "openrouter"},
		{"OpenAILike", ProviderOpenAILike, "openai-like"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}
