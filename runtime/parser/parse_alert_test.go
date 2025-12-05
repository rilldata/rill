package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAlertYAML_WithCronDefaultsRefUpdateFalse(t *testing.T) {
	files := map[string]string{
		"rill.yaml": "",
		"alerts/a1.yaml": `
                type: alert
                display_name: Test Cron Alert
                refresh:
                    cron: "* * * * *"
                data:
                    sql: select 1
                notify:
                    email:
                        recipients:
                            - hello@example.com
                `,
	}

	repo := makeRepo(t, files)
	ctx := context.Background()
	p, err := Parse(ctx, repo, "", "", "duckdb")
	require.NoError(t, err)
	// Ensure there are no parse errors for the alert
	require.Empty(t, p.Errors)

	res, ok := p.Resources[ResourceName{Kind: ResourceKindAlert, Name: "a1"}]
	require.True(t, ok)
	require.NotNil(t, res.AlertSpec.RefreshSchedule)
	require.Equal(t, "* * * * *", res.AlertSpec.RefreshSchedule.Cron)
	// When a cron is present and ref_update is omitted, default should be false
	require.False(t, res.AlertSpec.RefreshSchedule.RefUpdate)
}
