package salesforce

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	force "github.com/ForceCLI/force/lib"
)

// SOQL does not support `SELECT *`. The connector explorer's "Table" mode
// generates exactly that shape, so the driver expands it server-side: when
// the query is `SELECT * FROM <SObject>` (optionally followed by a WHERE /
// ORDER BY / LIMIT etc. clause), the `*` is replaced with the queryable
// field list from DescribeSObject.
var selectStarRegex = regexp.MustCompile(`(?is)^\s*SELECT\s+\*\s+FROM\s+([A-Za-z_][A-Za-z0-9_]*)(\b.*)?$`)

// expandSelectStar returns the original query unchanged unless it matches the
// `SELECT * FROM <SObject>` shape, in which case it rewrites the `*` into an
// explicit field list discovered via Salesforce's describe endpoint.
func expandSelectStar(session *force.Force, query string) (string, error) {
	matches := selectStarRegex.FindStringSubmatch(query)
	if len(matches) == 0 {
		return query, nil
	}
	sobject := matches[1]
	rest := matches[2]

	body, err := session.DescribeSObject(sobject)
	if err != nil {
		return "", fmt.Errorf("describing SObject %q to expand SELECT *: %w", sobject, err)
	}
	fields, err := queryableFieldNames(body)
	if err != nil {
		return "", err
	}
	if len(fields) == 0 {
		return "", fmt.Errorf("SObject %q has no queryable fields", sobject)
	}
	return fmt.Sprintf("SELECT %s FROM %s%s", strings.Join(fields, ", "), sobject, rest), nil
}

// queryableFieldNames extracts field names from a Salesforce describe response,
// skipping field types that are not allowed in a SOQL SELECT list (the
// compound `address` and `location` types — their sub-components like
// BillingStreet / BillingCity are returned as separate fields and are queried
// individually).
func queryableFieldNames(describeBody string) ([]string, error) {
	var desc struct {
		Fields []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"fields"`
	}
	if err := json.Unmarshal([]byte(describeBody), &desc); err != nil {
		return nil, fmt.Errorf("parsing describe response: %w", err)
	}
	out := make([]string, 0, len(desc.Fields))
	for _, f := range desc.Fields {
		if f.Name == "" {
			continue
		}
		switch f.Type {
		case "address", "location":
			// Compound types are not selectable; their components appear
			// as their own field entries.
			continue
		}
		out = append(out, f.Name)
	}
	return out, nil
}
