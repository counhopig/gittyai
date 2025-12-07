package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/counhopig/gittyai/agent"
	"github.com/counhopig/gittyai/llm"
	"github.com/counhopig/gittyai/task"
)

// Process defines how tasks are executed
type Process int

const (
	// processUnset is the zero value, used internally to detect unset
	processUnset Process = iota
	// Sequential executes tasks one after another
	Sequential
	// Parallel executes tasks concurrently
	Parallel
	// Hierarchical uses a manager LLM to orchestrate task assignments
	Hierarchical
)

// Orchestrator represents a group of agents working together
type Orchestrator struct {
	agents     []*agent.Agent
	tasks      []*task.Task
	process    Process
	managerLLM llm.LLM // Manager LLM for hierarchical orchestration
	goal       string  // High-level goal for hierarchical mode
	verbose    bool
}

// Config represents the configuration for creating an Orchestrator
type Config struct {
	Agents     []*agent.Agent
	Tasks      []*task.Task
	Process    Process
	ManagerLLM llm.LLM // Optional: LLM for intelligent task orchestration
	Goal       string  // Optional: High-level goal for hierarchical mode
	Verbose    bool
}

// New creates a new Orchestrator
func New(cfg Config) *Orchestrator {
	process := cfg.Process
	if process == processUnset {
		process = Sequential
	}

	return &Orchestrator{
		agents:     cfg.Agents,
		tasks:      cfg.Tasks,
		process:    process,
		managerLLM: cfg.ManagerLLM,
		goal:       cfg.Goal,
		verbose:    cfg.Verbose,
	}
}

// Kickoff starts the execution of all tasks
func (o *Orchestrator) Kickoff(ctx context.Context) ([]*TaskResult, error) {
	switch o.process {
	case Sequential:
		return o.executeSequential(ctx)
	case Parallel:
		return o.executeParallel(ctx)
	case Hierarchical:
		return o.executeHierarchical(ctx)
	default:
		return nil, fmt.Errorf("unknown process type: %v", o.process)
	}
}

// executeSequential runs tasks one by one
func (o *Orchestrator) executeSequential(ctx context.Context) ([]*TaskResult, error) {
	results := make([]*TaskResult, 0, len(o.tasks))

	for i, t := range o.tasks {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		fmt.Printf("\n[Task %d/%d] Starting: %s\n", i+1, len(o.tasks), t.Description)

		result, err := o.executeTask(ctx, t)
		if err != nil {
			return results, fmt.Errorf("task %d failed: %w", i, err)
		}

		results = append(results, result)
		fmt.Printf("[Task %d/%d] Completed\n", i+1, len(o.tasks))
	}

	return results, nil
}

// executeParallel runs tasks concurrently
func (o *Orchestrator) executeParallel(ctx context.Context) ([]*TaskResult, error) {
	results := make([]*TaskResult, len(o.tasks))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	fmt.Printf("\n[Parallel Execution] Starting %d tasks\n", len(o.tasks))

	for i, t := range o.tasks {
		select {
		case <-ctx.Done():
			return results[:i], ctx.Err()
		default:
		}

		wg.Add(1)
		go func(idx int, t *task.Task) {
			defer wg.Done()

			result, taskErr := o.executeTask(ctx, t)
			mu.Lock()
			if taskErr != nil {
				errs = append(errs, fmt.Errorf("task %d failed: %w", idx, taskErr))
			} else {
				results[idx] = result
			}
			mu.Unlock()
		}(i, t)
	}

	wg.Wait()

	if len(errs) > 0 {
		return results, errors.Join(errs...)
	}

	return results, nil
}

// executeHierarchical uses a manager LLM to intelligently orchestrate tasks
func (o *Orchestrator) executeHierarchical(ctx context.Context) ([]*TaskResult, error) {
	if o.managerLLM == nil {
		return nil, fmt.Errorf("hierarchical mode requires a manager LLM")
	}

	if len(o.agents) == 0 {
		return nil, fmt.Errorf("no agents available for orchestration")
	}

	fmt.Println("\n[Hierarchical Mode] Manager is planning task execution...")

	// If we have predefined tasks, let manager assign agents
	if len(o.tasks) > 0 {
		return o.orchestratePredefinedTasks(ctx)
	}

	// If we only have a goal, let manager decompose it into tasks
	if o.goal != "" {
		return o.orchestrateFromGoal(ctx)
	}

	return nil, fmt.Errorf("hierarchical mode requires either tasks or a goal")
}

// orchestratePredefinedTasks assigns agents to predefined tasks using manager LLM
func (o *Orchestrator) orchestratePredefinedTasks(ctx context.Context) ([]*TaskResult, error) {
	results := make([]*TaskResult, 0, len(o.tasks))

	// Build agent descriptions for the manager
	agentDescriptions := o.buildAgentDescriptions()

	for i, t := range o.tasks {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		// If task already has an agent assigned, use it
		if t.Agent != nil {
			fmt.Printf("\n[Task %d/%d] Using assigned agent '%s' for: %s\n", i+1, len(o.tasks), t.Agent.Name, t.Description)
			result, err := o.executeTask(ctx, t)
			if err != nil {
				return results, fmt.Errorf("task %d failed: %w", i, err)
			}
			results = append(results, result)
			fmt.Printf("[Task %d/%d] Completed\n", i+1, len(o.tasks))
			continue
		}

		// Ask manager to select the best agent
		selectedAgent, err := o.selectAgentForTask(ctx, t, agentDescriptions)
		if err != nil {
			return results, fmt.Errorf("manager failed to select agent for task %d: %w", i, err)
		}

		fmt.Printf("\n[Task %d/%d] Manager assigned '%s' for: %s\n", i+1, len(o.tasks), selectedAgent.Name, t.Description)

		// Create a new task with the selected agent
		assignedTask := t.WithAgent(selectedAgent)
		result, err := o.executeTask(ctx, assignedTask)
		if err != nil {
			return results, fmt.Errorf("task %d failed: %w", i, err)
		}
		results = append(results, result)
		fmt.Printf("[Task %d/%d] Completed\n", i+1, len(o.tasks))
	}

	return results, nil
}

// orchestrateFromGoal decomposes a high-level goal into tasks and executes them
func (o *Orchestrator) orchestrateFromGoal(ctx context.Context) ([]*TaskResult, error) {
	fmt.Printf("\n[Goal] %s\n", o.goal)
	fmt.Println("[Manager] Decomposing goal into tasks...")

	// Build agent descriptions
	agentDescriptions := o.buildAgentDescriptions()

	// Ask manager to create a plan
	plan, err := o.createExecutionPlan(ctx, agentDescriptions)
	if err != nil {
		return nil, fmt.Errorf("manager failed to create execution plan: %w", err)
	}

	if o.verbose {
		fmt.Printf("[Manager] Created plan with %d tasks\n", len(plan))
	}

	// Execute the plan
	results := make([]*TaskResult, 0, len(plan))
	previousResults := ""

	for i, step := range plan {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		fmt.Printf("\n[Step %d/%d] Agent '%s' executing: %s\n", i+1, len(plan), step.AgentName, step.TaskDescription)

		// Find the agent
		selectedAgent := o.findAgentByName(step.AgentName)
		if selectedAgent == nil {
			// Fallback to first agent if not found
			selectedAgent = o.agents[0]
			fmt.Printf("[Warning] Agent '%s' not found, using '%s' instead\n", step.AgentName, selectedAgent.Name)
		}

		// Create task with context from previous results
		taskDesc := step.TaskDescription
		if previousResults != "" && step.UseContext {
			taskDesc = fmt.Sprintf("%s\n\nContext from previous tasks:\n%s", taskDesc, previousResults)
		}

		newTask := task.New(task.Config{
			Description:    taskDesc,
			ExpectedOutput: step.ExpectedOutput,
			Agent:          selectedAgent,
		})

		result, err := o.executeTask(ctx, newTask)
		if err != nil {
			return results, fmt.Errorf("step %d failed: %w", i+1, err)
		}

		results = append(results, result)
		previousResults += fmt.Sprintf("\n--- %s (by %s) ---\n%s\n", step.TaskDescription, step.AgentName, result.Result)
		fmt.Printf("[Step %d/%d] Completed\n", i+1, len(plan))
	}

	return results, nil
}

// PlanStep represents a single step in the execution plan
type PlanStep struct {
	TaskDescription string `json:"task_description"`
	AgentName       string `json:"agent_name"`
	ExpectedOutput  string `json:"expected_output"`
	UseContext      bool   `json:"use_context"`
}

// buildAgentDescriptions creates a description of all available agents
func (o *Orchestrator) buildAgentDescriptions() string {
	var sb strings.Builder
	sb.WriteString("Available Agents:\n")
	for i, a := range o.agents {
		sb.WriteString(fmt.Sprintf("%d. Name: %s\n", i+1, a.Name))
		sb.WriteString(fmt.Sprintf("   Role: %s\n", a.Role))
		sb.WriteString(fmt.Sprintf("   Goal: %s\n", a.Goal))
		if a.Backstory != "" {
			sb.WriteString(fmt.Sprintf("   Backstory: %s\n", a.Backstory))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// selectAgentForTask asks the manager LLM to select the best agent for a task
func (o *Orchestrator) selectAgentForTask(ctx context.Context, t *task.Task, agentDescriptions string) (*agent.Agent, error) {
	prompt := fmt.Sprintf(`You are a manager responsible for assigning tasks to the best-suited agent.

%s
Task to assign:
Description: %s
Expected Output: %s

Based on the agents' roles and goals, which agent is best suited for this task?
Respond with ONLY the agent's name, nothing else.`, agentDescriptions, t.Description, t.ExpectedOutput)

	response, err := o.managerLLM.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Find the agent by name
	agentName := strings.TrimSpace(response)
	for _, a := range o.agents {
		if strings.EqualFold(a.Name, agentName) || strings.Contains(strings.ToLower(response), strings.ToLower(a.Name)) {
			return a, nil
		}
	}

	// Fallback to first agent if no match
	if o.verbose {
		fmt.Printf("[Manager] Could not match agent '%s', using '%s'\n", agentName, o.agents[0].Name)
	}
	return o.agents[0], nil
}

// createExecutionPlan asks the manager LLM to create an execution plan from a goal
func (o *Orchestrator) createExecutionPlan(ctx context.Context, agentDescriptions string) ([]PlanStep, error) {
	prompt := fmt.Sprintf(`You are a manager responsible for breaking down goals into tasks and assigning them to agents.

%s
Goal to achieve: %s

Create an execution plan to achieve this goal. For each step, specify:
1. The task description
2. Which agent should handle it (use exact agent name)
3. Expected output
4. Whether it needs context from previous tasks (true/false)

Respond in JSON format as an array of steps:
[
  {
    "task_description": "...",
    "agent_name": "...",
    "expected_output": "...",
    "use_context": false
  }
]

Keep the plan focused and efficient. Only include necessary steps.`, agentDescriptions, o.goal)

	response, err := o.managerLLM.Generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Extract JSON from response
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil, fmt.Errorf("manager did not return valid JSON plan")
	}

	var plan []PlanStep
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse execution plan: %w", err)
	}

	if len(plan) == 0 {
		return nil, fmt.Errorf("manager returned empty plan")
	}

	return plan, nil
}

// findAgentByName finds an agent by name (case-insensitive)
func (o *Orchestrator) findAgentByName(name string) *agent.Agent {
	for _, a := range o.agents {
		if strings.EqualFold(a.Name, name) {
			return a
		}
	}
	return nil
}

// extractJSON extracts JSON array from a string that might contain other text
func extractJSON(s string) string {
	start := strings.Index(s, "[")
	if start == -1 {
		return ""
	}

	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

// executeTask executes a single task
func (o *Orchestrator) executeTask(ctx context.Context, t *task.Task) (*TaskResult, error) {
	result, err := t.Execute(ctx)
	if err != nil {
		return nil, err
	}

	return &TaskResult{
		Task:   t,
		Result: result,
		Agent:  t.Agent.Name,
	}, nil
}

// TaskResult holds the result of a task execution
type TaskResult struct {
	Task   *task.Task
	Result string
	Agent  string
}

// String returns a formatted string of all results
func FormatResults(results []*TaskResult) string {
	output := "\n=== EXECUTION RESULTS ===\n\n"
	for i, r := range results {
		output += fmt.Sprintf("Task %d: %s\n", i+1, r.Task.Description)
		output += fmt.Sprintf("Agent: %s\n", r.Agent)
		output += fmt.Sprintf("Result:\n%s\n", r.Result)
		output += "------------------------\n\n"
	}
	return output
}
