package parser

import "fmt"

// parseChangeModeYAML parses the change mode from the YAML file.
func (p *Parser) parseChangeModeYAML(mode string) (string, error) {
	if mode == "" {
		return "reset", nil
	}

	switch mode {
	case "reset", "manual", "patch":
		return mode, nil
	default:
		return "", fmt.Errorf("unsupported change mode: %q (supported values: reset, manual, patch)", mode)
	}
}
