// Package safehttp provides HTTP clients for requests to URLs taken from project files.
// Hosted services must never be steered into internal networks through such URLs (server-side request forgery),
// and credentials attached to a request must never follow a redirect to another host.
package safehttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// NewHTTPClient returns an HTTP client that follows redirects only within the same origin.
// Unless allowPrivate is set, it refuses private network addresses.
func NewHTTPClient(allowPrivate bool, timeout time.Duration) *http.Client {
	return &http.Client{
		Transport:     NewTransport(allowPrivate, timeout),
		Timeout:       timeout,
		CheckRedirect: CheckSameOriginRedirect,
	}
}

// NewTransport returns an HTTP transport for talking to servers configured by project files.
// Unless allowPrivate is set, it refuses to dial addresses on loopback, private or link-local networks,
// which protects hosted services from being pointed at internal services through a connector URL.
func NewTransport(allowPrivate bool, connectTimeout time.Duration) *http.Transport {
	dialer := &net.Dialer{Timeout: connectTimeout}
	t := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: connectTimeout,
		MaxIdleConnsPerHost: 8,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   true,
	}
	if !allowPrivate {
		t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return safeDial(ctx, dialer, network, addr)
		}
	}
	return t
}

// CheckSameOriginRedirect is an http.Client redirect policy that follows at most 5 redirects, all within the origin of the original request.
// Every client that carries credentials on each hop must use it, so a redirect to another host never receives them.
func CheckSameOriginRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return errors.New("stopped after 5 redirects")
	}
	if !sameOrigin(req.URL, via[0].URL) {
		return fmt.Errorf("refusing to follow redirect to a different origin: %s", req.URL.Redacted())
	}
	return nil
}

// IsPrivateHost returns true for hostnames that are obviously local, without resolving them.
// Names that resolve to private addresses are caught later by the transport's dialer.
func IsPrivateHost(host string) bool {
	host = strings.ToLower(strings.TrimSuffix(host, "."))
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return isPrivateIP(ip)
	}
	return false
}

// safeDial resolves addr and dials only its public IP addresses.
func safeDial(ctx context.Context, dialer *net.Dialer, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, ip := range ips {
		if isPrivateIP(ip.IP) {
			lastErr = fmt.Errorf("host %q resolves to a private address, which is not allowed in hosted deployments", host)
			continue
		}
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("host %q did not resolve to any address", host)
	}
	return nil, lastErr
}

// isPrivateIP returns true for addresses that must not be reached from hosted services.
func isPrivateIP(ip net.IP) bool {
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() || ip.IsUnspecified()
}

// sameOrigin returns true if both URLs share scheme, host and port.
func sameOrigin(a, b *url.URL) bool {
	return strings.EqualFold(a.Scheme, b.Scheme) && strings.EqualFold(a.Host, b.Host)
}
