package drivers

import (
	"bytes"
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// RegistryStore is implemented by drivers capable of storing and looking up instances and repos.
type RegistryStore interface {
	FindInstances(ctx context.Context) ([]*Instance, error)
	FindInstance(ctx context.Context, id string) (*Instance, error)
	CreateInstance(ctx context.Context, instance *Instance) error
	DeleteInstance(ctx context.Context, id string) error
}

// Instance represents a single data project, meaning one OLAP connection, one repo connection,
// and one catalog connection.
type Instance struct {
	// Identifier
	ID string
	// Driver to connect to for OLAP (options: duckdb, druid)
	OLAPDriver string
	// DSN for connection to OLAP
	OLAPDSN string
	// Driver for reading/editing code artifacts (options: file, metastore)
	RepoDriver string
	// DSN for connecting to repo
	RepoDSN string
	// EmbedCatalog tells the runtime to store the instance's catalog in its OLAP store instead
	// of in the runtime's metadata store. Currently only supported for the duckdb driver.
	EmbedCatalog bool `db:"embed_catalog"`
	// CreatedOn is when the instance was created
	CreatedOn time.Time `db:"created_on"`
	// UpdatedOn is when the instance was last updated in the registry
	UpdatedOn time.Time `db:"updated_on"`
	// EnviornmentVariables
	Env EnviornmentVariables `db:"env"`
}

type EnviornmentVariables map[string]string

func (e EnviornmentVariables) Value() (driver.Value, error) {
	return e.String(), nil
}

func (e EnviornmentVariables) Scan(val interface{}) error {
	env := val.(string)
	if env == "" {
		return nil
	}

	m, err := Parse(env)
	if err != nil {
		return err
	}

	for key, value := range m {
		e[key] = value
	}
	return nil
}

func (e EnviornmentVariables) String() string {
	b := new(bytes.Buffer)
	i := 0
	for key, value := range e {
		fmt.Fprintf(b, "%s=%s", key, value)
		i++
		if i != len(e) {
			fmt.Fprintf(b, ";")
		}
	}
	return b.String()
}

// todo :: find a better place for this
func Parse(envString string) (map[string]string, error) {
	if envString == "" {
		return make(map[string]string), nil
	}

	envs := strings.Split(envString, ";")
	vars := make(map[string]string, len(envs))
	for _, env := range envs {
		keyvalue := strings.Split(env, "=")
		if len(keyvalue) != 2 {
			return nil, fmt.Errorf("invalid env string %q", env)
		}
		vars[keyvalue[0]] = keyvalue[1]
	}
	return vars, nil
}
