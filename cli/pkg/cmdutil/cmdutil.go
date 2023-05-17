package cmdutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/lensesio/tableprinter"
	"github.com/rilldata/rill/admin/client"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	TSFormatLayout     = "2006-01-02 15:04:05"
	defaultProjectName = "rill-untitled"
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

		return fmt.Errorf("not authenticated, please run 'rill login'")
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
	fmt.Fprint(os.Stdout, b.String())
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

func SelectPrompt(msg string, options []string, def string) string {
	prompt := &survey.Select{
		Message: msg,
		Options: options,
	}

	if contains(options, def) {
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

func InputPrompt(msg, def string) (string, error) {
	prompt := &survey.Input{
		Message: msg,
		Default: def,
	}
	result := def
	if err := survey.AskOne(prompt, &result); err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}
	return strings.TrimSpace(result), nil
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
}

func SelectPromptIfEmpty(input *string, msg string, options []string, def string) {
	if *input != "" {
		return
	}
	*input = SelectPrompt(msg, options, def)
}

func ProjectExists(ctx context.Context, c *client.Client, orgName, projectName string) (bool, error) {
	_, err := c.GetProject(ctx, &adminv1.GetProjectRequest{OrganizationName: orgName, Name: projectName})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func OrgExists(ctx context.Context, c *client.Client, name string) (bool, error) {
	_, err := c.GetOrganization(ctx, &adminv1.GetOrganizationRequest{Name: name})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
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

	TablePrinter(toMemberTable(members))
}

func PrintInvites(invites []*adminv1.UserInvite) {
	if len(invites) == 0 {
		return
	}

	SuccessPrinter("Pending user invites")
	TablePrinter(toInvitesTable(invites))
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
		Name:      m.UserName,
		Email:     m.UserEmail,
		RoleName:  m.RoleName,
		CreatedOn: m.CreatedOn.AsTime().Format(TSFormatLayout),
		UpdatedOn: m.UpdatedOn.AsTime().Format(TSFormatLayout),
	}
}

type member struct {
	Name      string `header:"name" json:"display_name"`
	Email     string `header:"email" json:"email"`
	RoleName  string `header:"role" json:"role_name"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_on"`
	UpdatedOn string `header:"updated_on,timestamp(ms|utc|human)" json:"updated_on"`
}

func toInvitesTable(invites []*adminv1.UserInvite) []*userInvite {
	allInvites := make([]*userInvite, 0, len(invites))

	for _, i := range invites {
		allInvites = append(allInvites, toInviteRow(i))
	}
	return allInvites
}

func toInviteRow(i *adminv1.UserInvite) *userInvite {
	return &userInvite{
		Email:     i.Email,
		RoleName:  i.Role,
		InvitedBy: i.InvitedBy,
	}
}

type userInvite struct {
	Email     string `header:"email" json:"email"`
	RoleName  string `header:"role" json:"role_name"`
	InvitedBy string `header:"invited_by" json:"invited_by"`
}

func contains(vals []string, key string) bool {
	for _, s := range vals {
		if key == s {
			return true
		}
	}
	return false
}

// ProjectNames returns names of all the projects in org deployed from githubURL
func ProjectNamesByGithubURL(ctx context.Context, c *client.Client, org, githubURL string) ([]string, error) {
	resp, err := c.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{
		OrganizationName: org,
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, p := range resp.Projects {
		if strings.EqualFold(p.GithubUrl, githubURL) {
			names = append(names, p.Name)
		}
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("No project with githubURL %q exist in org %q", githubURL, org)
	}

	return names, nil
}

func IsNameExistsErr(err error) bool {
	if st, ok := status.FromError(err); ok && st != nil {
		exists := strings.Contains(st.Message(), "violates unique constraint")
		return exists
	}
	return false
}

func OrgNames(ctx context.Context, c *client.Client) ([]string, error) {
	resp, err := c.ListOrganizations(ctx, &adminv1.ListOrganizationsRequest{})
	if err != nil {
		return nil, err
	}

	if len(resp.Organizations) == 0 {
		return nil, fmt.Errorf("You are not a member of any orgs")
	}

	var orgNames []string
	for _, org := range resp.Organizations {
		orgNames = append(orgNames, org.Name)
	}

	return orgNames, nil
}

func ProjectNamesByOrg(ctx context.Context, c *client.Client, orgName string) ([]string, error) {
	resp, err := c.ListProjectsForOrganization(ctx, &adminv1.ListProjectsForOrganizationRequest{OrganizationName: orgName})
	if err != nil {
		return nil, err
	}

	if len(resp.Projects) == 0 {
		return nil, fmt.Errorf("No projects found for org %q", orgName)
	}

	var projNames []string
	for _, proj := range resp.Projects {
		projNames = append(projNames, proj.Name)
	}

	return projNames, nil
}

func DefaultProjectName() string {
	// Get full path to project
	absPath, err := filepath.Abs(defaultProjectName)
	if err != nil {
		return ""
	}

	_, err = os.Stat(absPath)
	if os.IsNotExist(err) {
		return defaultProjectName
	}

	num := 1
	for err == nil {
		absPath := fmt.Sprint(absPath, num)

		_, err = os.Stat(absPath)
		if os.IsNotExist(err) {
			return fmt.Sprint(defaultProjectName, num)
		}
		num++
	}

	return ""
}

func SetFlagsByInputPrompts(cmd cobra.Command, flags ...string) error {
	var err error
	var value string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && slices.Contains(flags, f.Name) {
			if f.Value.Type() == "string" {
				value, err = InputPrompt(fmt.Sprintf("Enter the %s", f.Usage), "")
				if err != nil {
					fmt.Println("error while input prompt, error:", err)
					return
				}
			}

			if f.Value.Type() == "bool" {
				var public bool
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("Confirm \"%s\"?", f.Usage),
				}

				err = survey.AskOne(prompt, &public)
				if err != nil {
					return
				}

				value = fmt.Sprintf("%t", public)
			}

			err = f.Value.Set(value)
			if err != nil {
				fmt.Println("error while setting values, error:", err)
				return
			}
		}
	})

	return err
}

func FetchUserID(ctx context.Context, cfg *config.Config) (string, error) {
	c, err := Client(cfg)
	if err != nil {
		return "", err
	}
	defer c.Close()
	user, err := c.GetCurrentUser(ctx, &adminv1.GetCurrentUserRequest{})
	if err != nil {
		return "", err
	}
	return user.GetUser().GetId(), nil
}
