package templates

import (
	"fmt"
	"strings"
	"text/template"
)

// funcMap returns the template function map available to all template definitions.
func funcMap() template.FuncMap {
	return template.FuncMap{
		"renderProps": renderProps,
		"indent":      indent,
		"quote":       quote,
	}
}

// renderProps renders a slice of ProcessedProp as YAML key-value lines.
// Each property is rendered on its own line with appropriate formatting:
// quoted values get double quotes, unquoted values are rendered as-is.
func renderProps(props []ProcessedProp) string {
	if len(props) == 0 {
		return ""
	}
	var b strings.Builder
	for i, p := range props {
		if i > 0 {
			b.WriteByte('\n')
		}
		if p.Quoted {
			fmt.Fprintf(&b, "%s: %q", p.Key, p.Value)
		} else {
			fmt.Fprintf(&b, "%s: %s", p.Key, p.Value)
		}
	}
	return b.String()
}

// indent prepends each line of text with n spaces.
func indent(n int, text string) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = pad + line
		}
	}
	return strings.Join(lines, "\n")
}

// quote wraps a string in double quotes.
func quote(s string) string {
	return fmt.Sprintf("%q", s)
}
