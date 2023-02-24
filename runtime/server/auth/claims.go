package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

// Permission represents runtime access permissions.
type Permission int

const (
	// System-level permissions
	ManageInstances Permission = 0x00

	// Instance-level permissions
	ReadInstance Permission = 0x11
	EditInstance Permission = 0x12
	ReadRepo     Permission = 0x13
	EditRepo     Permission = 0x14
	ReadOLAP     Permission = 0x15
	ReadMetrics  Permission = 0x16
	ReadObjects  Permission = 0x17
)

// Claims resolves permissions for a requester.
type Claims interface {
	// Can resolves system-level permissions. It returns ErrForbidden if the action is not allowed.
	Can(p Permission) error
	// CanInstance resolves instance-level permissions. It returns ErrForbidden if the action is not allowed.
	CanInstance(instanceID string, p Permission) error
}

// ErrForbidden is returned by Claims when an action is not allowed.
var ErrForbidden = errors.New("not allowed")

// jwtClaims implements Claims and resolve permissions based on a JWT payload.
type jwtClaims struct {
	jwt.RegisteredClaims
	System    []Permission
	Instances map[string][]Permission
}

func (c *jwtClaims) Can(p Permission) error {
	for _, p2 := range c.System {
		if p2 == p {
			return nil
		}
	}
	return ErrForbidden
}

func (c *jwtClaims) CanInstance(instanceID string, p Permission) error {
	for _, p2 := range c.Instances[instanceID] {
		if p2 == p {
			return nil
		}
	}
	return c.Can(p)
}

// openClaims implements Claims and allows all actions.
// It should be used for servers with auth disabled.
type openClaims struct{}

func (c *openClaims) Can(p Permission) error {
	return nil
}

func (c *openClaims) CanInstance(instanceID string, p Permission) error {
	return nil
}
