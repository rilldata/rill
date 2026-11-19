package ai

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearchTokens(t *testing.T) {
	require.Equal(t, []string{"list", "pull", "requests", "listpullrequests"}, searchTokens("listPullRequests"))
	require.Equal(t, []string{"git", "hub", "github"}, searchTokens("GitHub"))
	require.Equal(t, []string{"list", "pull", "requests"}, searchTokens("list_pull-requests"))
	require.Equal(t, []string{"get", "url", "v2"}, searchTokens("Get URL: v2!"))
	require.Empty(t, searchTokens(" _- "))
}

func TestSearchExternalTools(t *testing.T) {
	doc := func(name, server, description string) *externalSearchDoc {
		return &externalSearchDoc{
			info:        &ExternalToolInfo{Name: name},
			name:        searchTokens(name),
			server:      searchTokens(server),
			description: searchTokens(description),
		}
	}
	docs := []*externalSearchDoc{
		doc("create_pull_request_review", "GitHub", "Reviews a pull request."),
		doc("list_pull_requests", "GitHub", "Lists pull requests awaiting review."),
		doc("search_issues", "Linear", "Search issues, e.g. for review."),
		doc("get_weather", "Weather", "Gets the forecast."),
	}
	names := func(query string, limit int) []string {
		var res []string
		for _, info := range searchExternalTools(docs, query, limit) {
			res = append(res, info.Name)
		}
		return res
	}

	tests := []struct {
		name  string
		query string
		limit int
		want  []string
	}{
		// Name matches score 3 per word, description matches 1; "request" matches "requests" by prefix.
		{"ranked by score", "pull request review", 5, []string{"create_pull_request_review", "list_pull_requests", "search_issues"}},
		{"limit", "pull request review", 1, []string{"create_pull_request_review"}},
		// The server name scores 2, behind name matches.
		{"server match", "linear pull", 5, []string{"create_pull_request_review", "list_pull_requests", "search_issues"}},
		// The server name scores 2, ahead of description matches.
		{"server before description", "forecast linear", 5, []string{"search_issues", "get_weather"}},
		{"ties by name", "github", 5, []string{"create_pull_request_review", "list_pull_requests"}},
		{"case and repeats ignored", "WEATHER weather", 5, []string{"get_weather"}},
		{"no match", "invoice", 5, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, names(tt.query, tt.limit))
		})
	}
}
