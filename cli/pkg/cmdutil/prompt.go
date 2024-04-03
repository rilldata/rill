package cmdutil

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func SelectPrompt(msg string, options []string, def string) string {
	prompt := &survey.Select{
		Message: msg,
		Options: options,
	}

	if slices.Contains(options, def) {
		prompt.Default = def
	}

	result := ""
	if err := survey.AskOne(prompt, &result); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return result
}

func ConfirmPrompt(msg, help string, def bool) bool {
	prompt := &survey.Confirm{
		Message: msg,
		Default: def,
	}

	if help != "" {
		prompt.Help = help
	}

	result := def
	if err := survey.AskOne(prompt, &result); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return result
}

func InputPrompt(msg, def string) string {
	prompt := &survey.Input{
		Message: msg,
		Default: def,
	}
	result := def
	if err := survey.AskOne(prompt, &result); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return strings.TrimSpace(result)
}

func StringPromptIfEmpty(input *string, msg string) {
	if *input != "" {
		return
	}

	prompt := []*survey.Question{{
		Prompt:   &survey.Input{Message: msg},
		Validate: survey.Required,
	}}
	if err := survey.Ask(prompt, input); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	*input = strings.TrimSpace(*input)
}

func SelectPromptIfEmpty(input *string, msg string, options []string, def string) {
	if *input != "" {
		return
	}
	*input = SelectPrompt(msg, options, def)
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
				val = InputPrompt(fmt.Sprintf("Enter the %s", f.Usage), f.DefValue)
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
