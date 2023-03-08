package cmdutil

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/lensesio/tableprinter"
	"github.com/rilldata/rill/cli/pkg/config"
	"github.com/spf13/cobra"
)

func CheckAuth(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// This will just check if token is present in the config
		if cfg.IsAuthenticated() {
			return nil
		}

		return fmt.Errorf("not authenticated, please run 'rill auth login'")
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

func PrintAsTable(v interface{}) {
	var b strings.Builder
	tableprinter.Print(&b, v)
	fmt.Fprintln(os.Stdout, b.String())
}
