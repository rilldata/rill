package cmdutil

import (
	"fmt"
	"slices"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

func SelectPrompt(msg string, options []string, def string) (string, error) {
	prompt := &survey.Select{
		Message: msg,
		Options: options,
	}

	if slices.Contains(options, def) {
		prompt.Default = def
	}

	result := ""
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	return result, nil
}

// SelectPromptWithDescriptions is like SelectPrompt but shows a description next to each option.
// The descriptions slice must be the same length as options.
func SelectPromptWithDescriptions(msg string, options, descriptions []string, def string) (string, error) {
	if len(options) != len(descriptions) {
		return "", fmt.Errorf("options and descriptions must be the same length")
	}

	prompt := &survey.Select{
		Message: msg,
		Options: options,
		Description: func(value string, index int) string {
			return descriptions[index]
		},
	}

	if slices.Contains(options, def) {
		prompt.Default = def
	}

	result := ""
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	return result, nil
}

func ConfirmPrompt(msg, help string, def bool) (bool, error) {
	prompt := &survey.Confirm{
		Message: msg,
		Default: def,
	}

	if help != "" {
		prompt.Help = help
	}

	result := def
	if err := survey.AskOne(prompt, &result); err != nil {
		return false, fmt.Errorf("prompt failed: %w", err)
	}
	return result, nil
}

func InputPrompt(msg, def string) (string, error) {
	prompt := &survey.Input{
		Message: msg,
		Default: def,
	}
	result := def
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	return strings.TrimSpace(result), nil
}

func StringPrompt(msg string) (string, error) {
	prompt := &survey.Input{
		Message: msg,
	}
	result := ""
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}
	return strings.TrimSpace(result), nil
}

func StringPromptIfEmpty(input *string, msg string) error {
	if *input != "" {
		return nil
	}

	prompt := []*survey.Question{{
		Prompt:   &survey.Input{Message: msg},
		Validate: survey.Required,
	}}
	if err := survey.Ask(prompt, input); err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}
	*input = strings.TrimSpace(*input)
	return nil
}

func SelectPromptIfEmpty(input *string, msg string, options []string, def string) error {
	if *input != "" {
		return nil
	}
	res, err := SelectPrompt(msg, options, def)
	if err != nil {
		return err
	}
	*input = res
	return nil
}
