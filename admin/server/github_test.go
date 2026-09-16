package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v71/github"
	"github.com/stretchr/testify/require"
)

func TestListOrganizationsPagination(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`[{"login":"org3"}]`))
			return
		}
		w.Header().Set("Link", fmt.Sprintf(`<%s%s?page=2>; rel="next"`, srv.URL, r.URL.Path))
		_, _ = w.Write([]byte(`[{"login":"org1"},{"login":"org2"}]`))
	}))
	defer srv.Close()

	client := github.NewClient(nil)
	base, err := url.Parse(srv.URL + "/")
	require.NoError(t, err)
	client.BaseURL = base
	client.UploadURL = base

	// Organizations of the authenticated user.
	orgs, err := listOrganizations(context.Background(), client, "")
	require.NoError(t, err)
	require.Len(t, orgs, 3)
	require.Equal(t, "org1", orgs[0].GetLogin())
	require.Equal(t, "org2", orgs[1].GetLogin())
	require.Equal(t, "org3", orgs[2].GetLogin())

	// Organizations of a named user.
	orgs, err = listOrganizations(context.Background(), client, "octocat")
	require.NoError(t, err)
	require.Len(t, orgs, 3)
}

func TestMirrorGitRepo(t *testing.T) {
	src := setupMirrorSourceRepo(t)
	dest := t.TempDir()

	runGitCommand(t, dest, "init", "--bare")
	err := mirrorGitRepo(context.Background(), src, dest, "", "")
	require.NoError(t, err)

	refs := runGitCommand(t, dest, "show-ref")
	require.Contains(t, refs, "refs/heads/main")
	require.Contains(t, refs, "refs/heads/feature")
	require.Contains(t, refs, "refs/tags/v1.0")
}

// setupMirrorSourceRepo creates a local git repository with commits on two branches and a tag.
func setupMirrorSourceRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runGitCommand(t, dir, "init")
	runGitCommand(t, dir, "checkout", "-b", "main")
	runGitCommand(t, dir, "config", "user.name", "Test User")
	runGitCommand(t, dir, "config", "user.email", "test@example.com")

	err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644)
	require.NoError(t, err)
	runGitCommand(t, dir, "add", "readme.txt")
	runGitCommand(t, dir, "commit", "-m", "initial commit")
	runGitCommand(t, dir, "tag", "v1.0")

	runGitCommand(t, dir, "checkout", "-b", "feature")
	err = os.WriteFile(filepath.Join(dir, "feature.txt"), []byte("feature"), 0644)
	require.NoError(t, err)
	runGitCommand(t, dir, "add", "feature.txt")
	runGitCommand(t, dir, "commit", "-m", "add feature")
	return dir
}

func runGitCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed: %s", args, string(output))
	return string(output)
}
