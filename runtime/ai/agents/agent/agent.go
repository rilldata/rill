package agent

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb/v2"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// SyntheticDataAgent is an AI agent that generates synthetic data using DuckDB SQL
type SyntheticDataAgent struct {
	*agent.Agent
	runner *runner.Runner
}

// NewSyntheticDataAgent creates a new SyntheticDataAgent
func NewSyntheticDataAgent(modelName string, runner *runner.Runner) *SyntheticDataAgent {
	a := agent.NewAgent("SyntheticDataAgent")
	sda := &SyntheticDataAgent{
		Agent:  a,
		runner: runner,
	}
	sda.WithModel(modelName)
	sda.configure()

	return sda
}

func (s *SyntheticDataAgent) configure() {
	s.SetSystemInstructions(`You generate realistic business data using DuckDB SQL. When given error feedback, fix the SQL and return only the corrected SQL query.

CONSTRAINTS:
- ONLY generate SELECT queries that return data rows
- Generate 20-30 columns total
- 1000-10000 rows (choose based on context)
- Time range: 30-180 days from current_timestamp

REQUIRED COLUMNS:
- 1 timestamp column (MANDATORY): Use date arithmetic like: current_date - (random() * N)::int OR now() - interval (random() * N) day
- 5-8 dimensions: categorical data relevant to domain
- 8-12 measures: numeric metrics relevant to domain
- Supporting columns: IDs, names, flags, etc.

DOMAIN INFERENCE:
- E-commerce: orders, customers, products, sales → order_id, customer_id, product_name, price, quantity
- Finance: transactions, accounts, payments → transaction_id, account_number, amount, transaction_type  
- Marketing: campaigns, leads, conversions → campaign_id, lead_source, conversion_rate, cost_per_click
- SaaS: users, subscriptions, usage → user_id, subscription_tier, feature_usage, churn_risk
- Healthcare: patients, treatments → patient_id, treatment_type, diagnosis, cost
- Manufacturing: production, quality → batch_id, product_type, quality_score, defect_rate

DUCKDB SYNTAX RULES:
- Use generate_series: FROM generate_series(1, N)
- Categorical values: ['Val1', 'Val2', 'Val3'][((random() * 3)::int + 1)]
- Random numbers: (random() * 100)::int, (random() * 1000)::decimal(10,2)
- Modulo for patterns: generate_series % 5
- Boolean flags: random() < 0.3
- Cast to text: ::text
- Date arithmetic: current_date - (random() * 90)::int OR now() - interval (random() * 90) day
- NEVER use current_timestamp with interval subtraction - use current_date or now() instead

ERROR FIXING:
When you receive an error, analyze it carefully and fix:
- Syntax errors: Check DuckDB-specific syntax
- Type errors: Add proper casting (::int, ::text, ::decimal)
- Array bounds: Ensure array indices are valid
- Function calls: Verify function names and parameters
- Date/time errors: Use current_date - N, now() - interval N day, or date_add(current_date, -N)
- Interval errors: Replace current_timestamp - interval with current_date - N or now() - interval N day

RESPONSE FORMAT:
Return ONLY the SQL query, no explanations or comments unless fixing an error.

EXAMPLE:
SELECT 
    generate_series AS order_id,
    current_date - (random() * 90)::int AS order_date,
    'CUST_' || (generate_series % 1000)::text AS customer_id,
    ['Electronics', 'Clothing', 'Books'][((random() * 3)::int + 1)] AS category,
    (random() * 500 + 10)::decimal(10,2) AS amount
FROM generate_series(1, 5000)`)

	s.WithTools(
		s.createValidateSQLTool(),
	)
}


func (s *SyntheticDataAgent) createValidateSQLTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sql": map[string]interface{}{
				"type":        "string",
				"description": "The SQL to validate",
			},
		},
		"required": []string{"sql"},
	}

	t := tool.NewFunctionTool(
		"validate_duckdb_sql",
		"Validates DuckDB SQL syntax and provides feedback on potential issues",
		s.validateSQL,
	)

	// Set the schema
	t.WithSchema(schema)

	return t
}

// sanitizeSQL strips markdown code fences and trims whitespace.
func sanitizeSQL(sql string) string {
    s := strings.TrimSpace(sql)
    if strings.HasPrefix(s, "```") {
        newline := strings.Index(s, "\n")
        if newline != -1 {
            rest := s[newline+1:]
            end := strings.Index(rest, "```")
            if end != -1 {
                s = rest[:end]
            } else {
                s = rest
            }
        }
    }
    return strings.TrimSpace(s)
}

func (s *SyntheticDataAgent) GenerateSQL(description string) (string, error) {
	if description == "" {
		return "", fmt.Errorf("description parameter is required")
	}

	ctx := context.Background()
	maxAttempts := 5
	
	log.Printf("🚀 Starting SQL generation for: %s", description)
	
	// Initial prompt to generate SQL
	prompt := fmt.Sprintf("Generate synthetic data SQL for: %s\n\nRequirements:\n1. Output ONLY raw SQL.\n2. Do NOT wrap with markdown code fences or add any explanatory text.", description)
	
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("🔄 Attempt %d/%d", attempt, maxAttempts)
		log.Printf("📝 Prompt: %s", prompt)
		
		// Call the AI agent using runner
		runResult, err := s.runner.Run(ctx, s.Agent, &runner.RunOptions{
			Input:    prompt,
			MaxTurns: 3,
		})
		if err != nil {
			log.Printf("❌ Agent execution failed on attempt %d: %v", attempt, err)
			if attempt == maxAttempts {
				return "", fmt.Errorf("failed to generate SQL after %d attempts. Last error: %w", maxAttempts, err)
			}
			continue
		}
		
		// Extract SQL from response
		var sql string
		if runResult != nil && runResult.FinalOutput != nil {
			if str, ok := runResult.FinalOutput.(string); ok {
				sql = sanitizeSQL(str)
				log.Printf("🤖 AI Generated SQL:\n%s", sql)
			} else {
				log.Printf("❌ Non-string response from agent on attempt %d: %T", attempt, runResult.FinalOutput)
				if attempt == maxAttempts {
					return "", fmt.Errorf("failed to get string response from agent after %d attempts", maxAttempts)
				}
				continue
			}
		} else {
			log.Printf("❌ No response from agent on attempt %d", attempt)
			if attempt == maxAttempts {
				return "", fmt.Errorf("no response from agent after %d attempts", maxAttempts)
			}
			continue
		}
		
		// Validate the generated SQL
		log.Printf("⚖️ Validating generated SQL...")
		valid, validationErr := s.validateSQLWithDuckDB(ctx, sql)
		if validationErr != nil {
			log.Printf("❌ Validation failed on attempt %d: %v", attempt, validationErr)
			if attempt == maxAttempts {
				return "", fmt.Errorf("failed to generate valid SQL after %d attempts. Last error: %w", maxAttempts, validationErr)
			}
			// Update prompt with error feedback for next attempt
			prompt = fmt.Sprintf("The previous SQL had an error. Please fix the SQL below. Return ONLY raw SQL (no fences, no explanation):\n\n%s\n\nError: %s\n\nGenerate corrected SQL for: %s", sql, validationErr.Error(), description)
			log.Printf("🔧 Updated prompt with error feedback for next attempt")
			continue
		}
		
		if valid {
			log.Printf("✅ SQL generation successful on attempt %d", attempt)
			return sql, nil
		}
	}
	
	log.Printf("❌ Failed to generate valid SQL after %d attempts", maxAttempts)
	return "", fmt.Errorf("failed to generate valid SQL after %d attempts", maxAttempts)
}



	func (s *SyntheticDataAgent) validateSQL(ctx context.Context, params map[string]interface{}) (interface{}, error) {
    sqlQuery, ok := params["sql"].(string)
    if !ok {
        return nil, fmt.Errorf("sql parameter is required")
    }

    // Sanitize in case markdown fences slipped in
    sqlQuery = sanitizeSQL(sqlQuery)

    valid, err := s.validateSQLWithDuckDB(ctx, sqlQuery)
    if err != nil || !valid {
        msg := "SQL validation failed"
        if err != nil {
            msg = err.Error()
        }
        return map[string]interface{}{
            "valid":   false,
            "message": msg,
            "details": map[string]interface{}{
                "sql":   sqlQuery,
                "error": msg,
            },
        }, nil
    }

    return map[string]interface{}{
        "valid":   true,
        "message": "SQL is valid and executable",
        "details": map[string]interface{}{
            "sql": sqlQuery,
        },
    }, nil
}

// validateSQLWithDuckDB validates SQL by creating a temporary view in DuckDB
func (s *SyntheticDataAgent) validateSQLWithDuckDB(ctx context.Context, sqlQuery string) (bool, error) {
	log.Printf("🔍 Validating SQL:\n%s\n", sqlQuery)
	
	// Create a temporary in-memory DuckDB connection
	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		log.Printf("❌ Failed to create DuckDB connector: %v", err)
		return false, fmt.Errorf("failed to create DuckDB connector: %w", err)
	}

	db := sqlx.NewDb(sql.OpenDB(connector), "duckdb")
	defer db.Close()

	// Test the SQL by creating a temporary view
	viewName := fmt.Sprintf("temp_view_%s", strings.ReplaceAll(uuid.NewString(), "-", ""))
	createViewSQL := fmt.Sprintf("CREATE TEMPORARY VIEW %s AS %s", viewName, sqlQuery)
	
	log.Printf("🧪 Testing SQL with view creation: %s", createViewSQL)
	
	_, err = db.ExecContext(ctx, createViewSQL)
	if err != nil {
		log.Printf("❌ SQL validation failed: %v", err)
		return false, fmt.Errorf("SQL validation failed: %w", err)
	}

	log.Printf("✅ SQL validation passed - view created successfully")

	// Clean up the temporary view
	dropViewSQL := fmt.Sprintf("DROP VIEW %s", viewName)
	_, err = db.ExecContext(ctx, dropViewSQL)
	if err != nil {
		log.Printf("⚠️ Failed to cleanup view %s: %v (non-critical)", viewName, err)
		// Log the error but don't fail validation since the main SQL is valid
		// The cleanup failure is not critical for validation
	} else {
		log.Printf("🧹 Successfully cleaned up view %s", viewName)
	}

	return true, nil
}

// ValidateSQL exposes the validation function for external use
func (s *SyntheticDataAgent) ValidateSQL(sqlQuery string) (bool, error) {
	return s.validateSQLWithDuckDB(context.Background(), sqlQuery)
}
