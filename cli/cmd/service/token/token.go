package token

import (
	"github.com/rilldata/rill/cli/pkg/cmdutil"
	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"github.com/spf13/cobra"
)

func TokenCmd(ch *cmdutil.Helper) *cobra.Command {
	tokenCmd := &cobra.Command{
		Use:               "token",
		Short:             "Manage service tokens",
		PersistentPreRunE: cmdutil.CheckChain(cmdutil.CheckAuth(ch.Config), cmdutil.CheckOrganization(ch.Config)),
	}

	tokenCmd.AddCommand(ListCmd(ch))
	tokenCmd.AddCommand(IssueCmd(ch))
	tokenCmd.AddCommand(RevokeCmd(ch))

	return tokenCmd
}

func toRow(s *adminv1.ServiceToken) *token {
	var expiresOn string
	if !s.ExpiresOn.AsTime().IsZero() {
		expiresOn = s.ExpiresOn.AsTime().Format(cmdutil.TSFormatLayout)
	}

	return &token{
		ID:        s.Id,
		CreatedOn: s.CreatedOn.AsTime().Format(cmdutil.TSFormatLayout),
		ExpiresOn: expiresOn,
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
