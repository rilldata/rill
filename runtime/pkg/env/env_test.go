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

func TestParseKeyVal(t *testing.T) {
	tests := []struct {
		arg     string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{
			arg:     "key=value",
			wantKey: "key",
			wantVal: "value",
			wantErr: false,
		},
		{
			arg:     "ENV_VAR=SOME_VALUE",
			wantKey: "ENV_VAR",
			wantVal: "SOME_VALUE",
			wantErr: false,
		},
		{
			arg:     "ENV_VAR",
			wantKey: "",
			wantVal: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if gotKey, gotVal, err := ParseKeyVal(tt.arg); gotKey != tt.wantKey || gotVal != tt.wantVal || (err == nil) == tt.wantErr {
			t.Errorf("For env = %v, got = %v, %v, %v, want %v, %v, %v", tt.arg, gotKey, gotVal, err, tt.wantKey, tt.wantVal, tt.wantErr)
		}
	}
}

func TestParseAndValidate(t *testing.T) {
	tests := []struct {
		arg     string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{
			arg:     "key=value",
			wantKey: "key",
			wantVal: "value",
			wantErr: false,
		},
		{
			arg:     "ENV_VAR=SOME_VALUE",
			wantKey: "ENV_VAR",
			wantVal: "SOME_VALUE",
			wantErr: false,
		},

		{
			arg:     "ENV_VAR",
			wantKey: "",
			wantVal: "",
			wantErr: true,
		},
		{
			arg:     "1ENV_VAR=SOME_VALUE",
			wantKey: "",
			wantVal: "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		if gotKey, gotVal, err := ParseAndValidate(tt.arg); gotKey != tt.wantKey || gotVal != tt.wantVal || (err == nil) == tt.wantErr {
			t.Errorf("For env = %v, got = %v, %v, %v, want %v, %v, %v", tt.arg, gotKey, gotVal, err, tt.wantKey, tt.wantVal, tt.wantErr)
		}
	}
}
