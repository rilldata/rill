package salesforce

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryableFieldNames(t *testing.T) {
	body := `{
		"name": "Account",
		"fields": [
			{"name": "Id", "type": "id"},
			{"name": "Name", "type": "string"},
			{"name": "BillingAddress", "type": "address"},
			{"name": "BillingStreet", "type": "string"},
			{"name": "BillingCity", "type": "string"},
			{"name": "Location", "type": "location"},
			{"name": "", "type": "string"}
		]
	}`
	got, err := queryableFieldNames(body)
	require.NoError(t, err)
	// Compound types (address, location) and empty names are filtered;
	// their atomic sub-components remain because Salesforce describes them
	// as separate fields.
	require.Equal(t, []string{"Id", "Name", "BillingStreet", "BillingCity"}, got)
}

func TestSelectStarRegex(t *testing.T) {
	cases := []struct {
		query   string
		wantObj string
		match   bool
	}{
		{"SELECT * FROM Account", "Account", true},
		{"select * from account", "account", true},
		{"  SELECT  *  FROM   Opportunity  ", "Opportunity", true},
		{"SELECT * FROM My_Custom__c WHERE Id != null", "My_Custom__c", true},
		{"SELECT * FROM Account\nORDER BY Name\nLIMIT 10", "Account", true},

		// Real explicit field lists must not be rewritten.
		{"SELECT Id, Name FROM Account", "", false},
		{"SELECT *, Id FROM Account", "", false},
		// FROM must be a bare identifier; subqueries / multiple objects don't
		// fit the simple shape we expand.
		{"SELECT * FROM (SELECT Id FROM Account)", "", false},
	}
	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			m := selectStarRegex.FindStringSubmatch(c.query)
			if !c.match {
				require.Len(t, m, 0, "expected no match")
				return
			}
			require.Greater(t, len(m), 1, "expected a match with capture group")
			require.Equal(t, c.wantObj, m[1])
		})
	}
}
