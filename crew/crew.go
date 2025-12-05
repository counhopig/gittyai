package crew

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/counhopig/gittyai/agent"
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
)

// Crew represents a group of agents working together
type Crew struct {
	agents  []*agent.Agent
	tasks   []*task.Task
	process Process
}

// Config represents the configuration for creating a Crew
type Config struct {
	Agents  []*agent.Agent
	Tasks   []*task.Task
	Process Process
}

// New creates a new Crew
func New(cfg Config) *Crew {
	process := cfg.Process
	if process == processUnset {
		process = Sequential
	}

	return &Crew{
		agents:  cfg.Agents,
		tasks:   cfg.Tasks,
		process: process,
	}
}

// Kickoff starts the execution of all tasks
func (c *Crew) Kickoff(ctx context.Context) ([]*TaskResult, error) {
	switch c.process {
	case Sequential:
		return c.executeSequential(ctx)
	case Parallel:
		return c.executeParallel(ctx)
	default:
		return nil, fmt.Errorf("unknown process type: %v", c.process)
	}
}

// executeSequential runs tasks one by one
func (c *Crew) executeSequential(ctx context.Context) ([]*TaskResult, error) {
	results := make([]*TaskResult, 0, len(c.tasks))

	for i, t := range c.tasks {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		fmt.Printf("\n[Task %d/%d] Starting: %s\n", i+1, len(c.tasks), t.Description)

		result, err := c.executeTask(ctx, t)
		if err != nil {
			return results, fmt.Errorf("task %d failed: %w", i, err)
		}

		results = append(results, result)
		fmt.Printf("[Task %d/%d] Completed\n", i+1, len(c.tasks))
	}

	return results, nil
}

// executeParallel runs tasks concurrently
func (c *Crew) executeParallel(ctx context.Context) ([]*TaskResult, error) {
	results := make([]*TaskResult, len(c.tasks))
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	fmt.Printf("\n[Parallel Execution] Starting %d tasks\n", len(c.tasks))

	for i, t := range c.tasks {
		select {
		case <-ctx.Done():
			return results[:i], ctx.Err()
		default:
		}

		wg.Add(1)
		go func(idx int, t *task.Task) {
			defer wg.Done()

			result, taskErr := c.executeTask(ctx, t)
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

// executeTask executes a single task
func (c *Crew) executeTask(ctx context.Context, t *task.Task) (*TaskResult, error) {
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
