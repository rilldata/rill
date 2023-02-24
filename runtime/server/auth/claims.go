package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

// Permission represents runtime access permissions.
type Permission int

const (
	// System-level permissions
	ManageInstances Permission = 0x00

	// Instance-level permissions
	ReadInstance  Permission = 0x11
	EditInstance  Permission = 0x12
	ReadRepo      Permission = 0x13
	EditRepo      Permission = 0x14
	ReadObjects   Permission = 0x15
	ReadOLAP      Permission = 0x16
	ReadMetrics   Permission = 0x17
	ReadProfiling Permission = 0x18
)

// Claims resolves permissions for a requester.
type Claims interface {
	// Can resolves system-level permissions.
	Can(p Permission) bool
	// CanInstance resolves instance-level permissions.
	CanInstance(instanceID string, p Permission) bool
}

// jwtClaims implements Claims and resolve permissions based on a JWT payload.
type jwtClaims struct {
	jwt.RegisteredClaims
	System    []Permission
	Instances map[string][]Permission
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

// openClaims implements Claims and allows all actions.
// It should be used for servers with auth disabled.
type openClaims struct{}

func (c *openClaims) Can(p Permission) bool {
	return true
}

func (c *openClaims) CanInstance(instanceID string, p Permission) bool {
	return true
}
