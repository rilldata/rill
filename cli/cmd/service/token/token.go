package token

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	"github.com/rilldata/rill/cli/pkg/config"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func TokenCmd(cfg *config.Config) *cobra.Command {
	tokenCmd := &cobra.Command{
		Use:               "token",
		Short:             "Manage service tokens",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(cfg), cmdutil.CheckOrganization(cfg)),
	}

	tokenCmd.AddCommand(IssueCmd(cfg))
	tokenCmd.AddCommand(ListCmd(cfg))
	tokenCmd.AddCommand(RevokeCmd(cfg))

	return tokenCmd
}

func toRow(s *adminv1.ServiceToken) *token {
	return &token{
		ID:        s.Id,
		CreatedOn: s.CreatedOn.AsTime().Format(cmdutil.TSFormatLayout),
		ExpiresOn: s.ExpiresOn.AsTime().Format(cmdutil.TSFormatLayout),
	}
}

func toTable(tkns []*adminv1.ServiceToken) []*token {
	tokens := make([]*token, 0, len(tkns))

	for _, t := range tkns {
		tokens = append(tokens, toRow(t))
	}

	return tokens
}

type token struct {
	ID        string `header:"id" json:"id"`
	CreatedOn string `header:"created_on,timestamp(ms|utc|human)" json:"created_on"`
	ExpiresOn string `header:"expires_on,timestamp(ms|utc|human)" json:"expires_on"`
}
