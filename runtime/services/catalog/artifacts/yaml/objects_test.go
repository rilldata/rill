package yaml

import (
	"reflect"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

func Test_fromExtractArtifact(t *testing.T) {
	tests := []struct {
		name    string
		input   *ExtractPolicy
		want    *runtimev1.Source_ExtractPolicy
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
			input: &ExtractPolicy{Row: &ExtractConfig{Strategy: "tail", Size: "23 KB"}},
			want: &runtimev1.Source_ExtractPolicy{
				RowsStrategy:   runtimev1.Source_ExtractPolicy_STRATEGY_TAIL,
				RowsLimitBytes: 23552,
			},
			wantErr: false,
		},
		{
			name:  "parse files",
			input: &ExtractPolicy{File: &ExtractConfig{Strategy: "head", Size: "23"}},
			want: &runtimev1.Source_ExtractPolicy{
				FilesStrategy: runtimev1.Source_ExtractPolicy_STRATEGY_HEAD,
				FilesLimit:    23,
			},
			wantErr: false,
		},
		{
			name:  "parse both",
			input: &ExtractPolicy{File: &ExtractConfig{Strategy: "tail", Size: "23"}, Row: &ExtractConfig{Strategy: "tail", Size: "512 B"}},
			want: &runtimev1.Source_ExtractPolicy{
				FilesStrategy:  runtimev1.Source_ExtractPolicy_STRATEGY_TAIL,
				FilesLimit:     23,
				RowsStrategy:   runtimev1.Source_ExtractPolicy_STRATEGY_TAIL,
				RowsLimitBytes: 512,
			},
			wantErr: false,
		},
		{
			name:  "more examples",
			input: &ExtractPolicy{File: &ExtractConfig{Strategy: "tail", Size: "23"}, Row: &ExtractConfig{Strategy: "tail", Size: "23 gb"}},
			want: &runtimev1.Source_ExtractPolicy{
				FilesStrategy:  runtimev1.Source_ExtractPolicy_STRATEGY_TAIL,
				FilesLimit:     23,
				RowsStrategy:   runtimev1.Source_ExtractPolicy_STRATEGY_TAIL,
				RowsLimitBytes: 23 * 1024 * 1024 * 1024,
			},
			wantErr: false,
		},
		{
			name:    "invalid",
			input:   &ExtractPolicy{File: &ExtractConfig{Strategy: "tail", Size: "23"}, Row: &ExtractConfig{Strategy: "tail", Size: "23%"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fromExtractArtifact(tt.input)
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
