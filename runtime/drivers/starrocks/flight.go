package starrocks

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/apache/arrow-go/v18/arrow/flight"
	"github.com/apache/arrow-go/v18/arrow/flight/flightsql"
	mysql "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// initFlightClient initializes the Arrow Flight SQL client.
// Must be called after MySQL connection is established (for host resolution).
func (c *connection) initFlightClient() error {
	host, err := c.flightSQLHost()
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", host, c.configProp.FlightSQLPort)

	// Build gRPC dial options
	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(256 * 1024 * 1024), // 256MB max message size
		),
	}

	// Configure TLS
	var cred credentials.TransportCredentials
	if c.configProp.SSL {
		cred = credentials.NewClientTLSFromCert(nil, "")
	} else {
		cred = insecure.NewCredentials()
	}
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(cred))

	// Add authentication interceptor
	username, password := c.flightSQLAuth()
	if username != "" {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(basicAuthUnaryInterceptor(username, password)))
		dialOpts = append(dialOpts, grpc.WithStreamInterceptor(basicAuthStreamInterceptor(username, password)))
	}

	client, err := flightsql.NewClient(addr, nil, nil, dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to create Flight SQL client at %s: %w", addr, err)
	}

	c.flightClient = client

	// Verify the Flight SQL connection with a simple Execute + DoGet roundtrip.
	// This ensures both FE (Execute) and BE (DoGet) are reachable.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	info, err := client.Execute(ctx, "SELECT 1")
	if err != nil {
		client.Close()
		c.flightClient = nil
		return fmt.Errorf("failed to verify Flight SQL connection at %s: %w", addr, err)
	}

	if len(info.Endpoint) > 0 {
		reader, err := c.doGetFromEndpoint(ctx, info.Endpoint[0])
		if err != nil {
			client.Close()
			c.flightClient = nil
			return fmt.Errorf("flight SQL DoGet verification failed at %s (FE Execute succeeded but BE DoGet failed; check flight_sql_be_addr config): %w", addr, err)
		}
		reader.Release()
	}

	c.logger.Info("Arrow Flight SQL client initialized", zap.String("addr", addr))
	return nil
}

// flightSQLHost returns the host to use for Arrow Flight SQL connection.
func (c *connection) flightSQLHost() (string, error) {
	if c.configProp.Host != "" {
		return c.configProp.Host, nil
	}
	// When using DSN mode, parse host from the MySQL DSN
	if c.configProp.DSN != "" {
		parsed, err := mysql.ParseDSN(c.configProp.DSN)
		if err != nil {
			return "", fmt.Errorf("failed to parse DSN for Flight SQL host: %w", err)
		}
		// Addr is "host:port"; use net.SplitHostPort for correct IPv6 handling
		host, _, err := net.SplitHostPort(parsed.Addr)
		if err != nil {
			// No port suffix; use Addr as-is (e.g., just a hostname)
			return parsed.Addr, nil
		}
		return host, nil
	}
	return "", fmt.Errorf("cannot determine Flight SQL host: neither host nor DSN is set")
}

// flightSQLAuth returns the username and password for Flight SQL authentication.
func (c *connection) flightSQLAuth() (string, string) {
	if c.configProp.Username != "" {
		return c.configProp.Username, c.configProp.Password
	}
	// When using DSN mode, parse credentials from the MySQL DSN
	if c.configProp.DSN != "" {
		parsed, err := mysql.ParseDSN(c.configProp.DSN)
		if err != nil {
			return "root", ""
		}
		return parsed.User, parsed.Passwd
	}
	return "root", ""
}

// doGetFromEndpoint performs a DoGet call, routing to the correct node (FE or BE)
// based on the endpoint's Location URIs. StarRocks returns data from BE nodes;
// the FE client cannot serve DoGet for tickets issued to BE.
// BE clients are pooled in the connection and reused across queries.
func (c *connection) doGetFromEndpoint(ctx context.Context, ep *flight.FlightEndpoint) (*flight.Reader, error) {
	// If no Location is specified, try the FE client (some setups proxy through FE)
	if len(ep.Location) == 0 {
		return c.flightClient.DoGet(ctx, ep.GetTicket())
	}

	// Parse the BE address from the Location URI
	beAddr, err := parseFlightLocation(ep.Location[0].GetUri())
	if err != nil {
		return nil, fmt.Errorf("failed to parse Flight endpoint location URI %q: %w", ep.Location[0].GetUri(), err)
	}

	// If a BE address override is configured (e.g., for Docker port mapping),
	// use it instead of the address from the Location URI.
	if c.configProp.FlightSQLBEAddr != "" {
		beAddr = c.configProp.FlightSQLBEAddr
	}

	c.logger.Debug("Routing DoGet to BE node", zap.String("be_addr", beAddr))

	beClient, err := c.getOrCreateBEClient(beAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to get BE flight client at %s: %w", beAddr, err)
	}

	reader, err := beClient.DoGet(ctx, ep.GetTicket())
	if err != nil {
		// Evict the cached client so the next query creates a fresh connection.
		// This handles BE restarts or network disruptions.
		c.evictBEClient(beAddr)
		return nil, fmt.Errorf("flight sql doget from BE %s: %w", beAddr, err)
	}

	return reader, nil
}

// getOrCreateBEClient returns a cached BE Flight SQL client for the given address,
// creating one if it doesn't exist. Clients are pooled for the lifetime of the
// connection and cleaned up in Close().
func (c *connection) getOrCreateBEClient(addr string) (*flightsql.Client, error) {
	c.beClientsMu.Lock()
	defer c.beClientsMu.Unlock()

	if c.beClients == nil {
		c.beClients = make(map[string]*flightsql.Client)
	}

	if client, ok := c.beClients[addr]; ok {
		return client, nil
	}

	client, err := c.createBEFlightClient(addr)
	if err != nil {
		return nil, err
	}

	c.beClients[addr] = client
	c.logger.Debug("Created and cached BE flight client", zap.String("addr", addr))
	return client, nil
}

// evictBEClient removes a cached BE client after a connection failure,
// so the next query creates a fresh connection.
func (c *connection) evictBEClient(addr string) {
	c.beClientsMu.Lock()
	defer c.beClientsMu.Unlock()

	if client, ok := c.beClients[addr]; ok {
		_ = client.Close()
		delete(c.beClients, addr)
		c.logger.Debug("Evicted failed BE flight client", zap.String("addr", addr))
	}
}

// createBEFlightClient creates a new Flight SQL client for a BE node address.
// Uses the same auth credentials and TLS settings as the FE client.
func (c *connection) createBEFlightClient(addr string) (*flightsql.Client, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(256 * 1024 * 1024),
		),
	}

	var cred credentials.TransportCredentials
	if c.configProp.SSL {
		cred = credentials.NewClientTLSFromCert(nil, "")
	} else {
		cred = insecure.NewCredentials()
	}
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(cred))

	username, password := c.flightSQLAuth()
	if username != "" {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(basicAuthUnaryInterceptor(username, password)))
		dialOpts = append(dialOpts, grpc.WithStreamInterceptor(basicAuthStreamInterceptor(username, password)))
	}

	return flightsql.NewClient(addr, nil, nil, dialOpts...)
}

// parseFlightLocation parses a Flight Location URI (e.g., "grpc+tcp://host:port")
// and returns the "host:port" address.
func parseFlightLocation(uri string) (string, error) {
	if uri == "" {
		return "", fmt.Errorf("empty location URI")
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("invalid location URI %q: %w", uri, err)
	}

	host := parsed.Hostname()
	port := parsed.Port()
	if host == "" {
		return "", fmt.Errorf("no host in location URI %q", uri)
	}
	if port == "" {
		return host, nil
	}
	return host + ":" + port, nil
}

// basicAuthUnaryInterceptor creates a gRPC unary interceptor for basic auth.
func basicAuthUnaryInterceptor(username, password string) grpc.UnaryClientInterceptor {
	encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Basic "+encoded)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// basicAuthStreamInterceptor creates a gRPC stream interceptor for basic auth.
func basicAuthStreamInterceptor(username, password string) grpc.StreamClientInterceptor {
	encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Basic "+encoded)
		return streamer(ctx, desc, cc, method, opts...)
	}
}
