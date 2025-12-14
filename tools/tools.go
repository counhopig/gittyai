package tools

import (
	"context"
	"encoding/json"

	"github.com/counhopig/gittyai/errors"
)

// Tool defines the interface for an agent's tool
type Tool interface {
	// Name returns the tool's name
	Name() string

	// Description returns a description of what the tool does
	Description() string

	// Execute runs the tool with the given arguments
	Execute(ctx context.Context, args map[string]interface{}) (string, error)

	// Args returns the expected argument structure
	Args() map[string]interface{}
}

// Registry manages a collection of tools
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) error {
	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return errors.Validationf("tool %s already registered", name)
	}
	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, errors.NotFound("tool", name)
	}
	return tool, nil
}

// Execute runs a tool by name with the given arguments
func (r *Registry) Execute(ctx context.Context, name string, args map[string]interface{}) (string, error) {
	tool, err := r.Get(name)
	if err != nil {
		return "", err
	}

	return tool.Execute(ctx, args)
}

// List returns all registered tool names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// BaseTool is a basic implementation of the Tool interface
type BaseTool struct {
	name        string
	description string
	args        map[string]interface{}
}

// NewBaseTool creates a base tool with template methods
func NewBaseTool(name, description string, args map[string]interface{}) *BaseTool {
	if args == nil {
		args = make(map[string]interface{})
	}
	return &BaseTool{
		name:        name,
		description: description,
		args:        args,
	}
}

func (b *BaseTool) Name() string                 { return b.name }
func (b *BaseTool) Description() string          { return b.description }
func (b *BaseTool) Args() map[string]interface{} { return b.args }

// ToolCall represents a call to a tool
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ParseToolCall parses a tool call from JSON
func ParseToolCall(data string) (*ToolCall, error) {
	var call ToolCall
	if err := json.Unmarshal([]byte(data), &call); err != nil {
		return nil, errors.Wrap(errors.ErrInternal, "failed to parse tool call", err).WithContext("json_length", len(data))
	}
	return &call, nil
}
