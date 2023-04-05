package cmdutil

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/lensesio/tableprinter"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

type PreRunCheck func(cmd *cobra.Command, args []string) error

func CheckChain(chain ...PreRunCheck) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		for _, fn := range chain {
			err := fn(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func CheckAuth(cfg *config.Config) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		// This will just check if token is present in the config
		if cfg.IsAuthenticated() {
			return nil
		}

		return fmt.Errorf("not authenticated, please run 'rill auth login'")
	}
}

func CheckOrg(cfg *config.Config) PreRunCheck {
	return func(cmd *cobra.Command, args []string) error {
		if cfg.Org != "" {
			return nil
		}

		return fmt.Errorf("no organization is set, pass `--org` or run `rill org switch`")
	}
}

func Spinner(prefix string) *spinner.Spinner {
	// We can some other spinners from here https://github.com/briandowns/spinner#:~:text=90%20Character%20Sets.%20Some%20examples%20below%3A
	sp := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	sp.Prefix = prefix
	// Other colour and font options can be changed
	err := sp.Color("green", "bold")
	if err != nil {
		fmt.Println("invalid color and attribute list, Error: ", err)
	}

	return sp
}

func TablePrinter(v interface{}) {
	var b strings.Builder
	tableprinter.Print(&b, v)
	fmt.Fprintln(os.Stdout, b.String())
}

func TextPrinter(str string) {
	boldGreen := color.New(color.FgGreen).Add(color.Underline).Add(color.Bold)
	boldGreen.Fprintln(color.Output, str)
}

// Create admin client
func Client(cfg *config.Config) (*client.Client, error) {
	cliVersion := cfg.Version.Number
	if cfg.Version.Number == "" {
		cliVersion = "unknown"
	}

	userAgent := fmt.Sprintf("rill-cli/%v", cliVersion)
	c, err := client.New(cfg.AdminURL, cfg.AdminToken(), userAgent)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func SelectPrompt(msg string, options []string, def string) string {
	prompt := &survey.Select{
		Message: msg,
		Options: options,
		Default: def,
	}
	result := def
	if err := survey.AskOne(prompt, &result); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}
	return result
}

func ConfirmPrompt(msg string, def bool) bool {
	prompt := &survey.Confirm{
		Message: msg,
		Default: def,
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
	return result
}
