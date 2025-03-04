package rillv1

import (
	"context"
	"fmt"
	"maps"
	"os"

	"github.com/joho/godotenv"
)

// parseAndMergeEnvironmentVariables parses the .env file at the given path and merges it with the existing env vars
func (p *Parser) parseAndMergeEnvironmentVariables(ctx context.Context, path string) error {
	// Unlikely, but if the path is empty, return an error
	if path == "" {
		return fmt.Errorf("empty path provided for .env file")
	}

	// Read the .env file
	data, err := p.Repo.Get(ctx, path)
	if err != nil {
		// If the file doesn't exist, return nil to exit early
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Unmarshal the .env file
	envMap, err := godotenv.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal .env file at %s: %w", path, err)
	}

	// Check if the map exists in the current state
	currentEnv, ok := p.DotEnv[path]
	if !ok {
		p.DotEnv[path] = envMap
		return nil
	}

	// Check if the map is the same as the current state
	if maps.Equal(currentEnv, envMap) {
		return nil
	}

	// Overwrite the current state with the new state
	p.DotEnv[path] = envMap

	return nil
}
