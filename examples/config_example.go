package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/counhopig/gittyai/config"
	"github.com/counhopig/gittyai/orchestrator"
)

func configExample() {
	// Example: Using a YAML configuration file

	configPath := "simple.yaml"

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("Config file not found: %s", configPath)
	}

	// Build orchestrator from configuration
	c, err := config.BuildFromConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to build orchestrator from config: %v", err)
	}

	fmt.Println("=== Running Agent from Config File ===")
	ctx := context.Background()
	results, err := c.Kickoff(ctx)
	if err != nil {
		log.Fatalf("Execution failed: %v", err)
	}

	fmt.Println(orchestrator.FormatResults(results))
}
