package salesforce

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDescribeSObject(t *testing.T) {
	// Trimmed shape of a real describe response: fields[].name and
	// fields[].type are the bits we use; extra keys must be ignored.
	body := `{
		"name": "Opportunity",
		"queryable": true,
		"fields": [
			{"name": "Id", "type": "id", "label": "Opportunity ID"},
			{"name": "Name", "type": "string", "length": 120},
			{"name": "Amount", "type": "currency"},
			{"name": "", "type": "string"}
		]
	}`

	schema, err := parseDescribeSObject(body)
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"Id":     "id",
		"Name":   "string",
		"Amount": "currency",
	}, schema)
}

func TestParseDescribeSObject_invalidJSON(t *testing.T) {
	_, err := parseDescribeSObject("{not json")
	require.Error(t, err)
}
