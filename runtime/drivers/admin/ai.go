package admin

import (
	"context"
	"fmt"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

const metricsViewYAMLSystemPrompt = `
You are an agent whose only task is to suggest relevant business KPI metrics based on a table schema in a valid YAML file and output it in a format below:

- name: name: <A unique name for the metric like average_sales etc..>
- expression: <SQL expression to calculate the KPI in the dialect user asks for>
- label: <A short descriptive label of the metric>
- description: <Short Description of the metric>
- valid_percent_of_total: <true if the metrics is summable otherwise false>
`

func (h *Handle) GenerateMetricsViewYAML(ctx context.Context, baseTable, sqlDialect string, schema *runtimev1.StructType) (string, error) {
	prompt := fmt.Sprintf(`
Give me 10 suggested metrics in a YAML file.
Use %q SQL dialect for the metric expressions.
Use only COUNT, SUM, MIN, MAX, AVG as aggregation functions.
Do not use any complex aggregations.
Do not use WHERE or FILTER in metrics definition.
The table name is %q.
Here is my table schema:
`, sqlDialect, baseTable)
	for _, field := range schema.Fields {
		prompt += fmt.Sprintf("- column=%s, type=%s\n", field.Name, field.Type.Code.String())
	}

	msgs := []*adminv1.CompletionMessage{
		{Role: "system", Data: metricsViewYAMLSystemPrompt},
		{Role: "user", Data: prompt},
	}

	res, err := h.admin.Complete(ctx, &adminv1.CompleteRequest{Messages: msgs})
	if err != nil {
		return "", err
	}

	return res.Message.Data, nil
}
