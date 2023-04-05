package cmdutil

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/lensesio/tableprinter"
	"github.com/manifoldco/promptui"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
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

func CheckOrganization(cfg *config.Config) PreRunCheck {
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

func SuccessPrinter(str string) {
	boldGreen := color.New(color.FgGreen).Add(color.Bold)
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

func PromptGetSelect(items []string, label string) string {
	prompt := promptui.Select{
		Label: label,
		Items: items,
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

func WarnPrinter(str string) {
	boldYellow := color.New(color.FgYellow).Add(color.Bold)
	boldYellow.Fprintln(color.Output, str)
}

func PrintMembers(members []*adminv1.Member) {
	if len(members) == 0 {
		WarnPrinter("No members found")
		return
	}

	SuccessPrinter("Members list \n")
	TablePrinter(toMemberTable(members))
}

func toMemberTable(members []*adminv1.Member) []*member {
	allMembers := make([]*member, 0, len(members))

	for _, m := range members {
		allMembers = append(allMembers, toMemberRow(m))
	}

	return allMembers
}

func toMemberRow(m *adminv1.Member) *member {
	return &member{
		ID:        m.UserId,
		Name:      m.UserName,
		Email:     m.UserEmail,
		RoleID:    m.RoleId,
		RoleName:  m.RoleName,
		CreatedOn: m.CreatedOn.AsTime().String(),
		UpdatedOn: m.UpdatedOn.AsTime().String(),
	}
}

type member struct {
	ID        string `header:"id" json:"id"`
	Name      string `header:"name" json:"display_name"`
	Email     string `header:"email" json:"email"`
	RoleID    string `header:"role_id" json:"role_id"`
	RoleName  string `header:"role_name" json:"role_name"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_on"`
	UpdatedOn string `header:"updated_on,timestamp(ms|utc|human)" json:"updated_on"`
}
