package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/agent"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
	"github.com/pontus-devoteam/agent-sdk-go/pkg/tool"
)

// ModelEditorAgent is a specialized AI agent that handles SQL model editing with context awareness
type ModelEditorAgent struct {
	*agent.Agent
	runner           *runner.Runner
	validatorAgent   *ModelValidatorAgent
	modelsDir        string
}

// NewModelEditorAgent creates a new ModelEditorAgent
func NewModelEditorAgent(modelName string, runner *runner.Runner, modelsDir string) *ModelEditorAgent {
	a := agent.NewAgent("ModelEditorAgent")
	
	// Create the ModelValidator Agent
	validatorAgent := NewModelValidatorAgent(modelName, runner, modelsDir)
	
	mea := &ModelEditorAgent{
		Agent:          a,
		runner:         runner,
		validatorAgent: validatorAgent,
		modelsDir:      modelsDir,
	}
	
	mea.WithModel(modelName)
	mea.configure()

	return mea
}

func (m *ModelEditorAgent) configure() {
	m.SetSystemInstructions(`You are a ModelEditor Agent that edits SQL models with context awareness.

Core responsibilities:
1. Analyze existing SQL structure and business logic
2. Apply modifications while preserving structure and intent
3. Maintain backward compatibility and data consistency
4. Follow SQL best practices

Key principles:
- Preserve column names/types unless explicitly changing them
- Use snake_case naming conventions
- Keep existing joins/filters intact unless modifying
- Add comments for complex logic
- Place new columns logically in SELECT statements
- Consider downstream impact when changing data types

Output requirements:
- Return ONLY the complete, updated SQL query (no markdown/explanations)
- Maintain proper formatting and original code style
- Ensure syntax correctness and logical consistency`)

	m.WithTools(
		m.createEditSQLTool(),
		m.createValidateEditTool(),
		m.createAnalyzeSQLTool(),
	)
}

func (m *ModelEditorAgent) createEditSQLTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"current_sql": map[string]interface{}{
				"type":        "string",
				"description": "The current SQL content of the model",
			},
			"edit_instructions": map[string]interface{}{
				"type":        "string",
				"description": "Detailed instructions for how to modify the SQL",
			},
			"model_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the model being edited (for context)",
			},
		},
		"required": []string{"current_sql", "edit_instructions"},
	}

	t := tool.NewFunctionTool(
		"edit_sql_with_context",
		"Edits SQL content based on current content and modification instructions",
		m.editSQLWithContext,
	)

	t.WithSchema(schema)
	return t
}

func (m *ModelEditorAgent) createValidateEditTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"original_sql": map[string]interface{}{
				"type":        "string",
				"description": "The original SQL content",
			},
			"updated_sql": map[string]interface{}{
				"type":        "string",
				"description": "The updated SQL content to validate",
			},
		},
		"required": []string{"original_sql", "updated_sql"},
	}

	t := tool.NewFunctionTool(
		"validate_sql_edit",
		"Validates that the SQL edit maintains structure and achieves the requested changes",
		m.validateSQLEdit,
	)

	t.WithSchema(schema)
	return t
}

func (m *ModelEditorAgent) createAnalyzeSQLTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sql_content": map[string]interface{}{
				"type":        "string",
				"description": "SQL content to analyze",
			},
		},
		"required": []string{"sql_content"},
	}

	t := tool.NewFunctionTool(
		"analyze_sql_structure",
		"Analyzes SQL structure to understand tables, columns, joins, and business logic",
		m.analyzeSQLStructure,
	)

	t.WithSchema(schema)
	return t
}

func (m *ModelEditorAgent) editSQLWithContext(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	currentSQL, ok := params["current_sql"].(string)
	if !ok {
		return nil, fmt.Errorf("current_sql parameter is required")
	}

	editInstructions, ok := params["edit_instructions"].(string)
	if !ok {
		return nil, fmt.Errorf("edit_instructions parameter is required")
	}

	modelName, _ := params["model_name"].(string)

	// Create a comprehensive prompt for the AI agent
	prompt := m.createEditPrompt(currentSQL, editInstructions, modelName)

	// Use the runner to process the edit
	runResult, err := m.runner.Run(ctx, m.Agent, &runner.RunOptions{
		Input:    prompt,
		MaxTurns: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process SQL edit: %w", err)
	}

	// Extract the updated SQL
	var updatedSQL string
	if runResult != nil && runResult.FinalOutput != nil {
		if str, ok := runResult.FinalOutput.(string); ok {
			updatedSQL = str
		} else {
			return nil, fmt.Errorf("expected string response from agent, got %T", runResult.FinalOutput)
		}
	} else {
		return nil, fmt.Errorf("no response from agent for edit request")
	}

	return map[string]interface{}{
		"updated_sql":        updatedSQL,
		"original_sql":       currentSQL,
		"edit_instructions":  editInstructions,
		"model_name":         modelName,
		"editing_successful": true,
	}, nil
}

func (m *ModelEditorAgent) validateSQLEdit(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	originalSQL, ok := params["original_sql"].(string)
	if !ok {
		return nil, fmt.Errorf("original_sql parameter is required")
	}

	updatedSQL, ok := params["updated_sql"].(string)
	if !ok {
		return nil, fmt.Errorf("updated_sql parameter is required")
	}

	// Sanitize updated SQL to remove Markdown code fences or language tags
	updatedSQL = sanitizeSQL(updatedSQL)

	// Basic validation checks
	validationResults := make(map[string]interface{})
	
	// Check if SQL is not empty
	if strings.TrimSpace(updatedSQL) == "" {
		validationResults["valid"] = false
		validationResults["error"] = "Updated SQL is empty"
		return validationResults, nil
	}

	// Check if it's still a SELECT statement (basic structure preservation)
	updatedTrimmed := strings.TrimSpace(strings.ToUpper(updatedSQL))
	if !strings.HasPrefix(updatedTrimmed, "SELECT") {
		validationResults["valid"] = false
		validationResults["error"] = "Updated SQL must be a SELECT statement"
		return validationResults, nil
	}

	// Check for basic SQL syntax elements
	requiredElements := []string{"SELECT", "FROM"}
	for _, element := range requiredElements {
		if !strings.Contains(strings.ToUpper(updatedSQL), element) {
			validationResults["valid"] = false
			validationResults["error"] = fmt.Sprintf("Missing required SQL element: %s", element)
			return validationResults, nil
		}
	}

	validationResults["valid"] = true
	validationResults["message"] = "SQL edit validation passed"
	validationResults["changes_detected"] = originalSQL != updatedSQL
	
	return validationResults, nil
}

func (m *ModelEditorAgent) analyzeSQLStructure(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	sqlContent, ok := params["sql_content"].(string)
	if !ok {
		return nil, fmt.Errorf("sql_content parameter is required")
	}

	analysis := make(map[string]interface{})
	
	// Basic SQL analysis
	sqlUpper := strings.ToUpper(sqlContent)
	
	// Check for main SQL components
	analysis["has_select"] = strings.Contains(sqlUpper, "SELECT")
	analysis["has_from"] = strings.Contains(sqlUpper, "FROM")
	analysis["has_where"] = strings.Contains(sqlUpper, "WHERE")
	analysis["has_group_by"] = strings.Contains(sqlUpper, "GROUP BY")
	analysis["has_order_by"] = strings.Contains(sqlUpper, "ORDER BY")
	analysis["has_join"] = strings.Contains(sqlUpper, "JOIN")
	analysis["has_cte"] = strings.Contains(sqlUpper, "WITH")
	
	// Count common SQL functions
	functions := []string{"COUNT", "SUM", "AVG", "MIN", "MAX", "CASE", "COALESCE"}
	functionCount := 0
	for _, fn := range functions {
		if strings.Contains(sqlUpper, fn) {
			functionCount++
		}
	}
	analysis["function_complexity"] = functionCount
	
	// Estimate complexity based on length and components
	complexity := "simple"
	if len(sqlContent) > 1000 || functionCount > 3 || strings.Contains(sqlUpper, "UNION") {
		complexity = "complex"
	} else if len(sqlContent) > 500 || functionCount > 1 {
		complexity = "medium"
	}
	analysis["complexity"] = complexity
	
	// Line count
	analysis["line_count"] = len(strings.Split(sqlContent, "\n"))
	analysis["character_count"] = len(sqlContent)
	
	return analysis, nil
}

func (m *ModelEditorAgent) createEditPrompt(currentSQL, editInstructions, modelName string) string {
	prompt := fmt.Sprintf(`Edit SQL model with context awareness.

MODEL: %s

CURRENT SQL:
%s

EDIT INSTRUCTIONS:
%s

Requirements:
- Analyze structure and preserve intent
- Maintain formatting and naming conventions
- Ensure valid column references
- Return ONLY raw SQL (no markdown/explanations)
- Must be single SELECT statement

Provide updated SQL:`, modelName, currentSQL, editInstructions)

	return prompt
}

// EditModel is the main entry point for editing a model
func (m *ModelEditorAgent) EditModel(ctx context.Context, modelName, currentSQL, editInstructions string) (string, error) {
	if currentSQL == "" {
		return "", fmt.Errorf("current SQL content is required")
	}
	
	if editInstructions == "" {
		return "", fmt.Errorf("edit instructions are required")
	}

	// First analyze the current SQL
	analysisParams := map[string]interface{}{
		"sql_content": currentSQL,
	}
	
	_, err := m.analyzeSQLStructure(ctx, analysisParams)
	if err != nil {
		return "", fmt.Errorf("failed to analyze SQL structure: %w", err)
	}

	// Perform the edit
	editParams := map[string]interface{}{
		"current_sql":       currentSQL,
		"edit_instructions": editInstructions,
		"model_name":        modelName,
	}

	editResult, err := m.editSQLWithContext(ctx, editParams)
	if err != nil {
		return "", fmt.Errorf("failed to edit SQL: %w", err)
	}

	// Extract the updated SQL
	var updatedSQL string
	if editResultMap, ok := editResult.(map[string]interface{}); ok {
		if sql, ok := editResultMap["updated_sql"].(string); ok {
			updatedSQL = sql
		}
	}

	if updatedSQL == "" {
		return "", fmt.Errorf("failed to extract updated SQL from edit result")
	}

	// Log the updated SQL generated by the LLM
	log.Printf("🆕 Updated SQL produced for model %s:\n%s", modelName, updatedSQL)

	// Validate the edit
	validationParams := map[string]interface{}{
		"original_sql": currentSQL,
		"updated_sql":  updatedSQL,
	}

	validationResult, err := m.validateSQLEdit(ctx, validationParams)
	if err != nil {
		log.Printf("❌ validateSQLEdit returned error: %v", err)
		return "", fmt.Errorf("failed to validate SQL edit: %w", err)
	}

	// Check validation results
	if validationResultMap, ok := validationResult.(map[string]interface{}); ok {
		if valid, ok := validationResultMap["valid"].(bool); ok && !valid {
			if errMsg, ok := validationResultMap["error"].(string); ok {
				log.Printf("⚠️ SQL edit validation failed: %s", errMsg)
				return "", fmt.Errorf("SQL edit validation failed: %s", errMsg)
			}
			return "", fmt.Errorf("SQL edit validation failed")
		}
	}

	return updatedSQL, nil
}

// ReadModelFile reads a model file from the models directory
func (m *ModelEditorAgent) ReadModelFile(modelName string) (string, error) {
	// Ensure .sql extension
	if !strings.HasSuffix(modelName, ".sql") {
		modelName += ".sql"
	}

	filePath := filepath.Join(m.modelsDir, modelName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read model file %s: %w", modelName, err)
	}

	return string(content), nil
}

// WriteModelFile writes updated content to a model file
func (m *ModelEditorAgent) WriteModelFile(modelName, content string) error {
	// Ensure models directory exists
	if err := os.MkdirAll(m.modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	// Ensure .sql extension
	if !strings.HasSuffix(modelName, ".sql") {
		modelName += ".sql"
	}

	filePath := filepath.Join(m.modelsDir, modelName)
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write model file %s: %w", modelName, err)
	}

	return nil
}

// EditModelFile is a convenience method that reads, edits, validates, and writes a model file
func (m *ModelEditorAgent) EditModelFile(ctx context.Context, modelName, editInstructions string) (*EditResult, error) {
	// Read current content
	currentSQL, err := m.ReadModelFile(modelName)
	if err != nil {
		return nil, fmt.Errorf("failed to read model file: %w", err)
	}

	// Edit the SQL
	updatedSQL, err := m.EditModel(ctx, modelName, currentSQL, editInstructions)
	if err != nil {
		return nil, fmt.Errorf("failed to edit model: %w", err)
	}

	// Sanitize to remove code fences before validation and saving
	updatedSQL = sanitizeSQL(updatedSQL)

	// Log the updated SQL that will be validated
	log.Printf("📄 Updated SQL to validate for model %s:\n%s", modelName, updatedSQL)

	// Validate the updated SQL using ModelValidator Agent
	log.Printf("🔍 Validating edited model: %s", modelName)
	validationResult, err := m.validatorAgent.ValidateModel(ctx, modelName, updatedSQL)
	if err != nil {
		log.Printf("❌ Validation agent returned error: %v", err)
	}
	if err != nil {
		return &EditResult{
			ModelName:        strings.TrimSuffix(modelName, ".sql"),
			OriginalSQL:      currentSQL,
			UpdatedSQL:       updatedSQL,
			EditInstructions: editInstructions,
			Success:          false,
			ValidationResult: nil,
			ValidationPassed: false,
			FileSaved:        false,
		}, fmt.Errorf("validation failed: %w", err)
	}

	result := &EditResult{
		ModelName:        strings.TrimSuffix(modelName, ".sql"),
		OriginalSQL:      currentSQL,
		UpdatedSQL:       updatedSQL,
		EditInstructions: editInstructions,
		ValidationResult: validationResult,
		ValidationPassed: validationResult.ValidationPassed,
		FileSaved:        false,
	}

	// Only save if validation passes
	if validationResult.ValidationPassed {
		log.Printf("✅ Validation passed, saving model file: %s", modelName)
		err = m.WriteModelFile(modelName, updatedSQL)
		if err != nil {
			result.Success = false
			return result, fmt.Errorf("failed to write validated model: %w", err)
		}
		result.FileSaved = true
		result.Success = true
		log.Printf("💾 Model %s successfully edited, validated, and saved", modelName)
	} else {
		log.Printf("❌ Validation failed for model %s with %d issues, details: %+v, file not saved", 
			modelName, len(validationResult.Issues), validationResult)
		result.Success = false
		return result, fmt.Errorf("model validation failed - file not saved due to %d validation issues", len(validationResult.Issues))
	}

	return result, nil
}



// EditResult contains the results of a model edit operation
type EditResult struct {
	ModelName         string             `json:"model_name"`
	OriginalSQL       string             `json:"original_sql"`
	UpdatedSQL        string             `json:"updated_sql"`
	EditInstructions  string             `json:"edit_instructions"`
	Success           bool               `json:"success"`
	ValidationResult  *ValidationResult  `json:"validation_result,omitempty"`
	ValidationPassed  bool               `json:"validation_passed"`
	FileSaved         bool               `json:"file_saved"`
}