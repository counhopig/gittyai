package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYAML loads and parses a YAML configuration file
func LoadYAML(path string) (*Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	project := &Project{}
	if err := yaml.Unmarshal(data, project); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate the parsed project
	if err := project.Validate(); err != nil {
		return nil, fmt.Errorf("invalid project configuration: %w", err)
	}

	return project, nil
}

// SaveYAML saves the project configuration to a YAML file
func SaveYAML(project *Project, path string) error {
	if err := project.Validate(); err != nil {
		return fmt.Errorf("invalid project configuration: %w", err)
	}

	data, err := yaml.Marshal(project)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// Validate checks if the project configuration is valid
func (p *Project) Validate() error {
	if p.Project == "" {
		return fmt.Errorf("project name is required")
	}

	if len(p.Agents) == 0 {
		return fmt.Errorf("at least one agent is required")
	}

	if len(p.Tasks) == 0 {
		return fmt.Errorf("at least one task is required")
	}

	// Check that LLM config is set
	if p.LLM.Provider == "" {
		return fmt.Errorf("LLM provider is required")
	}

	// Validate agents
	agentNames := make(map[string]bool)
	for _, agent := range p.Agents {
		if agent.Name == "" {
			return fmt.Errorf("agent name is required")
		}
		if agentNames[agent.Name] {
			return fmt.Errorf("duplicate agent name: %s", agent.Name)
		}
		agentNames[agent.Name] = true
	}

	// Validate tasks
	for _, task := range p.Tasks {
		if task.Description == "" {
			return fmt.Errorf("task description is required")
		}
		if task.Agent == "" {
			return fmt.Errorf("task agent is required")
		}
		if !agentNames[task.Agent] {
			return fmt.Errorf("task references non-existent agent: %s", task.Agent)
		}
	}

	return nil
}
