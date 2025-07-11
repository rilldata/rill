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

// ProjectEditorAgent is an AI agent that manages SQL model creation and coordination
// It can hand off to SyntheticDataAgent for generating synthetic data models
// and to ModelEditorAgent for editing existing models
type ProjectEditorAgent struct {
	*agent.Agent
	runner             *runner.Runner
	syntheticDataAgent *SyntheticDataAgent
	modelEditorAgent   *ModelEditorAgent
	projectDir         string
	modelsDir          string
}

// NewProjectEditorAgent creates a new ProjectEditorAgent with handoff capabilities
func NewProjectEditorAgent(modelName string, runner *runner.Runner, syntheticDataAgent *SyntheticDataAgent, projectDir string) *ProjectEditorAgent {
	a := agent.NewAgent("ProjectEditorAgent")
	modelsDir := filepath.Join(projectDir, "models")

	// Create the ModelEditor Agent
	modelEditorAgent := NewModelEditorAgent(modelName, runner, modelsDir)

	pea := &ProjectEditorAgent{
		Agent:              a,
		runner:             runner,
		syntheticDataAgent: syntheticDataAgent,
		modelEditorAgent:   modelEditorAgent,
		projectDir:         projectDir,
		modelsDir:          modelsDir,
	}

	pea.WithModel(modelName)
	pea.configure()
	pea.setupHandoffs()

	return pea
}

func (p *ProjectEditorAgent) configure() {
	p.SetSystemInstructions(`You are a ProjectEditor Agent that orchestrates SQL model management for Rill projects.

Your primary responsibilities:
1. Create new SQL model files in the models/ directory
2. Coordinate model editing by handing off to ModelEditorAgent
3. Infer model names from user prompts and context
4. Hand off to SyntheticDataAgent for synthetic data generation
5. Manage the project file structure and coordination between specialized agents

DECISION LOGIC:
- If user mentions "synthetic data", "generate data", or "sample data" -> Hand off to SyntheticDataAgent
- If user wants to edit existing models -> Hand off to ModelEditorAgent with full context
- If user provides specific SQL or wants to create new model -> Create new model file directly
- Always infer appropriate model names from context (e.g., "sales data" -> "sales_model")

FILE NAMING CONVENTIONS:
- Use snake_case for model names
- Add "_model" suffix if not already present
- Save files as .sql in models/ directory
- Examples: sales_model.sql, customer_analytics_model.sql, revenue_report_model.sql

HANDOFF SCENARIOS:
1. SYNTHETIC DATA GENERATION:
   - User asks for "synthetic sales data" -> Hand off to SyntheticDataAgent
   - User asks for "sample customer data" -> Hand off to SyntheticDataAgent
   - User asks for "generate test data for marketing" -> Hand off to SyntheticDataAgent

2. MODEL EDITING:
   - User says "edit my sales model" -> Hand off to ModelEditorAgent
   - User asks to "update the customer model to include..." -> Hand off to ModelEditorAgent
   - User wants to "modify the revenue calculations" -> Hand off to ModelEditorAgent
   - ModelEditorAgent handles reading existing content and context-aware editing

3. DIRECT MODEL CREATION:
   - User provides specific SQL -> Create new model file directly
   - User asks for new business logic without editing existing -> Create new model

COORDINATION RESPONSIBILITIES:
- Route requests to the most appropriate specialized agent
- Manage file system operations (create directories, list files, etc.)
- Provide project-level overview and status
- Handle simple file operations that don't require specialized agents

When working with files:
- Always check if models/ directory exists, create if needed
- Use appropriate error handling for file operations
- Provide clear feedback about file operations and agent handoffs
- Trust specialized agents to handle their domain expertise`)

	p.WithTools(
		p.createListModelsToool(),
		p.createReadModelTool(),
		p.createWriteModelTool(),
	)
}

func (p *ProjectEditorAgent) setupHandoffs() {
	// Register proper agent handoffs instead of custom tool wrappers
	p.Agent.WithHandoffs(
		p.modelEditorAgent.Agent,
		p.syntheticDataAgent.Agent,
	)
}

func (p *ProjectEditorAgent) createListModelsToool() tool.Tool {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}

	t := tool.NewFunctionTool(
		"list_models",
		"Lists all existing SQL model files in the models directory",
		p.listModels,
	)

	t.WithSchema(schema)
	return t
}

func (p *ProjectEditorAgent) createReadModelTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"model_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the model file to read (without .sql extension)",
			},
		},
		"required": []string{"model_name"},
	}

	t := tool.NewFunctionTool(
		"read_model",
		"Reads the content of an existing SQL model file",
		p.readModel,
	)

	t.WithSchema(schema)
	return t
}

func (p *ProjectEditorAgent) createWriteModelTool() tool.Tool {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"model_name": map[string]interface{}{
				"type":        "string",
				"description": "Name of the model file to write (without .sql extension)",
			},
			"sql_content": map[string]interface{}{
				"type":        "string",
				"description": "SQL content to write to the model file",
			},
		},
		"required": []string{"model_name", "sql_content"},
	}

	t := tool.NewFunctionTool(
		"write_model",
		"Writes SQL content to a model file, creating or updating as needed",
		p.writeModel,
	)

	t.WithSchema(schema)
	return t
}

func (p *ProjectEditorAgent) listModels(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Ensure models directory exists
	if err := os.MkdirAll(p.modelsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create models directory: %w", err)
	}

	files, err := os.ReadDir(p.modelsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read models directory: %w", err)
	}

	var models []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			modelName := strings.TrimSuffix(file.Name(), ".sql")
			models = append(models, modelName)
		}
	}

	return map[string]interface{}{
		"models":    models,
		"count":     len(models),
		"directory": p.modelsDir,
	}, nil
}

func (p *ProjectEditorAgent) readModel(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	modelName, ok := params["model_name"].(string)
	if !ok {
		return nil, fmt.Errorf("model_name parameter is required")
	}

	// Ensure .sql extension
	if !strings.HasSuffix(modelName, ".sql") {
		modelName += ".sql"
	}

	filePath := filepath.Join(p.modelsDir, modelName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read model file %s: %w", modelName, err)
	}

	return map[string]interface{}{
		"model_name": strings.TrimSuffix(modelName, ".sql"),
		"file_path":  filePath,
		"content":    string(content),
	}, nil
}

func (p *ProjectEditorAgent) writeModel(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	modelName, ok := params["model_name"].(string)
	if !ok {
		return nil, fmt.Errorf("model_name parameter is required")
	}

	sqlContent, ok := params["sql_content"].(string)
	if !ok {
		return nil, fmt.Errorf("sql_content parameter is required")
	}

	// Ensure models directory exists
	if err := os.MkdirAll(p.modelsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create models directory: %w", err)
	}

	// Ensure .sql extension
	if !strings.HasSuffix(modelName, ".sql") {
		modelName += ".sql"
	}

	filePath := filepath.Join(p.modelsDir, modelName)

	// Check if file exists to determine if this is create or update
	var operation string
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		operation = "created"
	} else {
		operation = "updated"
	}

	err := os.WriteFile(filePath, []byte(sqlContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write model file %s: %w", modelName, err)
	}

	return map[string]interface{}{
		"model_name": strings.TrimSuffix(modelName, ".sql"),
		"file_path":  filePath,
		"operation":  operation,
		"success":    true,
	}, nil
}

func (p *ProjectEditorAgent) handoffToModelEditor(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	modelName, ok := params["model_name"].(string)
	if !ok {
		return nil, fmt.Errorf("model_name parameter is required")
	}

	editInstructions, ok := params["edit_instructions"].(string)
	if !ok {
		return nil, fmt.Errorf("edit_instructions parameter is required")
	}

	// Hand off to ModelEditorAgent
	if p.modelEditorAgent == nil {
		return nil, fmt.Errorf("ModelEditorAgent is not available for handoff")
	}

	// Use the ModelEditor agent to edit the model file with retry on validation failure
	maxAttempts := 3
	attempt := 1
	var editResult *EditResult
	var err error
	currentInstructions := editInstructions
	for attempt <= maxAttempts {
		log.Printf("📝 ModelEditor attempt %d/%d for model %s", attempt, maxAttempts, modelName)
		editResult, err = p.modelEditorAgent.EditModelFile(ctx, modelName, currentInstructions)
		if err == nil {
			// Success
			break
		}

		// Check if the error is due to validation
		if strings.Contains(err.Error(), "validation failed") || strings.Contains(err.Error(), "SQL edit validation failed") {
			log.Printf("⚠️ Validation error on attempt %d: %v", attempt, err)
			// Append validation feedback to instructions and retry
			attempt++
			currentInstructions = fmt.Sprintf("%s\n\nPlease fix the following validation errors and try again: %v", editInstructions, err)
			continue
		}
		// Non-validation error, abort
		log.Printf("❌ Non-validation error on attempt %d: %v", attempt, err)
		return nil, fmt.Errorf("failed to edit model via ModelEditorAgent: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to edit model via ModelEditorAgent after %d attempts: %w", maxAttempts, err)
	}

	return map[string]interface{}{
		"handoff_successful": true,
		"agent_used":         "ModelEditorAgent",
		"model_name":         editResult.ModelName,
		"original_sql":       editResult.OriginalSQL,
		"updated_sql":        editResult.UpdatedSQL,
		"edit_instructions":  editResult.EditInstructions,
		"edit_successful":    editResult.Success,
		"validation_passed":  editResult.ValidationPassed,
		"file_saved":         editResult.FileSaved,
		"validation_result":  editResult.ValidationResult,
	}, nil
}

func (p *ProjectEditorAgent) handoffToSyntheticData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	description, ok := params["description"].(string)
	if !ok {
		return nil, fmt.Errorf("description parameter is required")
	}

	// Infer model name if not provided
	modelName, ok := params["model_name"].(string)
	if !ok || modelName == "" {
		modelName = p.inferModelName(description)
	}

	// Ensure proper naming convention
	modelName = p.normalizeModelName(modelName)

	// Hand off to SyntheticDataAgent
	if p.syntheticDataAgent == nil {
		return nil, fmt.Errorf("SyntheticDataAgent is not available for handoff")
	}

	sql, err := p.syntheticDataAgent.GenerateSQL(description)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL via SyntheticDataAgent: %w", err)
	}

	// Save the generated SQL to a model file
	writeParams := map[string]interface{}{
		"model_name":  modelName,
		"sql_content": sql,
	}

	result, err := p.writeModel(ctx, writeParams)
	if err != nil {
		return nil, fmt.Errorf("failed to save generated model: %w", err)
	}

	return map[string]interface{}{
		"handoff_successful": true,
		"generated_sql":      sql,
		"model_saved":        result,
		"model_name":         modelName,
	}, nil
}

// Helper function to infer model name from description
func (p *ProjectEditorAgent) inferModelName(description string) string {
	description = strings.ToLower(description)

	// Common data types and their model names
	keywords := map[string]string{}

	// Check for keywords in description
	for keyword, modelName := range keywords {
		if strings.Contains(description, keyword) {
			return modelName
		}
	}

	// Default fallback
	return "generated_model"
}

// Helper function to normalize model name
func (p *ProjectEditorAgent) normalizeModelName(modelName string) string {
	// Convert to lowercase and replace spaces/hyphens with underscores
	modelName = strings.ToLower(modelName)
	modelName = strings.ReplaceAll(modelName, " ", "_")
	modelName = strings.ReplaceAll(modelName, "-", "_")

	// Remove any non-alphanumeric characters except underscores
	var result strings.Builder
	for _, char := range modelName {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_' {
			result.WriteRune(char)
		}
	}
	modelName = result.String()

	// Ensure it ends with _model if not already
	if !strings.HasSuffix(modelName, "_model") {
		modelName += "_model"
	}

	return modelName
}

// ProcessUserRequest is the main entry point for handling user requests
func (p *ProjectEditorAgent) ProcessUserRequest(ctx context.Context, userPrompt string) (interface{}, error) {
	// Use the runner to process the user request
	runResult, err := p.runner.Run(ctx, p.Agent, &runner.RunOptions{
		Input:    userPrompt,
		MaxTurns: 10,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to process user request: %w", err)
	}

	return runResult, nil
}

// InferOperationType determines whether to create new or edit existing model
func (p *ProjectEditorAgent) InferOperationType(ctx context.Context, userPrompt string) (string, string, error) {
	userPrompt = strings.ToLower(userPrompt)

	// Check for edit keywords
	editKeywords := []string{"edit", "update", "modify", "change", "fix", "improve", "alter", "add", "append", "insert"}
	for _, keyword := range editKeywords {
		if strings.Contains(userPrompt, keyword) {
			// Try to infer which model to edit
			modelName := p.inferModelName(userPrompt)

			// Check if the model exists
			readParams := map[string]interface{}{"model_name": modelName}
			_, err := p.readModel(ctx, readParams)
			if err == nil {
				return "edit_with_context", modelName, nil
			}

			// If inferred model doesn't exist, try to find similar models
			listResult, listErr := p.listModels(ctx, map[string]interface{}{})
			if listErr == nil {
				if modelList, ok := listResult.(map[string]interface{}); ok {
					if models, ok := modelList["models"].([]string); ok {
						// Look for partial matches
						for _, existingModel := range models {
							if strings.Contains(existingModel, strings.TrimSuffix(modelName, "_model")) {
								return "edit_with_context", existingModel, nil
							}
						}
					}
				}
			}
		}
	}

	// Check for synthetic data keywords
	syntheticKeywords := []string{"synthetic", "generate", "sample", "test data", "demo data", "fake data"}
	for _, keyword := range syntheticKeywords {
		if strings.Contains(userPrompt, keyword) {
			modelName := p.inferModelName(userPrompt)
			return "synthetic", modelName, nil
		}
	}

	// Default to create new
	modelName := p.inferModelName(userPrompt)
	return "create", modelName, nil
}

// ProcessUserRequestWithContext is an enhanced version that handles context passing for edits
func (p *ProjectEditorAgent) ProcessUserRequestWithContext(ctx context.Context, userPrompt string) (interface{}, error) {
	// First, determine the operation type
	operationType, modelName, err := p.InferOperationType(ctx, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to infer operation type: %w", err)
	}

	// Handle different operation types
	switch operationType {
	case "edit_with_context":
		// For edits, hand off to ModelEditor agent
		editParams := map[string]interface{}{
			"model_name":        modelName,
			"edit_instructions": userPrompt,
		}
		return p.handoffToModelEditor(ctx, editParams)

	case "synthetic":
		// Hand off to SyntheticDataAgent
		syntheticParams := map[string]interface{}{
			"description": userPrompt,
			"model_name":  modelName,
		}
		return p.handoffToSyntheticData(ctx, syntheticParams)

	default:
		// For create operations or other cases, use the standard processing
		return p.ProcessUserRequest(ctx, userPrompt)
	}
}
