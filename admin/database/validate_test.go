package database

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type slugTester struct {
	Name string `validate:"slug"`
}

func TestValidateSlug(t *testing.T) {
	require.NoError(t, Validate(&slugTester{Name: "helloworld"}))
	require.NoError(t, Validate(&slugTester{Name: "_hello-world"}))
	require.NoError(t, Validate(&slugTester{Name: "ab"}))
	require.Error(t, Validate(&slugTester{Name: ""}))
	require.Error(t, Validate(&slugTester{Name: "hello world"}))
	require.Error(t, Validate(&slugTester{Name: "-hello"}))
}

func TestNormalizeProjectTags(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		want    []string
		wantErr string
	}{
		{name: "nil", tags: nil, want: []string{}},
		{name: "lowercases, trims, deduplicates and sorts", tags: []string{" Sandbox", "prod", "PROD ", "stage"}, want: []string{"prod", "sandbox", "stage"}},
		{name: "hyphens, underscores and digits", tags: []string{"team_a", "eu-west-1", "2026"}, want: []string{"2026", "eu-west-1", "team_a"}},
		{name: "letters in other scripts", tags: []string{"운영", "Été"}, want: []string{"été", "운영"}},
		{name: "composed and decomposed spellings are one tag", tags: []string{"\u00e9t\u00e9", "e\u0301te\u0301"}, want: []string{"\u00e9t\u00e9"}},
		{name: "max length", tags: []string{strings.Repeat("a", MaxProjectTagLength)}, want: []string{strings.Repeat("a", MaxProjectTagLength)}},
		{name: "max tags after deduplication", tags: append(numberedTags(MaxProjectTags), "tag-0"), want: numberedTags(MaxProjectTags)},
		{name: "empty", tags: []string{"prod", " "}, wantErr: "must not be empty"},
		{name: "too long", tags: []string{strings.Repeat("a", MaxProjectTagLength+1)}, wantErr: "longer than 40 characters"},
		{name: "space", tags: []string{"data science"}, wantErr: `"data science" must start with a letter or digit`},
		{name: "comma", tags: []string{"prod,stage"}, wantErr: "only letters, digits, hyphens and underscores"},
		{name: "leading hyphen", tags: []string{"-prod"}, wantErr: "must start with a letter or digit"},
		{name: "punctuation", tags: []string{"env:prod"}, wantErr: "only letters, digits, hyphens and underscores"},
		{name: "too many", tags: numberedTags(MaxProjectTags + 1), wantErr: "at most 10 tags"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeProjectTags(tt.tags)
			if tt.wantErr != "" {
				require.ErrorIs(t, err, ErrValidation)
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// numberedTags returns n distinct, sorted tags.
func numberedTags(n int) []string {
	res := make([]string, n)
	for i := range res {
		res[i] = fmt.Sprintf("tag-%d", i)
	}
	slices.Sort(res)
	return res
}
