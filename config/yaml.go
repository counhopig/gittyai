package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/counhopig/gittyai/errors"
)

// LoadYAML loads and parses a YAML configuration file
func LoadYAML(path string) (*Project, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(errors.ErrMissingConfig, fmt.Sprintf("failed to read file %s", path), err).WithContext("path", path)
	}

	project := &Project{}
	if err := yaml.Unmarshal(data, project); err != nil {
		return nil, errors.Wrap(errors.ErrInvalidConfig, "failed to parse YAML", err)
	}

	// Validate the parsed project
	if err := project.Validate(); err != nil {
		return nil, errors.Wrap(errors.ErrInvalidConfig, "invalid project configuration", err)
	}

	return project, nil
}

// SaveYAML saves the project configuration to a YAML file
func SaveYAML(project *Project, path string) error {
	if err := project.Validate(); err != nil {
		return errors.Wrap(errors.ErrInvalidConfig, "invalid project configuration", err)
	}

	data, err := yaml.Marshal(project)
	if err != nil {
		return errors.Wrap(errors.ErrInternal, "failed to marshal YAML", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.Wrap(errors.ErrInternal, fmt.Sprintf("failed to write file %s", path), err).WithContext("path", path)
	}

	return nil
}

// Validate checks if the project configuration is valid
func (p *Project) Validate() error {
	if p.Project == "" {
		return errors.RequiredField("project name")
	}

	if len(p.Agents) == 0 {
		return errors.Validation("at least one agent is required")
	}

	if len(p.Tasks) == 0 {
		return errors.Validation("at least one task is required")
	}

	// Check that LLM config is set
	if p.LLM.Provider == "" {
		return errors.RequiredField("LLM provider")
	}

	// Validate agents
	agentNames := make(map[string]bool)
	for _, agent := range p.Agents {
		if agent.Name == "" {
			return errors.RequiredField("agent name")
		}
		if agentNames[agent.Name] {
			return errors.Validationf("duplicate agent name: %s", agent.Name)
		}
		agentNames[agent.Name] = true
	}

	// Validate tasks
	for _, task := range p.Tasks {
		if task.Description == "" {
			return errors.RequiredField("task description")
		}
		if task.Agent == "" {
			return errors.RequiredField("task agent")
		}
		if !agentNames[task.Agent] {
			return errors.Validationf("task references non-existent agent: %s", task.Agent)
		}
	}

	return nil
}
