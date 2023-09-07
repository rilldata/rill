package duckdbsql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEvaluateBool(t *testing.T) {
	cases := []struct {
		expr    string
		want    bool
		wantErr bool
	}{
		{
			expr: "true",
			want: true,
		},
		{
			expr: "false",
			want: false,
		},
		{
			expr: "1 = 1",
			want: true,
		},
		{
			expr: "1 < 0",
			want: false,
		},
		{
			expr: "true = 'true'",
			want: true,
		},
		{
			expr: "'hello' = 'world'",
			want: false,
		},
		{
			expr: "CASE WHEN 1 = 1 THEN true ELSE false END",
			want: true,
		},
		{
			expr:    "0 + 0",
			wantErr: true,
		},
		{
			expr:    "1 + 1",
			wantErr: true,
		},
		{
			expr:    "CURRENT_TIMESTAMP",
			wantErr: true,
		},
		{
			expr:    "10 * 'foo'",
			wantErr: true,
		},
		{
			expr:    "syntax 'error",
			wantErr: true,
		},
		{
			expr:    "",
			wantErr: true,
		},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprintf("Case%d", i), func(t *testing.T) {
			got, err := EvaluateBool(tc.expr)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.want, got)
			}
		})
	}
}
