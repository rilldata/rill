package gitutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSplitGithubRemote and TestNormalizeGithubRemote pin the URL grammar previously provided
// by go-git's transport.NewEndpoint.
func TestSplitGithubRemote(t *testing.T) {
	cases := []struct {
		remote  string
		account string
		repo    string
		ok      bool
	}{
		{"https://github.com/org/repo.git", "org", "repo", true},
		{"https://github.com/org/repo", "org", "repo", true},
		{"https://user:tok@github.com/org/repo.git", "org", "repo", true},
		{"http://github.com/org/repo.git", "org", "repo", true},
		{"ssh://git@github.com/org/repo.git", "org", "repo", true},
		{"git@github.com:org/repo.git", "org", "repo", true},
		{"github.com:org/repo.git", "org", "repo", true},
		{"git@github.com:22:org/repo.git", "org", "repo", true},
		{"https://github.com:443/org/repo.git", "org", "repo", true},
		{"https://GitHub.com/org/repo.git", "org", "repo", true},
		// invalid or non-GitHub remotes
		{"https://github.com/org/repo/", "", "", false},
		{"https://github.com/repo.git", "", "", false},
		{"https://github.com/org/sub/repo.git", "", "", false},
		{"https://gitlab.com/org/repo.git", "", "", false},
		{"git@gitlab.com:org/repo.git", "", "", false},
		{"/local/path/repo.git", "", "", false},
		{"org/repo", "", "", false},
		{"", "", "", false},
	}
	for _, c := range cases {
		account, repo, ok := SplitGithubRemote(c.remote)
		require.Equal(t, c.ok, ok, "remote %q", c.remote)
		require.Equal(t, c.account, account, "remote %q", c.remote)
		require.Equal(t, c.repo, repo, "remote %q", c.remote)
	}
}

func TestNormalizeGithubRemote(t *testing.T) {
	cases := []struct {
		remote     string
		normalized string
		ok         bool
	}{
		{"https://github.com/org/repo.git", "https://github.com/org/repo.git", true},
		{"https://github.com/org/repo", "https://github.com/org/repo.git", true},
		// credentials must be stripped (relied on by the deploy flow)
		{"https://user:tok@github.com/org/repo.git", "https://github.com/org/repo.git", true},
		{"git@github.com:org/repo.git", "https://github.com/org/repo.git", true},
		{"ssh://git@github.com/org/repo.git", "https://github.com/org/repo.git", true},
		{"http://github.com/org/repo.git", "https://github.com/org/repo.git", true},
		{"https://gitlab.com/org/repo.git", "", false},
		{"/local/path/repo.git", "", false},
	}
	for _, c := range cases {
		normalized, err := NormalizeGithubRemote(c.remote)
		if !c.ok {
			require.Error(t, err, "remote %q", c.remote)
			continue
		}
		require.NoError(t, err, "remote %q", c.remote)
		require.Equal(t, c.normalized, normalized, "remote %q", c.remote)
	}
}

// TestSplitRemote pins the host/path extraction that CloneRepo uses to derive repository names.
func TestSplitRemote(t *testing.T) {
	cases := []struct {
		remote string
		host   string
		path   string
		ok     bool
	}{
		{"https://github.com/org/repo.git", "github.com", "/org/repo.git", true},
		{"git@github.com:org/repo.git", "github.com", "org/repo.git", true},
		{"ssh://git@github.com/org/repo.git", "github.com", "/org/repo.git", true},
		{"/tmp/foo/repo.git", "", "/tmp/foo/repo.git", true},
		{"repo.git", "", "repo.git", true},
	}
	for _, c := range cases {
		host, path, ok := splitRemote(c.remote)
		require.Equal(t, c.ok, ok, "remote %q", c.remote)
		require.Equal(t, c.host, host, "remote %q", c.remote)
		require.Equal(t, c.path, path, "remote %q", c.remote)
	}
}
