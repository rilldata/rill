package rillv1beta

import "testing"

func TestProjectConfig_SanitizedName(t *testing.T) {
	tests := []struct {
		testName string
		Name     string
		want     string
	}{
		{
			testName: "normal",
			Name:     "rill-devel0per",
			want:     "rill-devel0per",
		},
		{
			testName: "extra spaces",
			Name:     "   rill  devel0per  ",
			want:     "rill-devel0per",
		},
		{
			testName: "non aplphanumeric",
			Name:     "r:ll  :  1     devel💣💣::per",
			want:     "r-ll-1-devel-per",
		},
		{
			testName: "totally invalid",
			Name:     "💣:💣:💣",
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			p := &ProjectConfig{
				Name: tt.Name,
			}
			if got := p.SanitizedName(); got != tt.want {
				t.Errorf("ProjectConfig.SanitizedName() = %v, want %v", got, tt.want)
			}
		})
	}
}
