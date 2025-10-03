package ai

import (
	"bytes"
	"strings"
	"text/template"
)

var templateFuncs = template.FuncMap{
	"backticks": func() string {
		return "```"
	},
}

func executeTemplate(templ string, data map[string]any) (string, error) {
	tmpl, err := template.New("").Funcs(templateFuncs).Parse(templ)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

func mustExecuteTemplate(templ string, data map[string]any) string {
	result, err := executeTemplate(templ, data)
	if err != nil {
		panic(err)
	}
	return result
}
