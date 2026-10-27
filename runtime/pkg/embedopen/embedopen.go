// Package embedopen implements the "rill_open" param of open links in emails from reports and alerts created from an embed.
// The runtime appends the param to the embedder's open URL.
// The embedder forwards it to the admin service's GetIFrame API without reading it,
// and the admin service decodes it to open the report or alert in the embed.
//
// The param is an unpadded base64url encoding of a JSON object.
// It isn't signed: it only identifies a resource, and the embed access token decides access.
package embedopen

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

// QueryParam is the name of the query param that Rill appends to the embedder's open URL.
const QueryParam = "rill_open"

// Kind is the kind of resource that an open link points to.
type Kind string

const (
	KindReport Kind = "report"
)

// Valid returns true if the kind is known.
func (k Kind) Valid() bool {
	switch k {
	case KindReport:
		return true
	}
	return false
}

// Payload is the information encoded in the param.
type Payload struct {
	Kind Kind   `json:"kind"`
	Name string `json:"name"`
	// ExecutionTime is the time of the run, used to resolve relative time ranges; optional.
	ExecutionTime time.Time `json:"execution_time,omitzero"`
}

// Validate returns an error if the payload is not valid.
func (p *Payload) Validate() error {
	if !p.Kind.Valid() {
		return fmt.Errorf("invalid kind %q", p.Kind)
	}
	if p.Name == "" {
		return errors.New("missing name")
	}
	return nil
}

// Encode encodes the payload as the value of the param.
func Encode(p *Payload) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	// Encode the execution time as an RFC3339 timestamp in UTC without fractional seconds.
	j := *p
	j.ExecutionTime = p.ExecutionTime.UTC().Truncate(time.Second)
	data, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// Decode decodes and validates a value of the param.
func Decode(s string) (*Payload, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid encoding: %w", err)
	}
	p := &Payload{}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("invalid payload: %w", err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Link returns the open URL with the encoded payload added as the param.
// It keeps any existing query params of the open URL.
func Link(openURL string, p *Payload) (string, error) {
	u, err := url.Parse(openURL)
	if err != nil {
		return "", fmt.Errorf("invalid open URL %q: %w", openURL, err)
	}
	v, err := Encode(p)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set(QueryParam, v)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ValidateOpenURL returns an error if the URL is not an absolute http(s) URL.
func ValidateOpenURL(openURL string) error {
	u, err := url.Parse(openURL)
	if err != nil {
		return fmt.Errorf("invalid open URL %q: %w", openURL, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("invalid open URL %q: must be an absolute http or https URL", openURL)
	}
	return nil
}
