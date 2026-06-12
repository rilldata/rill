package salesforce

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	force "github.com/ForceCLI/force/lib"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestBulk2QueryJobNextStreamsPages verifies that Next() streams each results
// page to a temp file and follows the Sforce-Locator cursor until Salesforce
// signals the final page with an empty locator.
func TestBulk2QueryJobNextStreamsPages(t *testing.T) {
	page1 := "Id,Name\n001000000000001AAA,Test 1\n"
	page2 := "Id,Name\n001000000000002BBB,Test 2\n"

	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "maxRecords=100000") {
			t.Errorf("expected maxRecords=100000, got query %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "text/csv")
		calls++
		if calls == 1 {
			w.Header().Set("Sforce-Locator", "PAGE2")
			_, _ = w.Write([]byte(page1))
			return
		}
		if !strings.Contains(r.URL.RawQuery, "locator=PAGE2") {
			t.Errorf("expected locator=PAGE2 on second call, got query %q", r.URL.RawQuery)
		}
		w.Header().Set("Sforce-Locator", "null")
		_, _ = w.Write([]byte(page2))
	}))
	defer server.Close()

	job := makeBulk2QueryJob(testSession(server.URL), zap.NewNop())
	job.jobID = "7501234567890QUERY"
	defer func() { require.NoError(t, job.Close()) }()

	got := readAllPages(t, job)
	require.Equal(t, []string{page1, page2}, got)
	require.Equal(t, 2, calls)
}

// TestBulk2QueryJobNextSinglePage verifies that a job whose first page is also
// its last (empty locator) returns one file and then io.EOF.
func TestBulk2QueryJobNextSinglePage(t *testing.T) {
	csv := "Id,Name\n001000000000001AAA,Only\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Sforce-Locator", "null")
		_, _ = w.Write([]byte(csv))
	}))
	defer server.Close()

	job := makeBulk2QueryJob(testSession(server.URL), zap.NewNop())
	job.jobID = "7501234567890QUERY"
	defer func() { require.NoError(t, job.Close()) }()

	require.Equal(t, []string{csv}, readAllPages(t, job))
}

// TestBulk2QueryJobNextCleansUpBetweenPages verifies that, by default, the temp
// file from the previous page is removed when the next page is fetched.
func TestBulk2QueryJobNextCleansUpBetweenPages(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		calls++
		if calls == 1 {
			w.Header().Set("Sforce-Locator", "PAGE2")
			_, _ = w.Write([]byte("Id\n1\n"))
			return
		}
		w.Header().Set("Sforce-Locator", "null")
		_, _ = w.Write([]byte("Id\n2\n"))
	}))
	defer server.Close()

	job := makeBulk2QueryJob(testSession(server.URL), zap.NewNop())
	job.jobID = "7501234567890QUERY"
	defer func() { require.NoError(t, job.Close()) }()

	first, err := job.Next(context.Background())
	require.NoError(t, err)
	require.Len(t, first, 1)

	_, err = job.Next(context.Background())
	require.NoError(t, err)

	_, statErr := os.Stat(first[0])
	require.True(t, os.IsNotExist(statErr), "previous page temp file should be removed")
}

// TestBulk2QueryJobNextHttpError verifies that an HTTP error from Salesforce is
// surfaced and no temp file is leaked.
func TestBulk2QueryJobNextHttpError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`[{"errorCode":"INVALIDJOB","message":"job not found"}]`))
	}))
	defer server.Close()

	job := makeBulk2QueryJob(testSession(server.URL), zap.NewNop())
	job.jobID = "7501234567890QUERY"
	defer func() { require.NoError(t, job.Close()) }()

	_, err := job.Next(context.Background())
	require.Error(t, err)
	require.Empty(t, job.tempFilePaths)
}

func testSession(instanceURL string) *force.Force {
	return &force.Force{
		Credentials: &force.ForceSession{
			InstanceUrl: instanceURL,
			AccessToken: "test-token",
		},
	}
}

func readAllPages(t *testing.T, job *bulk2QueryJob) []string {
	t.Helper()
	var contents []string
	for {
		paths, err := job.Next(context.Background())
		if errors.Is(err, io.EOF) {
			return contents
		}
		require.NoError(t, err)
		for _, p := range paths {
			data, err := os.ReadFile(p)
			require.NoError(t, err)
			contents = append(contents, string(data))
		}
	}
}
