package fieldselectorpb

import (
	"slices"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name     string
		selector *runtimev1.FieldSelector
		all      []string
		want     []string
		wantErr  bool
	}{
		{
			name:     "nil selector",
			selector: nil,
			all:      []string{"a", "b", "c"},
			wantErr:  true,
		},
		{
			name:     "all selector",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_All{All: true}},
			all:      []string{"a", "b", "c"},
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "empty selector",
			selector: &runtimev1.FieldSelector{},
			all:      []string{"a", "b", "c"},
			wantErr:  true,
		},
		{
			name:     "fields selector",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: []string{"a", "b"}}}},
			all:      []string{"a", "b", "c"},
			want:     []string{"a", "b"},
		},
		{
			name:     "fields selector not found",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: []string{"a", "d"}}}},
			all:      []string{"a", "b", "c"},
			wantErr:  true,
		},
		{
			name:     "fields selector invert",
			selector: &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: []string{"a", "b"}}}},
			all:      []string{"a", "b", "c"},
			want:     []string{"c"},
		},
		{
			name:     "fields selector invert all",
			selector: &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_Fields{Fields: &runtimev1.StringListValue{Values: []string{"a", "b", "c"}}}},
			all:      []string{"a", "b", "c"},
			want:     nil,
		},
		{
			name:     "regex selector",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_Regex{Regex: "a|b"}},
			all:      []string{"a", "b", "c"},
			want:     []string{"a", "b"},
		},
		{
			name:     "regex selector invert",
			selector: &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_Regex{Regex: "a|b"}},
			all:      []string{"a", "b", "c"},
			want:     []string{"c"},
		},
		{
			name:     "regex selector invert all",
			selector: &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_Regex{Regex: ".*"}},
			all:      []string{"a", "b", "c"},
			want:     nil,
		},
		{
			name:     "invalid regex",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_Regex{Regex: "a|b|("}},
			all:      []string{"a", "b", "c"},
			wantErr:  true,
		},
		{
			name:     "duckdb selector",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_DuckdbExpression{DuckdbExpression: "a, b"}},
			all:      []string{"a", "b", "c"},
			want:     []string{"a", "b"},
		},
		{
			name:     "duckdb selector invert",
			selector: &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_DuckdbExpression{DuckdbExpression: "a, b"}},
			all:      []string{"a", "b", "c"},
			want:     []string{"c"},
		},
		{
			name:     "duckdb selector not found",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_DuckdbExpression{DuckdbExpression: "a, d"}},
			all:      []string{"a", "b", "c"},
			wantErr:  true,
		},
		{
			name:     "duckdb columns regex",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_DuckdbExpression{DuckdbExpression: "COLUMNS('a|b')"}},
			all:      []string{"a", "b", "c"},
			want:     []string{"a", "b"},
		},
		{
			name:     "duckdb columns regex invert",
			selector: &runtimev1.FieldSelector{Invert: true, Selector: &runtimev1.FieldSelector_DuckdbExpression{DuckdbExpression: "COLUMNS('a|b')"}},
			all:      []string{"a", "b", "c"},
			want:     []string{"c"},
		},
		{
			name:     "duckdb columns wildcard except",
			selector: &runtimev1.FieldSelector{Selector: &runtimev1.FieldSelector_DuckdbExpression{DuckdbExpression: "* EXCLUDE (c)"}},
			all:      []string{"a", "b", "c"},
			want:     []string{"a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resolve(tt.selector, tt.all)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("Resolve() got = %v, want %v", got, tt.want)
			}
		})
	}
}
