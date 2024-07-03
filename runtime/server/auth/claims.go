package auth

import (
	"encoding/json"

	"github.com/golang-jwt/jwt/v4"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Claims resolves permissions for a requester.
type Claims interface {
	// Subject returns the token subject if present (usually a user or service ID)
	Subject() string
	// Can resolves system-level permissions.
	Can(p Permission) bool
	// CanInstance resolves instance-level permissions.
	CanInstance(instanceID string, p Permission) bool
	// Attributes returns the token attributes used in template rendering.
	Attributes() map[string]any
	// SecurityRules are optional security rules to apply *in addition* to the rules defined in the requested resources themselves.
	// This provides a way to embed/inline additional security restrictions for a specific token.
	// This option is currently leveraged by the admin service to enforce restrictions for magic auth tokens.
	SecurityRules() []*runtimev1.SecurityRule
}

// jwtClaims implements Claims and resolve permissions based on a JWT payload.
type jwtClaims struct {
	jwt.RegisteredClaims
	System    []Permission            `json:"sys,omitempty"`
	Instances map[string][]Permission `json:"ins,omitempty"`
	Attrs     map[string]any          `json:"attr,omitempty"`
	Security  []json.RawMessage       `json:"sec,omitempty"` // []*runtimev1.SecurityRule serialized with protojson
}

var _ Claims = (*jwtClaims)(nil)

func (c *jwtClaims) Subject() string {
	return c.RegisteredClaims.Subject
}

func (c *jwtClaims) Can(p Permission) bool {
	for _, p2 := range c.System {
		if p2 == p {
			return true
		}
	}
	return false
}

func (c *jwtClaims) CanInstance(instanceID string, p Permission) bool {
	for _, p2 := range c.Instances[instanceID] {
		if p2 == p {
			return true
		}
	}
	return c.Can(p)
}

func (c *jwtClaims) Attributes() map[string]any {
	return c.Attrs
}

func (c jwtClaims) SecurityRules() []*runtimev1.SecurityRule {
	if len(c.Security) == 0 {
		return nil
	}

	rules := make([]*runtimev1.SecurityRule, len(c.Security))
	for i, data := range c.Security {
		rule := &runtimev1.SecurityRule{}
		err := protojson.Unmarshal(data, rule)
		if err != nil {
			panic(err)
		}
		rules[i] = rule
	}

	return rules
}

// openClaims implements Claims and allows all actions.
// It is used for servers with auth disabled.
type openClaims struct{}

var _ Claims = (*openClaims)(nil)

func (c openClaims) Subject() string {
	return ""
}

func (c openClaims) Can(p Permission) bool {
	return true
}

func (c openClaims) CanInstance(instanceID string, p Permission) bool {
	return true
}

func (c openClaims) Attributes() map[string]any {
	return nil
}

func (c openClaims) SecurityRules() []*runtimev1.SecurityRule {
	return nil
}

// anonClaims implements Claims with no permissions.
// It is used for unauthorized requests when auth is enabled.
type anonClaims struct{}

var _ Claims = (*anonClaims)(nil)

func (c anonClaims) Subject() string {
	return ""
}

func (c anonClaims) Can(p Permission) bool {
	return false
}

func (c anonClaims) CanInstance(instanceID string, p Permission) bool {
	return false
}

func (c anonClaims) Attributes() map[string]any {
	return nil
}

func (c anonClaims) SecurityRules() []*runtimev1.SecurityRule {
	return nil
}

// devJWTClaims implements Claims and allows all actions but have user attributes for access policies.
// It is used for mimicking user attributes on local when auth is disabled.
type devJWTClaims struct {
	jwt.RegisteredClaims
	Attrs map[string]any `json:"attr,omitempty"`
}

var _ Claims = (*devJWTClaims)(nil)

func (c devJWTClaims) Subject() string {
	return ""
}

func (c devJWTClaims) Can(p Permission) bool {
	return true
}

func (c devJWTClaims) CanInstance(instanceID string, p Permission) bool {
	return true
}

func (c devJWTClaims) Attributes() map[string]any {
	return c.Attrs
}

func (c devJWTClaims) SecurityRules() []*runtimev1.SecurityRule {
	return nil
}
