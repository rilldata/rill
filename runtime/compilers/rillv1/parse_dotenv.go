package rillv1

import (
	"context"

	"github.com/joho/godotenv"
)

// parseDotEnv parses the env file present at repo root
func (p *Parser) parseDotEnv(ctx context.Context) error {
	data, err := p.Repo.Get(ctx, ".env")
	if err != nil {
		return err
	}
	envMap, err := godotenv.Unmarshal(data)
	if err != nil {
		return err
	}

	p.EnvVariables = envMap
	return nil
}
