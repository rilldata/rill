package agent

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/marcboeker/go-duckdb/v2"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// ModelValidatorAgent is a specialized AI agent that validates SQL models for correctness and quality
type ModelValidatorAgent struct {
	*agent.Agent
	runner    *runner.Runner
	modelsDir string
}

// ValidationResult contains the results of a model validation
type ValidationResult struct {
	ModelName         string                 `json:"model_name"`
	Valid             bool                   `json:"valid"`
	SyntaxValid       bool                   `json:"syntax_valid"`
	ExecutionValid    bool                   `json:"execution_valid"`
	QualityScore      int                    `json:"quality_score"`
	Issues            []ValidationIssue      `json:"issues"`
	Warnings          []ValidationWarning    `json:"warnings"`
	Suggestions       []ValidationSuggestion `json:"suggestions"`
	ValidationTime    time.Time              `json:"validation_time"`
	ExecutionTime     time.Duration          `json:"execution_time"`
	RowCount          *int64                 `json:"row_count,omitempty"`
	ColumnCount       *int                   `json:"column_count,omitempty"`
	EstimatedSize     *int64                 `json:"estimated_size,omitempty"`
	ValidationPassed  bool                   `json:"validation_passed"`
	SQL               string                 `json:"sql"`
}

// ValidationIssue represents a validation error that prevents the model from being saved
type ValidationIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Message     string `json:"message"`
	Line        *int   `json:"line,omitempty"`
	Column      *int   `json:"column,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// ValidationWarning represents a potential issue that doesn't prevent saving but should be addressed
type ValidationWarning struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// ValidationSuggestion represents an improvement recommendation
type ValidationSuggestion struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	Impact      string `json:"impact"`
	Priority    string `json:"priority"`
}

// NewModelValidatorAgent creates a new ModelValidatorAgent
func NewModelValidatorAgent(modelName string, runner *runner.Runner, modelsDir string) *ModelValidatorAgent {
	a := agent.NewAgent("ModelValidatorAgent")
	mva := &ModelValidatorAgent{
		Agent:     a,
		runner:    runner,
		modelsDir: modelsDir,
	}
	
	mva.WithModel(modelName)
	mva.configure()

	return mva
}

func (m *ModelValidatorAgent) configure() {
	m.SetSystemInstructions(`You are a ModelValidator Agent specialized in validating SQL models for correctness, performance, and quality.

Your primary responsibilities:
1. Validate SQL syntax and semantic correctness
2. Check for performance issues and optimization opportunities
3. Ensure adherence to best practices and coding standards
4. Identify potential data quality issues
5. Provide actionable feedback and suggestions for improvement

VALIDATION CATEGORIES:

1. SYNTAX VALIDATION:
   - SQL syntax correctness
   - Proper table and column references
   - Correct function usage
   - Valid data types and casting

2. EXECUTION VALIDATION:
   - Query can be executed without errors
   - All referenced tables/columns exist (when possible)
   - No infinite loops or resource exhaustion
   - Reasonable execution time

3. QUALITY VALIDATION:
   - Consistent naming conventions (snake_case)
   - Proper indentation and formatting
   - Meaningful column aliases
   - Appropriate use of comments
   - No hardcoded values where parameters should be used

4. PERFORMANCE VALIDATION:
   - Efficient join strategies
   - Proper use of WHERE clauses
   - Avoid SELECT * when specific columns are needed
   - Reasonable data volume expectations
   - Index-friendly query patterns

5. BUSINESS LOGIC VALIDATION:
   - Logical consistency in calculations
   - Proper handling of NULL values
   - Appropriate data type conversions
   - Sensible default values

VALIDATION LEVELS:
- ERROR: Critical issues that prevent execution or cause incorrect results
- WARNING: Issues that work but are suboptimal or risky
- SUGGESTION: Improvements that enhance maintainability or performance

OUTPUT REQUIREMENTS:
- Provide specific, actionable feedback
- Include line numbers when possible
- Suggest concrete improvements
- Explain the reasoning behind recommendations
- Prioritize issues by severity and impact`)

	m.WithTools(
		m.createValidateSQLTool(),
		m.createCheckSyntaxTool(),
		m.createAnalyzePerformanceTool(),
		m.createCheckQualityTool(),
	)
}

func (m *ModelValidatorAgent) createValidateSQLTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sql": map[string]interface{}{
				"type":        "string",
				"description": "The SQL content to validate",
			},
			"model_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the model being validated",
			},
		},
		"required": []string{"sql"},
	}

	t := tool.NewFunctionTool(
		"validate_sql_comprehensive",
		"Performs comprehensive SQL validation including syntax, execution, and quality checks",
		m.validateSQLComprehensive,
	)

	t.WithSchema(schema)
	return t
}

func (m *ModelValidatorAgent) createCheckSyntaxTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sql": map[string]interface{}{
				"type":        "string",
				"description": "The SQL content to check for syntax errors",
			},
		},
		"required": []string{"sql"},
	}

	t := tool.NewFunctionTool(
		"check_sql_syntax",
		"Checks SQL syntax for errors and basic structural issues",
		m.checkSQLSyntax,
	)

	t.WithSchema(schema)
	return t
}

func (m *ModelValidatorAgent) createAnalyzePerformanceTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sql": map[string]interface{}{
				"type":        "string",
				"description": "The SQL content to analyze for performance issues",
			},
		},
		"required": []string{"sql"},
	}

	t := tool.NewFunctionTool(
		"analyze_sql_performance",
		"Analyzes SQL for potential performance issues and optimization opportunities",
		m.analyzeSQLPerformance,
	)

	t.WithSchema(schema)
	return t
}

func (m *ModelValidatorAgent) createCheckQualityTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sql": map[string]interface{}{
				"type":        "string",
				"description": "The SQL content to check for quality and best practices",
			},
		},
		"required": []string{"sql"},
	}

	t := tool.NewFunctionTool(
		"check_sql_quality",
		"Checks SQL for adherence to best practices and coding standards",
		m.checkSQLQuality,
	)

	t.WithSchema(schema)
	return t
}

func (m *ModelValidatorAgent) validateSQLComprehensive(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	sqlContent, ok := params["sql"].(string)
	if !ok {
		return nil, fmt.Errorf("sql parameter is required")
	}

	modelName, _ := params["model_name"].(string)

	startTime := time.Now()
	result := &ValidationResult{
		ModelName:        modelName,
		ValidationTime:   startTime,
		Issues:          []ValidationIssue{},
		Warnings:        []ValidationWarning{},
		Suggestions:     []ValidationSuggestion{},
		SQL:             sqlContent,
	}

	// 1. Syntax validation
	log.Printf("🔍 Validating SQL syntax for model: %s", modelName)
	syntaxValid, syntaxIssues := m.validateSyntaxWithDuckDB(ctx, sqlContent)
	result.SyntaxValid = syntaxValid
	result.Issues = append(result.Issues, syntaxIssues...)

	// 2. Execution validation (if syntax is valid)
	if syntaxValid {
		log.Printf("✅ Syntax validation passed, checking execution...")
		executionValid, executionIssues, stats := m.validateExecutionWithDuckDB(ctx, sqlContent)
		result.ExecutionValid = executionValid
		result.Issues = append(result.Issues, executionIssues...)
		
		if stats != nil {
			result.RowCount = stats.RowCount
			result.ColumnCount = stats.ColumnCount
			result.EstimatedSize = stats.EstimatedSize
		}
	} else {
		log.Printf("❌ Syntax validation failed, skipping execution validation")
		result.ExecutionValid = false
	}

	// 3. Quality validation
	log.Printf("🎯 Checking SQL quality and best practices...")
	qualityScore, qualityWarnings, qualitySuggestions := m.validateQuality(sqlContent)
	result.QualityScore = qualityScore
	result.Warnings = append(result.Warnings, qualityWarnings...)
	result.Suggestions = append(result.Suggestions, qualitySuggestions...)

	// 4. Performance analysis
	log.Printf("⚡ Analyzing performance characteristics...")
	perfWarnings, perfSuggestions := m.analyzePerformance(sqlContent)
	result.Warnings = append(result.Warnings, perfWarnings...)
	result.Suggestions = append(result.Suggestions, perfSuggestions...)

	// Calculate overall validation status
	result.Valid = result.SyntaxValid && result.ExecutionValid
	result.ValidationPassed = result.Valid && len(result.Issues) == 0
	result.ExecutionTime = time.Since(startTime)

	log.Printf("📊 Validation complete - Valid: %t, Quality Score: %d, Issues: %d, Warnings: %d", 
		result.Valid, result.QualityScore, len(result.Issues), len(result.Warnings))

	return result, nil
}

func (m *ModelValidatorAgent) checkSQLSyntax(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	sqlContent, ok := params["sql"].(string)
	if !ok {
		return nil, fmt.Errorf("sql parameter is required")
	}

	valid, issues := m.validateSyntaxWithDuckDB(ctx, sqlContent)
	
	return map[string]interface{}{
		"syntax_valid": valid,
		"issues":       issues,
	}, nil
}

func (m *ModelValidatorAgent) analyzeSQLPerformance(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	sqlContent, ok := params["sql"].(string)
	if !ok {
		return nil, fmt.Errorf("sql parameter is required")
	}

	warnings, suggestions := m.analyzePerformance(sqlContent)
	
	return map[string]interface{}{
		"warnings":    warnings,
		"suggestions": suggestions,
	}, nil
}

func (m *ModelValidatorAgent) checkSQLQuality(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	sqlContent, ok := params["sql"].(string)
	if !ok {
		return nil, fmt.Errorf("sql parameter is required")
	}

	score, warnings, suggestions := m.validateQuality(sqlContent)
	
	return map[string]interface{}{
		"quality_score": score,
		"warnings":      warnings,
		"suggestions":   suggestions,
	}, nil
}

// ExecutionStats contains statistics from SQL execution
type ExecutionStats struct {
	RowCount      *int64
	ColumnCount   *int
	EstimatedSize *int64
}

func (m *ModelValidatorAgent) validateSyntaxWithDuckDB(ctx context.Context, sqlContent string) (bool, []ValidationIssue) {
	var issues []ValidationIssue
	
	// Create a temporary in-memory DuckDB connection
	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		issues = append(issues, ValidationIssue{
			Type:     "connection_error",
			Severity: "error",
			Message:  fmt.Sprintf("Failed to create DuckDB connector: %v", err),
		})
		return false, issues
	}

	db := sqlx.NewDb(sql.OpenDB(connector), "duckdb")
	defer db.Close()

	// Test the SQL by creating a temporary view
	viewName := fmt.Sprintf("temp_validation_view_%s", strings.ReplaceAll(uuid.NewString(), "-", ""))
	createViewSQL := fmt.Sprintf("CREATE TEMPORARY VIEW %s AS %s", viewName, sqlContent)
	
	_, err = db.ExecContext(ctx, createViewSQL)
	if err != nil {
		issues = append(issues, ValidationIssue{
			Type:       "syntax_error",
			Severity:   "error",
			Message:    fmt.Sprintf("SQL syntax validation failed: %v", err),
			Suggestion: "Check for missing commas, incorrect function names, or invalid syntax",
		})
		return false, issues
	}

	// Clean up the temporary view
	dropViewSQL := fmt.Sprintf("DROP VIEW %s", viewName)
	_, _ = db.ExecContext(ctx, dropViewSQL)

	return true, issues
}

func (m *ModelValidatorAgent) validateExecutionWithDuckDB(ctx context.Context, sqlContent string) (bool, []ValidationIssue, *ExecutionStats) {
	var issues []ValidationIssue
	var stats *ExecutionStats
	
	// Create a temporary in-memory DuckDB connection
	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		issues = append(issues, ValidationIssue{
			Type:     "connection_error",
			Severity: "error",
			Message:  fmt.Sprintf("Failed to create DuckDB connector: %v", err),
		})
		return false, issues, nil
	}

	db := sqlx.NewDb(sql.OpenDB(connector), "duckdb")
	defer db.Close()

	// Execute the SQL with a LIMIT to prevent resource exhaustion
	limitedSQL := fmt.Sprintf("SELECT * FROM (%s) LIMIT 100", sqlContent)
	
	rows, err := db.QueryContext(ctx, limitedSQL)
	if err != nil {
		issues = append(issues, ValidationIssue{
			Type:       "execution_error",
			Severity:   "error",
			Message:    fmt.Sprintf("SQL execution failed: %v", err),
			Suggestion: "Check table/column references and function usage",
		})
		return false, issues, nil
	}
	defer rows.Close()

	// Get column information
	columns, err := rows.Columns()
	if err == nil {
		stats = &ExecutionStats{
			ColumnCount: new(int),
		}
		*stats.ColumnCount = len(columns)
	}

	// Count rows (limited sample)
	rowCount := int64(0)
	for rows.Next() {
		rowCount++
	}
	
	if stats == nil {
		stats = &ExecutionStats{}
	}
	stats.RowCount = &rowCount

	// Estimate size (rough calculation)
	if stats.ColumnCount != nil {
		estimatedSize := rowCount * int64(*stats.ColumnCount) * 50 // Rough estimate: 50 bytes per field
		stats.EstimatedSize = &estimatedSize
	}

	return true, issues, stats
}

func (m *ModelValidatorAgent) validateQuality(sqlContent string) (int, []ValidationWarning, []ValidationSuggestion) {
	var warnings []ValidationWarning
	var suggestions []ValidationSuggestion
	score := 100 // Start with perfect score and deduct points

	sqlUpper := strings.ToUpper(sqlContent)
	_ = strings.ToLower(sqlContent) // Keep for potential future use

	// Check for SELECT *
	if strings.Contains(sqlUpper, "SELECT *") {
		score -= 15
		warnings = append(warnings, ValidationWarning{
			Type:       "select_star",
			Message:    "Using SELECT * can impact performance and maintainability",
			Suggestion: "Specify explicit column names for better performance and clarity",
		})
	}

	// Check for consistent naming (prefer snake_case)
	if strings.Contains(sqlContent, "camelCase") || strings.ContainsAny(sqlContent, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		score -= 5
		suggestions = append(suggestions, ValidationSuggestion{
			Type:     "naming_convention",
			Message:  "Consider using consistent snake_case naming convention",
			Impact:   "Improves code readability and maintainability",
			Priority: "medium",
		})
	}

	// Check for proper indentation
	lines := strings.Split(sqlContent, "\n")
	hasProperIndentation := false
	for _, line := range lines {
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			hasProperIndentation = true
			break
		}
	}
	
	if !hasProperIndentation && len(lines) > 3 {
		score -= 10
		suggestions = append(suggestions, ValidationSuggestion{
			Type:     "formatting",
			Message:  "SQL could benefit from proper indentation",
			Impact:   "Improves readability and maintainability",
			Priority: "low",
		})
	}

	// Check for hardcoded values that should be parameters
	if strings.Contains(sqlContent, "'2023'") || strings.Contains(sqlContent, "'2024'") {
		score -= 5
		suggestions = append(suggestions, ValidationSuggestion{
			Type:     "parameterization",
			Message:  "Consider parameterizing hardcoded date values",
			Impact:   "Makes queries more flexible and reusable",
			Priority: "medium",
		})
	}

	// Check for potential performance issues
	if strings.Contains(sqlUpper, "ORDER BY") && !strings.Contains(sqlUpper, "LIMIT") {
		score -= 5
		warnings = append(warnings, ValidationWarning{
			Type:       "performance",
			Message:    "ORDER BY without LIMIT can be expensive on large datasets",
			Suggestion: "Consider adding LIMIT or ensure this is intentional",
		})
	}

	// Check for comments
	hasComments := strings.Contains(sqlContent, "--") || strings.Contains(sqlContent, "/*")
	if !hasComments && len(sqlContent) > 500 {
		score -= 5
		suggestions = append(suggestions, ValidationSuggestion{
			Type:     "documentation",
			Message:  "Complex queries benefit from explanatory comments",
			Impact:   "Improves code maintainability and team collaboration",
			Priority: "low",
		})
	}

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score, warnings, suggestions
}

func (m *ModelValidatorAgent) analyzePerformance(sqlContent string) ([]ValidationWarning, []ValidationSuggestion) {
	var warnings []ValidationWarning
	var suggestions []ValidationSuggestion

	sqlUpper := strings.ToUpper(sqlContent)

	// Check for potentially expensive operations
	if strings.Contains(sqlUpper, "CROSS JOIN") {
		warnings = append(warnings, ValidationWarning{
			Type:       "performance",
			Message:    "CROSS JOIN can produce very large result sets",
			Suggestion: "Ensure this is intentional and consider adding WHERE conditions",
		})
	}

	// Check for subqueries that could be JOINs
	subqueryCount := strings.Count(sqlUpper, "SELECT")
	if subqueryCount > 2 {
		suggestions = append(suggestions, ValidationSuggestion{
			Type:     "performance",
			Message:  "Multiple subqueries detected - consider using JOINs instead",
			Impact:   "Can improve query performance and readability",
			Priority: "medium",
		})
	}

	// Check for functions in WHERE clauses
	if strings.Contains(sqlUpper, "WHERE") && (strings.Contains(sqlUpper, "UPPER(") || strings.Contains(sqlUpper, "LOWER(")) {
		suggestions = append(suggestions, ValidationSuggestion{
			Type:     "performance",
			Message:  "Functions in WHERE clauses can prevent index usage",
			Impact:   "May impact query performance on large datasets",
			Priority: "medium",
		})
	}

	return warnings, suggestions
}

// ValidateModel is the main entry point for validating a model
func (m *ModelValidatorAgent) ValidateModel(ctx context.Context, modelName, sqlContent string) (*ValidationResult, error) {
	if sqlContent == "" {
		return nil, fmt.Errorf("SQL content is required for validation")
	}

	params := map[string]interface{}{
		"sql":        sqlContent,
		"model_name": modelName,
	}

	result, err := m.validateSQLComprehensive(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	validationResult, ok := result.(*ValidationResult)
	if !ok {
		return nil, fmt.Errorf("unexpected validation result type")
	}

	return validationResult, nil
}

// SaveValidatedModel saves a model only if validation passes
func (m *ModelValidatorAgent) SaveValidatedModel(ctx context.Context, modelName, sqlContent string) (*ValidationResult, error) {
	// First validate the model
	validationResult, err := m.ValidateModel(ctx, modelName, sqlContent)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Only save if validation passes
	if !validationResult.ValidationPassed {
		log.Printf("❌ Model %s validation failed, not saving file", modelName)
		return validationResult, fmt.Errorf("model validation failed with %d issues", len(validationResult.Issues))
	}

	// Validation passed, save the file
	err = m.saveModelFile(modelName, sqlContent)
	if err != nil {
		return validationResult, fmt.Errorf("failed to save validated model: %w", err)
	}

	log.Printf("✅ Model %s validated and saved successfully", modelName)
	return validationResult, nil
}

func (m *ModelValidatorAgent) saveModelFile(modelName, sqlContent string) error {
	// Ensure models directory exists
	if err := os.MkdirAll(m.modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	// Ensure .sql extension
	if !strings.HasSuffix(modelName, ".sql") {
		modelName += ".sql"
	}

	filePath := filepath.Join(m.modelsDir, modelName)
	err := os.WriteFile(filePath, []byte(sqlContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write model file %s: %w", modelName, err)
	}

	return nil
}