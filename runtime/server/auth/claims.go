package auth

import (
	"encoding/json"

	"github.com/golang-jwt/jwt/v4"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
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
	// SecurityClaims returns a representation of the claims for use with runtime package's security policy enforcement.
	SecurityClaims() *runtime.SecurityClaims
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

func (c *jwtClaims) SecurityClaims() *runtime.SecurityClaims {
	if c.Can(ManageInstances) {
		return &runtime.SecurityClaims{
			UserAttributes: c.Attrs,
			SkipChecks:     true,
		}
	}

	attrs := c.Attrs
	if attrs == nil {
		attrs = make(map[string]any)
	}
	attrs["id"] = c.Subject()
	if cid, ok := c.Attrs["creator_id"]; ok {
		attrs["creator_id"] = cid
	}

	var rules []*runtimev1.SecurityRule
	if len(c.Security) > 0 {
		rules = make([]*runtimev1.SecurityRule, len(c.Security))
		for i, data := range c.Security {
			rule := &runtimev1.SecurityRule{}
			err := protojson.Unmarshal(data, rule)
			if err != nil {
				panic(err)
			}
			rules[i] = rule
		}
	}

	return &runtime.SecurityClaims{
		UserAttributes:  attrs,
		AdditionalRules: rules,
	}
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

func (c openClaims) SecurityClaims() *runtime.SecurityClaims {
	return &runtime.SecurityClaims{
		UserAttributes: map[string]any{"admin": true},
		SkipChecks:     true,
	}
}

// anonClaims implements Claims with no permissions.
// It is used for unauthorized requests when auth is enabled.
type anonClaims struct{}

var _ Claims = (*anonClaims)(nil)

var emptySecurityClaims = &runtime.SecurityClaims{
	UserAttributes: map[string]any{},
}

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

func (c anonClaims) SecurityClaims() *runtime.SecurityClaims {
	return emptySecurityClaims
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

func (c devJWTClaims) SecurityClaims() *runtime.SecurityClaims {
	return &runtime.SecurityClaims{
		UserAttributes: c.Attrs,
	}
}
