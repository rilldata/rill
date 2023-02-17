package drivers

import (
	"context"
	"reflect"
	"testing"
)

func TestNewEnvVariables(t *testing.T) {
	tests := []struct {
		name      string
		yamlFile  string
		envString string
		want      EnviornmentVariables
		wantErr   bool
	}{
		{name: "no env", yamlFile: "", envString: "", want: map[string]map[string]string{"env": {}}, wantErr: false},
		{
			name:      "env",
			yamlFile:  "",
			envString: "timeout=100;region=us-east-1;format=csv,tsv;host=127.0.0.1:9090;limit=limit 100",
			want:      map[string]map[string]string{"env": {"timeout": "100", "region": "us-east-1", "format": "csv,tsv", "host": "127.0.0.1:9090", "limit": "limit 100"}},
			wantErr:   false,
		},
		{
			name: "env and yaml",
			yamlFile: `---
compiler: rill-beta
rill_version: next
name: ad-bids
env:
    timeout: 1000
    value: 1 3 4			
`,
			envString: "timeout=100;region=us-east-1;format=csv,tsv;host=127.0.0.1:9090;limit2=;limit=limit 100;",
			want:      map[string]map[string]string{"env": {"timeout": "100", "region": "us-east-1", "format": "csv,tsv", "host": "127.0.0.1:9090", "limit": "limit 100", "value": "1 3 4", "limit2": ""}},
			wantErr:   false,
		},
		{
			name: "only yaml",
			yamlFile: `---
compiler: rill-beta
rill_version: next
name: ad-bids
env:
    timeout: 1000
    value: 1 3 4			
`,
			envString: "",
			want:      map[string]map[string]string{"env": {"timeout": "1000", "value": "1 3 4"}},
			wantErr:   false,
		},
		{
			name: "invalid env",
			yamlFile: `---
compiler: rill-beta
rill_version: next
name: ad-bids
env:
    timeout: 1000
    value: 1 3 4			
`,
			envString: "timeout=100;value;",
			want:      nil,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEnvVariables(context.Background(), tt.yamlFile, tt.envString)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEnvVariables() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEnvVariables() = %v, want %v", got, tt.want)
			}
		})
	}
}
