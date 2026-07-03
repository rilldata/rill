package urlutils

import "testing"

func TestSlugify(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"Overview", "overview"},
		{"  Trimmed  ", "trimmed"},
		{"Sales & Marketing", "sales-marketing"},
		{"Q1 2024 Report", "q1-2024-report"},
		{"Multiple   spaces", "multiple-spaces"},
		{"--leading/trailing--", "leading-trailing"},
		{"!!!", ""},
	}
	for _, c := range cases {
		if got := Slugify(c.in); got != c.want {
			t.Errorf("Slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
