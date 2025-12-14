package task

import (
	"context"
	"fmt"

	"github.com/counhopig/gittyai/agent"
	"github.com/counhopig/gittyai/errors"
)

// Task represents a unit of work to be completed
type Task struct {
	Description    string
	ExpectedOutput string
	Agent          *agent.Agent
	Context        []string // References to previous tasks for context
}

// Config represents the configuration for creating a Task
type Config struct {
	Description    string
	ExpectedOutput string
	Agent          *agent.Agent
	Context        []string
}

// New creates a new Task
func New(cfg Config) *Task {
	return &Task{
		Description:    cfg.Description,
		ExpectedOutput: cfg.ExpectedOutput,
		Agent:          cfg.Agent,
		Context:        cfg.Context,
	}
}

// WithAgent sets the agent for the task
func (t *Task) WithAgent(a *agent.Agent) *Task {
	newTask := *t
	newTask.Agent = a
	return &newTask
}

// Execute runs the task and returns the result
func (t *Task) Execute(ctx context.Context) (string, error) {
	if t.Agent == nil {
		return "", errors.Validationf("task '%s' has no agent assigned", t.Description)
	}

	// Build prompt from task description and expected output
	prompt := t.Description
	if len(t.ExpectedOutput) > 0 {
		prompt += fmt.Sprintf("\n\nExpected output: %s", t.ExpectedOutput)
	}

	result, err := t.Agent.Execute(ctx, prompt)
	if err != nil {
		return "", errors.Wrap(errors.ErrInternal, "task execution failed", err).WithContext("task_description", t.Description).WithContext("agent", t.Agent.Name)
	}

	return result, nil
}

// String returns a string representation of the task
func (t *Task) String() string {
	agentName := "unassigned"
	if t.Agent != nil {
		agentName = t.Agent.Name
	}
	return fmt.Sprintf("Task{Description: %s, Agent: %s}", t.Description, agentName)
}
