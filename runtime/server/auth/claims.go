package auth

import (
	"encoding/json"
	"slices"

	"github.com/golang-jwt/jwt/v4"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"google.golang.org/protobuf/encoding/protojson"
)

// ClaimsProvider resolves claims from an underlying source.
type ClaimsProvider interface {
	Claims(instanceID string) *runtime.SecurityClaims
}

// jwtClaims implements a ClaimsProvider that resolves claims from a JWT payload.
type jwtClaims struct {
	jwt.RegisteredClaims
	System    []runtime.Permission            `json:"sys,omitempty"`
	Instances map[string][]runtime.Permission `json:"ins,omitempty"`
	Attrs     map[string]any                  `json:"attr,omitempty"`
	Security  []json.RawMessage               `json:"sec,omitempty"` // []*runtimev1.SecurityRule serialized with protojson
}

var _ ClaimsProvider = (*jwtClaims)(nil)

func (c *jwtClaims) Claims(instanceID string) *runtime.SecurityClaims {
	attrs := c.Attrs
	if attrs == nil {
		attrs = make(map[string]any)
	}
	if c.RegisteredClaims.Subject != "" {
		attrs["id"] = c.RegisteredClaims.Subject
	}

	var permissions []runtime.Permission
	permissions = append(permissions, c.System...)
	if c.Instances != nil {
		permissions = append(permissions, c.Instances[instanceID]...)
	}

	if slices.Contains(permissions, runtime.ManageInstances) {
		return &runtime.SecurityClaims{
			UserID:         c.RegisteredClaims.Subject,
			UserAttributes: attrs,
			Permissions:    permissions,
			SkipChecks:     true,
		}
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
		UserID:          c.RegisteredClaims.Subject,
		UserAttributes:  attrs,
		Permissions:     permissions,
		AdditionalRules: rules,
	}
}

// devJWT implements ClaimsProvider and allows all actions but have user attributes for access policies.
// It is used for mimicking user attributes on local when auth is disabled.
type devJWT struct {
	jwt.RegisteredClaims
	Attrs map[string]any `json:"attr,omitempty"`
}

func NewDevToken(attr map[string]any) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &devJWT{
		Attrs: attr,
	})
	res, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", err
	}

	return res, nil
}

var _ ClaimsProvider = (*devJWT)(nil)

func (c *devJWT) Claims(instanceID string) *runtime.SecurityClaims {
	return &runtime.SecurityClaims{
		UserAttributes: c.Attrs,
		Permissions:    []runtime.Permission{runtime.AllPermissions},
	}
}

// wrappedClaims implements a ClaimsProvider that resolves claims from an in-memory value.
type wrappedClaims struct {
	claims *runtime.SecurityClaims
}

var _ ClaimsProvider = (*wrappedClaims)(nil)

func (c wrappedClaims) Claims(instanceID string) *runtime.SecurityClaims {
	return c.claims
}
