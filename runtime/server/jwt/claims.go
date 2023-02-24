package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

var ErrForbidden = errors.New("not allowed")

type Permission int

const (
	// System permissions
	ManageInstances Permission = 0x00
	// Instance permissions
	ReadInstance    Permission = 0x11
	EditInstance    Permission = 0x12
	ReadRepo        Permission = 0x13
	EditRepo        Permission = 0x14
	ReadOLAP        Permission = 0x15
	ReadMetrics     Permission = 0x16
	ReadObjects     Permission = 0x17
	ReadObjectState Permission = 0x18
)

type Claims interface {
	Can(p Permission) error
	CanInstance(instanceID string, p Permission) error
}

// jwtClaims represents the payload of a JWT
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

// openClaims allows all actions. It's used for servers with auth disabled.
type openClaims struct{}

func (c *openClaims) Can(p Permission) error {
	return nil
}

func (c *openClaims) CanInstance(instanceID string, p Permission) error {
	return nil
}
