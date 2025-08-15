package client

import (
	"context"
	"fmt"
	"net/url"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// Retry policy for requests to the runtime.
// For details, see https://github.com/grpc/grpc/blob/master/doc/service_config.md and https://grpc.io/docs/guides/retry/.
const retryPolicy = `{"methodConfig": [{
	"name": [{}],
	"retryPolicy": {
		"maxAttempts": 5,
		"initialBackoff": ".1s",
		"maxBackoff": "20s",
		"backoffMultiplier": 10,
		"retryableStatusCodes": ["UNAVAILABLE"]
	}
}]}`

// Client connects to a runtime server.
// It's a thin wrapper around the generated gRPC client for proto/rill/runtime/v1.
type Client struct {
	runtimev1.RuntimeServiceClient
	conn *grpc.ClientConn
}

// New creates a new Client and opens a connection. You must call Close() when done with the client.
func New(runtimeHost, bearerToken string) (*Client, error) {
	uri, err := url.Parse(runtimeHost)
	if err != nil {
		return nil, err
	}

	opts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(retryPolicy),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	if uri.Scheme == "http" {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, ""))) // NOTE: Defaults to host's root certs
		// There must be a port. Default to TLS port.
		if uri.Port() == "" {
			uri.Host = fmt.Sprintf("%s:443", uri.Host)
		}
	}

	if bearerToken != "" {
		secure := uri.Scheme != "http"
		opts = append(opts, grpc.WithPerRPCCredentials(bearerAuth{token: bearerToken, secure: secure}))
	}

	conn, err := grpc.NewClient(uri.Host, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		RuntimeServiceClient: runtimev1.NewRuntimeServiceClient(conn),
		conn:                 conn,
	}, nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// ConnectorServiceClient returns a ConnectorServiceClient using the same connection.
func (c *Client) ConnectorServiceClient() runtimev1.ConnectorServiceClient {
	return runtimev1.NewConnectorServiceClient(c.conn)
}

// QueryServiceClient returns a QueryServiceClient using the same connection.
func (c *Client) QueryServiceClient() runtimev1.QueryServiceClient {
	return runtimev1.NewQueryServiceClient(c.conn)
}

// bearerAuth implements credentials.PerRPCCredentials for adding a bearer authorization token in the metadata of a gRPC client's requests.
type bearerAuth struct {
	token  string
	secure bool
}

func (t bearerAuth) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + t.token,
	}, nil
}

func (t bearerAuth) RequireTransportSecurity() bool {
	return t.secure
}
