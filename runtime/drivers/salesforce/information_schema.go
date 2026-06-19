package salesforce

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	force "github.com/ForceCLI/force/lib"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pagination"
)

// AsInformationSchema implements drivers.Handle.
func (c *connection) AsInformationSchema() (drivers.InformationSchema, bool) {
	return c, true
}

// ListDatabaseSchemas returns a single empty database/schema pair. Salesforce
// orgs do not have nested catalogs the way relational warehouses do, so the
// SObjects in the org are exposed as the tables under this one entry.
func (c *connection) ListDatabaseSchemas(ctx context.Context, pageSize uint32, pageToken string) ([]*drivers.DatabaseSchemaInfo, string, error) {
	return []*drivers.DatabaseSchemaInfo{{Database: "", DatabaseSchema: ""}}, "", nil
}

// ListTables returns the queryable SObjects in the connected org.
func (c *connection) ListTables(ctx context.Context, database, databaseSchema string, pageSize uint32, pageToken string) ([]*drivers.TableInfo, string, error) {
	session, err := c.authenticateFromConfig()
	if err != nil {
		return nil, "", err
	}

	sobjects, err := session.ListSobjects()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list Salesforce SObjects: %w", err)
	}

	// Filter to queryable SObjects and pull their API names, then sort so
	// pagination is stable across calls.
	names := make([]string, 0, len(sobjects))
	for _, so := range sobjects {
		if !sobjectBool(so, "queryable") {
			continue
		}
		name, _ := so["name"].(string)
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	limit := pagination.ValidPageSize(pageSize, drivers.DefaultPageSize)
	start := 0
	if pageToken != "" {
		var startAfter string
		if err := pagination.UnmarshalPageToken(pageToken, &startAfter); err != nil {
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
		// Find the first name strictly greater than the previous page's tail.
		start = sort.SearchStrings(names, startAfter)
		if start < len(names) && names[start] == startAfter {
			start++
		}
	}

	end := min(start+limit, len(names))

	res := make([]*drivers.TableInfo, 0, end-start)
	for _, n := range names[start:end] {
		res = append(res, &drivers.TableInfo{Name: n})
	}

	next := ""
	if end < len(names) {
		next = pagination.MarshalPageToken(names[end-1])
	}
	return res, next, nil
}

// GetTable returns the column names and SOQL types for the given SObject.
func (c *connection) GetTable(ctx context.Context, database, databaseSchema, table string) (*drivers.TableMetadata, error) {
	session, err := c.authenticateFromConfig()
	if err != nil {
		return nil, err
	}

	body, err := session.DescribeSObject(table)
	if err != nil {
		return nil, fmt.Errorf("failed to describe SObject %q: %w", table, err)
	}

	schema, err := parseDescribeSObject(body)
	if err != nil {
		return nil, err
	}
	return &drivers.TableMetadata{Schema: schema}, nil
}

// parseDescribeSObject extracts a name→SOQL-type map from the JSON body of a
// Salesforce describe call.
func parseDescribeSObject(body string) (map[string]string, error) {
	var desc struct {
		Fields []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(body), &desc); err != nil {
		return nil, fmt.Errorf("failed to parse SObject describe response: %w", err)
	}

	schema := make(map[string]string, len(desc.Fields))
	for _, f := range desc.Fields {
		if f.Name == "" {
			continue
		}
		schema[f.Name] = f.Type
	}
	return schema, nil
}

// authenticateFromConfig builds a force session using only the driver's
// connection config (no per-source overrides). Used by InformationSchema where
// there is no model/source context.
func (c *connection) authenticateFromConfig() (*force.Force, error) {
	opts := c.authOptions()
	if opts.Endpoint == "" {
		return nil, fmt.Errorf("the property 'endpoint' is required for Salesforce")
	}
	session, err := authenticate(opts)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	return session, nil
}

// authOptions builds an authenticationOptions from the connection config.
// It does not validate required fields; callers should check the result
// (e.g. via selectAuthMode) before using it.
func (c *connection) authOptions() authenticationOptions {
	clientID, _ := c.config["client_id"].(string)
	if clientID == "" {
		clientID = defaultClientID
	}

	endpoint, _ := c.config["endpoint"].(string)
	username, _ := c.config["username"].(string)
	password, _ := c.config["password"].(string)
	key, _ := c.config["key"].(string)
	clientSecret, _ := c.config["client_secret"].(string)

	return authenticationOptions{
		Username:     username,
		Password:     password,
		JWT:          key,
		Endpoint:     endpoint,
		ConnectedApp: clientID,
		ClientSecret: clientSecret,
	}
}

// sobjectBool reads a boolean attribute from a ForceSobject map, defaulting
// to false when the key is missing or not a bool.
func sobjectBool(so force.ForceSobject, key string) bool {
	v, ok := so[key].(bool)
	return ok && v
}
