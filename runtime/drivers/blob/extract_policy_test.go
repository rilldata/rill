package blob

import (
	"reflect"
	"testing"
)

func Test_fromExtractArtifact(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		want    *ExtractPolicy
		wantErr bool
	}{
		{
			name:    "nil input",
			input:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name:  "parse row",
			input: map[string]any{"rows": map[string]any{"strategy": "tail", "size": "23 KB"}},
			want: &ExtractPolicy{
				RowsStrategy:   ExtractPolicyStrategyTail,
				RowsLimitBytes: 23552,
			},
			wantErr: false,
		},
		{
			name:  "parse files",
			input: map[string]any{"files": map[string]any{"strategy": "head", "size": "23"}},
			want: &ExtractPolicy{
				FilesStrategy: ExtractPolicyStrategyHead,
				FilesLimit:    23,
			},
			wantErr: false,
		},
		{
			name:  "parse both",
			input: map[string]any{"files": map[string]any{"strategy": "tail", "size": "23"}, "rows": map[string]any{"strategy": "tail", "size": "512 B"}},
			want: &ExtractPolicy{
				FilesStrategy:  ExtractPolicyStrategyTail,
				FilesLimit:     23,
				RowsStrategy:   ExtractPolicyStrategyTail,
				RowsLimitBytes: 512,
			},
			wantErr: false,
		},
		{
			name:  "more examples",
			input: map[string]any{"files": map[string]any{"strategy": "tail", "size": "23"}, "rows": map[string]any{"strategy": "tail", "size": "23 gb"}},
			want: &ExtractPolicy{
				FilesStrategy:  ExtractPolicyStrategyTail,
				FilesLimit:     23,
				RowsStrategy:   ExtractPolicyStrategyTail,
				RowsLimitBytes: 23 * 1024 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   map[string]any{"files": map[string]any{"strategy": "tail", "size": "23"}, "rows": map[string]any{"strategy": "tail", "size": "23%"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseExtractPolicy(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("fromExtractArtifact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fromExtractArtifact() = %v, want %v", got, tt.want)
			}
		})
	}
}
