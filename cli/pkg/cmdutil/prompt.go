package cmdutil

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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

func SetFlagsByInputPrompts(cmd cobra.Command, flags ...string) error {
	var err error
	var val string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && slices.Contains(flags, f.Name) {
			switch f.Value.Type() {
			case "bool":
				var public bool
				defVal, err := strconv.ParseBool(f.DefValue)
				if err != nil {
					return
				}

				prompt := &survey.Confirm{
					Message: fmt.Sprintf("Confirm \"%s\"?", f.Usage),
					Default: defVal,
				}

				err = survey.AskOne(prompt, &public)
				if err != nil {
					return
				}

				val = fmt.Sprintf("%t", public)
			default:
				val, err = InputPrompt(fmt.Sprintf("Enter the %s", f.Usage), f.DefValue)
				if err != nil {
					fmt.Println("error while input prompt, error:", err)
					return
				}
			}

			err = f.Value.Set(val)
			if err != nil {
				fmt.Println("error while setting values, error:", err)
				return
			}
		}
	})

	return err
}
