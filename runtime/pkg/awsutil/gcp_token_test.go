package awsutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud.google.com/go/compute/metadata"
	"github.com/stretchr/testify/require"
)

func TestGCPMetadataTokenRetriever(t *testing.T) {
	requestPath := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Metadata-Flavor") != "Google" {
			http.Error(w, "missing metadata header", http.StatusForbidden)
			return
		}
		requestPath <- r.URL.RequestURI()
		_, _ = w.Write([]byte("  signed-token\n"))
	}))
	t.Cleanup(server.Close)
	t.Setenv("GCE_METADATA_HOST", strings.TrimPrefix(server.URL, "http://"))

	retriever := GCPMetadataTokenRetriever{
		Audience: "rill audience/with spaces",
		client:   metadata.NewClient(server.Client()),
	}
	token, err := retriever.GetIdentityToken()
	require.NoError(t, err)
	require.Equal(t, []byte("signed-token"), token)
	require.Equal(t, "/computeMetadata/v1/instance/service-accounts/default/identity?audience=rill+audience%2Fwith+spaces&format=full", <-requestPath)
}

func TestGCPMetadataTokenRetrieverRejectsEmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(" \n"))
	}))
	t.Cleanup(server.Close)
	t.Setenv("GCE_METADATA_HOST", strings.TrimPrefix(server.URL, "http://"))

	retriever := GCPMetadataTokenRetriever{client: metadata.NewClient(server.Client())}
	_, err := retriever.GetIdentityToken()
	require.ErrorContains(t, err, "empty identity token")
}
