package database

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Drivers is a registry of drivers
var Drivers = make(map[string]Driver)

// Register registers a new driver.
func Register(name string, driver Driver) {
	if Drivers[name] != nil {
		panic(fmt.Errorf("already registered database driver with name '%s'", name))
	}
	Drivers[name] = driver
}

// Open opens a new database connection.
func Open(driver, dsn string) (DB, error) {
	d, ok := Drivers[driver]
	if !ok {
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}

	db, err := d.Open(dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Driver is the interface for DB drivers.
type Driver interface {
	Open(dsn string) (DB, error)
}

// DB is the interface for a database connection.
type DB interface {
	Close() error
	Migrate(ctx context.Context) error
	FindMigrationVersion(ctx context.Context) (int, error)

	FindOrganizations(ctx context.Context) ([]*Organization, error)
	FindOrganizationByName(ctx context.Context, name string) (*Organization, error)
	CreateOrganization(ctx context.Context, name string, description string) (*Organization, error)
	UpdateOrganization(ctx context.Context, name string, description string) (*Organization, error)
	DeleteOrganization(ctx context.Context, name string) error

	FindProjects(ctx context.Context, orgName string) ([]*Project, error)
	FindProjectByName(ctx context.Context, orgName string, name string) (*Project, error)
	CreateProject(ctx context.Context, orgID string, name string, description string) (*Project, error)
	UpdateProject(ctx context.Context, id string, description string) (*Project, error)
	DeleteProject(ctx context.Context, id string) error

	FindUsers(ctx context.Context) ([]*User, error)
	FindUser(ctx context.Context, id string) (*User, error)
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, email, displayName, photoURL string) (*User, error)
	UpdateUser(ctx context.Context, id, displayName, photoURL string) (*User, error)
	DeleteUser(ctx context.Context, id string) error

	FindUserAuthTokens(ctx context.Context, userID string) ([]*UserAuthToken, error)
	FindUserAuthToken(ctx context.Context, id string) (*UserAuthToken, error)
	CreateUserAuthToken(ctx context.Context, opts *CreateUserAuthTokenOptions) (*UserAuthToken, error)
	DeleteUserAuthToken(ctx context.Context, id string) error

	// CreateAuthCode inserts the authorization code data into the store.
	CreateAuthCode(ctx context.Context, code *AuthCode) error
	// FindAuthCodeByDeviceCode retrieves the authorization code data from the store
	FindAuthCodeByDeviceCode(ctx context.Context, deviceCode string) (*AuthCode, error)
	// FindAuthCodeByUserCode retrieves the authorization code data from the store
	FindAuthCodeByUserCode(ctx context.Context, userCode string) (*AuthCode, error)
	// UpdateAuthCode updates the authorization code data in the store
	UpdateAuthCode(ctx context.Context, userCode, userID string, approvalState AuthCodeApprovalState) error
	// DeleteAuthCode deletes the authorization code data from the store
	DeleteAuthCode(ctx context.Context, deviceCode string) error
}

// ErrNotFound is returned for single row queries that return no values.
var ErrNotFound = errors.New("database: not found")

// Entity is an enum representing the entities in this package.
type Entity string

const (
	EntityOrganization  Entity = "Organization"
	EntityProject       Entity = "Project"
	EntityUser          Entity = "User"
	EntityUserAuthToken Entity = "UserAuthToken"
	EntityClient        Entity = "Client"
)

type AuthCodeApprovalState int

const (
	Pending  AuthCodeApprovalState = 0
	Approved AuthCodeApprovalState = 1
	Rejected AuthCodeApprovalState = 2
)

type AuthCode struct {
	ID            string                `db:"id"`
	DeviceCode    string                `db:"device_code"`
	UserCode      string                `db:"user_code"`
	Expiry        time.Time             `db:"expires_at"`
	ApprovalState AuthCodeApprovalState `db:"approval_state"`
	ClientID      string                `db:"client_id"`
	UserID        string                `db:"user_id"`
}

// Organization represents a tenant.
type Organization struct {
	ID          string
	Name        string
	Description string
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
}

// Project represents one Git connection.
// Projects belong to an organization.
type Project struct {
	ID             string
	OrganizationID string `db:"organization_id"`
	Name           string
	Description    string
	CreatedOn      time.Time `db:"created_on"`
	UpdatedOn      time.Time `db:"updated_on"`
}

// User is a person registered in Rill.
// Users may belong to multiple organizations and projects.
type User struct {
	ID          string
	Email       string
	DisplayName string    `db:"display_name"`
	PhotoURL    string    `db:"photo_url"`
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
}

// UserAuthToken is a persistent API token for a user.
type UserAuthToken struct {
	ID           string
	SecretHash   []byte    `db:"secret_hash"`
	UserID       string    `db:"user_id"`
	DisplayName  string    `db:"display_name"`
	AuthClientID *string   `db:"auth_client_id"`
	CreatedOn    time.Time `db:"created_on"`
}

// CreateUserAuthTokenOptions defines options for creating a UserAuthToken.
type CreateUserAuthTokenOptions struct {
	ID           string
	SecretHash   []byte
	UserID       string
	DisplayName  string
	AuthClientID *string
}

// AuthClient is a client that requests and consumes auth tokens.
type AuthClient struct {
	ID          string
	DisplayName string
	CreatedOn   time.Time `db:"created_on"`
	UpdatedOn   time.Time `db:"updated_on"`
}

// Hard-coded auth client IDs (created in the migrations).
const (
	AuthClientIDRillWeb = "12345678-0000-0000-0000-000000000001"
	AuthClientIDRillCLI = "12345678-0000-0000-0000-000000000002"
)
