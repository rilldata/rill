package rillv1

import (
	"context"

	"github.com/joho/godotenv"
)

// parseDotEnv parses the .env file at the given path and merges it with the existing env vars
func (p *Parser) parseDotEnv(ctx context.Context, path string) error {
	data, err := p.Repo.Get(ctx, path)
	if err != nil {
		return err
	}
	envMap, err := godotenv.Unmarshal(data)
	if err != nil {
		return err
	}

	if p.DotEnv == nil {
		p.DotEnv = envMap
		return nil
	}

	p.DotEnv = mergeDotEnvMaps(p.DotEnv, envMap)

	return nil
}

// parseDotEnvs parses and merges .env files at the given paths
func (p *Parser) parseDotEnvs(ctx context.Context, paths []string) error {
	for _, path := range paths {
		err := p.parseDotEnv(ctx, path)
		if err != nil {
			return err
		}
	}
	return nil
}

// mergeDotEnvMaps merges two env maps
func mergeDotEnvMaps(a, b map[string]string) map[string]string {
	merged := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		merged[k] = v
	}
	for k, v := range b {
		merged[k] = v
	}
	return merged
}
