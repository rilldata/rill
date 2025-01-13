package env

import "testing"

func TestIsValidName(t *testing.T) {
	tests := []struct {
		arg     string
		wantErr bool
	}{
		{
			arg:     "connector.duckdb.admin1",
			wantErr: false,
		},
		{
			arg:     "connector.duckdb.admin_1",
			wantErr: false,
		},
		{
			arg:     "1connector.duckdb.admin_1",
			wantErr: true,
		},
		{
			arg:     "connector.duckdb.admin-1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if got := ValidateName(tt.arg); (got == nil) == tt.wantErr {
			t.Errorf("For env = %v, got = %v, want %v", tt.arg, got, tt.wantErr)
		}
	}
}
