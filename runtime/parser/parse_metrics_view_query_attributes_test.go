package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateQueryAttributes(t *testing.T) {
	tests := []struct {
		name    string
		attrs   map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple attributes",
			attrs:   map[string]string{"partner_id": "acme_corp", "region": "us-west"},
			wantErr: false,
		},
		{
			name:    "valid with underscores and hyphens",
			attrs:   map[string]string{"partner_id": "value1", "user-role": "admin", "app.env": "prod"},
			wantErr: false,
		},
		{
			name:    "valid with dots in key",
			attrs:   map[string]string{"app.environment": "production"},
			wantErr: false,
		},
		{
			name:    "valid with template",
			attrs:   map[string]string{"partner_id": "{{ .user.partner_id }}"},
			wantErr: false,
		},
		{
			name:    "empty attributes map",
			attrs:   map[string]string{},
			wantErr: false,
		},
		{
			name:    "nil attributes map",
			attrs:   nil,
			wantErr: false,
		},
		{
			name:    "empty key",
			attrs:   map[string]string{"": "value"},
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name:    "invalid key with spaces",
			attrs:   map[string]string{"partner id": "value"},
			wantErr: true,
			errMsg:  "contains invalid characters",
		},
		{
			name:    "invalid key with special chars",
			attrs:   map[string]string{"partner@id": "value"},
			wantErr: true,
			errMsg:  "contains invalid characters",
		},
		{
			name:    "invalid key with SQL injection",
			attrs:   map[string]string{"partner'; DROP TABLE users--": "value"},
			wantErr: true,
			errMsg:  "contains invalid characters",
		},
		{
			name:    "template with dangerous pattern should pass",
			attrs:   map[string]string{"query": "{{ .user.custom_query }}"},
			wantErr: false,
		},
		{
			name:    "mixed safe and template values",
			attrs:   map[string]string{"env": "production", "partner_id": "{{ .user.partner_id }}"},
			wantErr: false,
		},
		{
			name:    "valid uppercase key",
			attrs:   map[string]string{"PARTNER_ID": "value"},
			wantErr: false,
		},
		{
			name:    "valid numeric in key",
			attrs:   map[string]string{"partner_id_123": "value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQueryAttributes(tt.attrs)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsValidQueryAttributeKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{"simple alphanumeric", "partner_id", true},
		{"with hyphen", "partner-id", true},
		{"with dot", "app.environment", true},
		{"with numbers", "key123", true},
		{"uppercase", "PARTNER_ID", true},
		{"mixed case", "PartnerId", true},
		{"empty string", "", false},
		{"with space", "partner id", false},
		{"with special char", "partner@id", false},
		{"with slash", "partner/id", false},
		{"with quotes", "partner'id", false},
		{"with semicolon", "partner;id", false},
		{"unicode", "партнер", false},
		{"emoji", "partner🎉", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidQueryAttributeKey(tt.key)
			require.Equal(t, tt.valid, result)
		})
	}
}

// generateManyAttributes creates a map with n attributes for testing limits
func generateManyAttributes(n int) map[string]string {
	attrs := make(map[string]string, n)
	for i := 0; i < n; i++ {
		attrs[strings.Repeat("a", i+1)] = "value"
	}
	return attrs
}
