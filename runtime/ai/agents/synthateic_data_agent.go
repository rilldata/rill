package agents

import (
	"context"
	"fmt"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/ai/tools"
)

// NewSyntheticDataAgent creates a new SyntheticDataAgent
func NewSyntheticDataAgent(ctx context.Context, instanceID, modelName string, r *runtime.Runtime) (*agent.Agent, error) {
	a := agent.NewAgent("SyntheticDataAgent")
	a.WithModel(modelName)

	olap, release, err := r.OLAP(ctx, instanceID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get OLAP connection: %w", err)
	}
	defer release()

	a.SetSystemInstructions(fmt.Sprintf(`You are an Agent that generates synthetic data for Rill projects based.
- You generate realistic business data using %s SQL.
- Create a rill model file with the generated SQL. The file extension should be .sql and saved in the models/ directory.
- Must ensure the SQL does not contain a trailing semicolon.
- The model name should be in snake_case and should be inferred from the context (e.g., "sales data" -> "sales_model").
- Use the "create_and_reconcile_resource" tool to create and validate the model in the Rill project.
- If the validation fails fix the SQL and try again.
CONSTRAINTS:
- ONLY generate SELECT queries that return data rows
- Generate 20-30 columns total
- 1000-10000 rows (choose based on context)
- Time range: 30-180 days from current_timestamp
- Do not ask for user input, generate SQL directly
REQUIRED COLUMNS:
- 1 timestamp column (MANDATORY): Use date arithmetic like: current_date - (random() * N)::int OR now() - interval (random() * N) day
- 5-8 dimensions: categorical data relevant to domain
- 8-12 measures: numeric metrics relevant to domain
- Supporting columns: IDs, names, flags, etc.
RESPONSE FORMAT:
Must include the model name in the response.
EXAMPLE:
SELECT 
    generate_series AS order_id,
    current_date - (random() * 90)::int AS order_date,
    'CUST_' || (generate_series %% 1000)::text AS customer_id,
    ['Electronics', 'Clothing', 'Books'][((random() * 3)::int + 1)] AS category,
    (random() * 500 + 10)::decimal(10,2) AS amount
FROM generate_series(1, 5000)`, olap.Dialect()))

	a.WithTools(tools.CreateAndReconcileResource(instanceID, r))
	return a, nil
}
