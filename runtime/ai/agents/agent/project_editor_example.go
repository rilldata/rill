package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/pontus-devoteam/agent-sdk-go/pkg/runner"
)

// ExampleProjectEditorUsage demonstrates how to use the ProjectEditor Agent
func ExampleProjectEditorUsage() {
	// Initialize the runner (this would typically be done in main.go)
	runner := runner.NewRunner()
	
	// Create SyntheticDataAgent
	syntheticAgent := NewSyntheticDataAgent("claude-3-5-sonnet-20241022", runner)
	
	// Create ProjectEditor Agent with handoff capability
	projectDir := "/path/to/your/rill/project"
	projectEditor := NewProjectEditorAgent("claude-3-5-sonnet-20241022", runner, syntheticAgent, projectDir)
	
	ctx := context.Background()
	
	// Example 1: Generate synthetic sales data (hands off to SyntheticDataAgent)
	fmt.Println("=== Example 1: Generate Synthetic Sales Data ===")
	result1, err := projectEditor.ProcessUserRequest(ctx, "Generate synthetic sales data with customer information, product details, and transaction amounts")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Result: %+v\n", result1)
	}
	
	// Example 2: Create a custom model
	fmt.Println("\n=== Example 2: Create Custom Marketing Model ===")
	result2, err := projectEditor.ProcessUserRequest(ctx, "Create a marketing campaign model that tracks campaign performance, cost per click, and conversion rates")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Result: %+v\n", result2)
	}
	
	// Example 3: Edit existing model with context
	fmt.Println("\n=== Example 3: Edit Existing Model with Context ===")
	result3, err := projectEditor.ProcessUserRequestWithContext(ctx, "Edit the sales model to include discount information and seasonal adjustments")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Result: %+v\n", result3)
	}
	
	// Example 4: List all models
	fmt.Println("\n=== Example 4: List All Models ===")
	result4, err := projectEditor.ProcessUserRequest(ctx, "List all the model files in this project")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Result: %+v\n", result4)
	}
	
	// Example 5: Generate test data for specific domain
	fmt.Println("\n=== Example 5: Generate E-commerce Test Data ===")
	result5, err := projectEditor.ProcessUserRequest(ctx, "Generate sample e-commerce data with orders, customers, and products for testing")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Result: %+v\n", result5)
	}
}

// ExampleDirectToolUsage shows how to use the agent tools directly
func ExampleDirectToolUsage() {
	// Initialize agents
	runner := runner.NewRunner()
	syntheticAgent := NewSyntheticDataAgent("claude-3-5-sonnet-20241022", runner)
	projectDir := "/path/to/your/rill/project"
	projectEditor := NewProjectEditorAgent("claude-3-5-sonnet-20241022", runner, syntheticAgent, projectDir)
	
	ctx := context.Background()
	
	// Example of direct tool usage
	fmt.Println("=== Direct Tool Usage Examples ===")
	
	// 1. List models
	fmt.Println("\n1. List Models:")
	listResult, err := projectEditor.listModels(ctx, map[string]interface{}{})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Models: %+v\n", listResult)
	}
	
	// 2. Generate synthetic data via handoff
	fmt.Println("\n2. Generate Synthetic Data:")
	syntheticResult, err := projectEditor.handoffToSyntheticData(ctx, map[string]interface{}{
		"description": "Generate sales transaction data with customer demographics and product information",
		"model_name":  "sales_analytics_model",
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Synthetic Data Result: %+v\n", syntheticResult)
	}
	
	// 3. Read a model
	fmt.Println("\n3. Read Model:")
	readResult, err := projectEditor.readModel(ctx, map[string]interface{}{
		"model_name": "sales_analytics_model",
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Model Content: %+v\n", readResult)
	}
	
	// 4. Write a custom model
	fmt.Println("\n4. Write Custom Model:")
	customSQL := `SELECT 
    generate_series AS id,
    current_date - (random() * 30)::int AS date,
    ['A', 'B', 'C'][((random() * 3)::int + 1)] AS category,
    (random() * 1000)::decimal(10,2) AS amount
FROM generate_series(1, 1000)`
	
	writeResult, err := projectEditor.writeModel(ctx, map[string]interface{}{
		"model_name":  "custom_test_model",
		"sql_content": customSQL,
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Write Result: %+v\n", writeResult)
	}
	
	// 5. Edit model via ModelEditor agent handoff
	fmt.Println("\n5. Edit Model via ModelEditor Agent Handoff:")
	editResult, err := projectEditor.handoffToModelEditor(ctx, map[string]interface{}{
		"model_name":        "custom_test_model",
		"edit_instructions": "Add a new column for region and modify the amount calculation to include tax",
	})
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Edit Result: %+v\n", editResult)
	}
}

// ExampleInferenceLogic demonstrates the inference capabilities
func ExampleInferenceLogic() {
	runner := runner.NewRunner()
	syntheticAgent := NewSyntheticDataAgent("claude-3-5-sonnet-20241022", runner)
	projectDir := "/path/to/your/rill/project"
	projectEditor := NewProjectEditorAgent("claude-3-5-sonnet-20241022", runner, syntheticAgent, projectDir)
	
	ctx := context.Background()
	
	// Test inference logic
	testPrompts := []string{
		"Generate synthetic sales data",
		"Edit the customer model to add age information",
		"Create a new marketing campaign tracker",
		"Update the existing revenue model",
		"Generate sample e-commerce transaction data",
		"Modify the product model to include categories",
		"Create test data for user engagement metrics",
	}
	
	fmt.Println("=== Inference Logic Examples ===")
	for i, prompt := range testPrompts {
		operation, modelName, err := projectEditor.InferOperationType(ctx, prompt)
		if err != nil {
			log.Printf("Error inferring operation for prompt %d: %v", i+1, err)
			continue
		}
		fmt.Printf("Prompt %d: %s\n", i+1, prompt)
		fmt.Printf("  Operation: %s\n", operation)
		fmt.Printf("  Model Name: %s\n", modelName)
		fmt.Printf("  ---\n")
	}
}

// ExampleUsagePatterns shows common usage patterns
func ExampleUsagePatterns() {
	fmt.Println("=== Common Usage Patterns ===")
	
	fmt.Println("\n1. SYNTHETIC DATA GENERATION:")
	fmt.Println("   - 'Generate synthetic sales data'")
	fmt.Println("   - 'Create sample customer data for testing'")
	fmt.Println("   - 'Generate demo e-commerce transactions'")
	fmt.Println("   - 'Create test data for marketing campaigns'")
	
	fmt.Println("\n2. MODEL EDITING:")
	fmt.Println("   - 'Edit the sales model to include discounts'")
	fmt.Println("   - 'Update customer model with demographic info'")
	fmt.Println("   - 'Modify the revenue model to add regions'")
	fmt.Println("   - 'Fix the product model query'")
	
	fmt.Println("\n3. NEW MODEL CREATION:")
	fmt.Println("   - 'Create a new inventory tracking model'")
	fmt.Println("   - 'Build a customer lifetime value model'")
	fmt.Println("   - 'Create a marketing ROI analysis model'")
	fmt.Println("   - 'Generate a subscription analytics model'")
	
	fmt.Println("\n4. PROJECT MANAGEMENT:")
	fmt.Println("   - 'List all model files'")
	fmt.Println("   - 'Show me the customer model'")
	fmt.Println("   - 'What models exist in this project?'")
	fmt.Println("   - 'Read the sales model content'")
	
	fmt.Println("\n5. HANDOFF SCENARIOS:")
	fmt.Println("   - User mentions 'synthetic', 'generate', 'sample', 'test data' -> SyntheticDataAgent")
	fmt.Println("   - User mentions 'edit', 'update', 'modify', 'change' -> Edit existing with context")
	fmt.Println("   - User provides specific SQL -> Direct model creation")
	fmt.Println("   - User asks for new business logic -> Create new model")
	
	fmt.Println("\n6. MODEL EDITOR AGENT HANDOFFS:")
	fmt.Println("   - ProjectEditor hands off editing tasks to specialized ModelEditor")
	fmt.Println("   - ModelEditor automatically reads existing model content")
	fmt.Println("   - Includes original SQL in the context for AI processing")
	fmt.Println("   - Maintains structure and intent of original model")
	fmt.Println("   - Provides comprehensive before/after comparison")
	fmt.Println("   - Examples:")
	fmt.Println("     • 'Edit sales model to add discount column' -> ModelEditorAgent")
	fmt.Println("     • 'Update customer model with demographics' -> ModelEditorAgent")
	fmt.Println("     • 'Generate synthetic e-commerce data' -> SyntheticDataAgent")
	
	fmt.Println("\n7. SPECIALIZED AGENT ARCHITECTURE:")
	fmt.Println("   - ProjectEditor: Orchestrates and coordinates requests")
	fmt.Println("   - ModelEditor: Specialized in context-aware SQL editing")
	fmt.Println("   - ModelValidator: Validates SQL for correctness and quality")
	fmt.Println("   - SyntheticDataAgent: Specialized in generating test data")
	fmt.Println("   - Each agent has domain-specific expertise and tools")
	
	fmt.Println("\n8. VALIDATION WORKFLOW:")
	fmt.Println("   - ModelEditor edits SQL with full context")
	fmt.Println("   - ModelValidator validates syntax, execution, and quality")
	fmt.Println("   - File is saved ONLY if validation passes")
	fmt.Println("   - Validation results include quality scores and suggestions")
	fmt.Println("   - Failed validation prevents corrupted models from being saved")
}

// ExampleModelEditorUsage demonstrates direct usage of the ModelEditor agent
func ExampleModelEditorUsage() {
	runner := runner.NewRunner()
	modelsDir := "/path/to/your/rill/project/models"
	modelEditor := NewModelEditorAgent("claude-3-5-sonnet-20241022", runner, modelsDir)
	
	ctx := context.Background()
	
	fmt.Println("=== ModelEditor Agent Direct Usage ===")
	
	// Example 1: Edit a model with specific instructions
	fmt.Println("\n1. Edit Model with Context:")
	editResult, err := modelEditor.EditModelFile(ctx, "sales_model", "Add a discount_percentage column and modify the total_amount calculation to apply discounts")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Edit successful: %+v\n", editResult)
		fmt.Printf("Original SQL length: %d characters\n", len(editResult.OriginalSQL))
		fmt.Printf("Updated SQL length: %d characters\n", len(editResult.UpdatedSQL))
	}
	
	// Example 2: Read and analyze SQL structure
	fmt.Println("\n2. Read Model File:")
	content, err := modelEditor.ReadModelFile("customer_model")
	if err != nil {
		log.Printf("Error reading model: %v", err)
	} else {
		fmt.Printf("Model content preview: %.100s...\n", content)
	}
	
	// Example 3: Complex edit with multiple changes
	fmt.Println("\n3. Complex Model Edit:")
	complexEditResult, err := modelEditor.EditModelFile(ctx, "revenue_model", `
		Please make the following changes:
		1. Add a new column 'fiscal_quarter' calculated from the date
		2. Include tax calculations (8.5% rate)
		3. Add a region grouping based on customer location
		4. Ensure all monetary values are rounded to 2 decimal places
	`)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Complex edit successful: %t\n", complexEditResult.Success)
		fmt.Printf("Changes applied: %s\n", complexEditResult.EditInstructions)
	}
}

// ExampleAgentCoordination shows how agents work together
func ExampleAgentCoordination() {
	fmt.Println("=== Agent Coordination Example ===")
	
	fmt.Println("\nTYPICAL WORKFLOW:")
	fmt.Println("1. User: 'Generate synthetic sales data and then edit it to include regions'")
	fmt.Println("   → ProjectEditor analyzes request")
	fmt.Println("   → Hands off to SyntheticDataAgent for data generation")
	fmt.Println("   → SyntheticDataAgent creates sales_model.sql")
	fmt.Println("   → ProjectEditor hands off to ModelEditor for region addition")
	fmt.Println("   → ModelEditor reads existing sales_model.sql content")
	fmt.Println("   → ModelEditor applies region modifications with context")
	fmt.Println("   → Final result: contextually-aware edited model")
	
	fmt.Println("\nAGENT SPECIALIZATIONS:")
	fmt.Println("• ProjectEditor: Request routing, file management, coordination")
	fmt.Println("• SyntheticDataAgent: DuckDB SQL generation, data validation")
	fmt.Println("• ModelEditor: Context-aware editing, SQL analysis, structure preservation")
	
	fmt.Println("\nBENEFITS:")
	fmt.Println("• Separation of concerns - each agent focuses on its expertise")
	fmt.Println("• Context preservation - editing maintains original intent")
	fmt.Println("• Scalability - easy to add new specialized agents")
	fmt.Println("• Reliability - specialized validation for each domain")
}

// ExampleModelValidatorUsage demonstrates the ModelValidator agent
func ExampleModelValidatorUsage() {
	runner := runner.NewRunner()
	modelsDir := "/path/to/your/rill/project/models"
	validator := NewModelValidatorAgent("claude-3-5-sonnet-20241022", runner, modelsDir)
	
	ctx := context.Background()
	
	fmt.Println("=== ModelValidator Agent Usage ===")
	
	// Example 1: Validate good SQL
	fmt.Println("\n1. Validate Well-Written SQL:")
	goodSQL := `SELECT 
		customer_id,
		order_date,
		SUM(amount) AS total_amount,
		COUNT(*) AS order_count
	FROM orders 
	WHERE order_date >= current_date - interval '30 days'
	GROUP BY customer_id, order_date
	ORDER BY total_amount DESC
	LIMIT 100`
	
	result1, err := validator.ValidateModel(ctx, "good_example", goodSQL)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Validation passed: %t\n", result1.ValidationPassed)
		fmt.Printf("Quality score: %d/100\n", result1.QualityScore)
		fmt.Printf("Issues: %d, Warnings: %d, Suggestions: %d\n", 
			len(result1.Issues), len(result1.Warnings), len(result1.Suggestions))
	}
	
	// Example 2: Validate problematic SQL
	fmt.Println("\n2. Validate Problematic SQL:")
	badSQL := `SELECT * FROM orders WHERE UPPER(customer_name) = 'JOHN' ORDER BY amount`
	
	result2, err := validator.ValidateModel(ctx, "problematic_example", badSQL)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Validation passed: %t\n", result2.ValidationPassed)
		fmt.Printf("Quality score: %d/100\n", result2.QualityScore)
		fmt.Printf("Issues: %d, Warnings: %d, Suggestions: %d\n", 
			len(result2.Issues), len(result2.Warnings), len(result2.Suggestions))
		
		// Show some validation feedback
		for _, warning := range result2.Warnings {
			fmt.Printf("⚠️  Warning: %s\n", warning.Message)
		}
		for _, suggestion := range result2.Suggestions {
			fmt.Printf("💡 Suggestion: %s\n", suggestion.Message)
		}
	}
	
	// Example 3: Save with validation
	fmt.Println("\n3. Save Model with Validation:")
	testSQL := `SELECT 
		generate_series AS id,
		current_date - (random() * 90)::int AS date,
		(random() * 1000)::decimal(10,2) AS amount
	FROM generate_series(1, 1000)`
	
	saveResult, err := validator.SaveValidatedModel(ctx, "test_model", testSQL)
	if err != nil {
		log.Printf("Save failed: %v", err)
	} else {
		fmt.Printf("Model saved successfully: %t\n", saveResult.ValidationPassed)
		fmt.Printf("Execution time: %v\n", saveResult.ExecutionTime)
		if saveResult.RowCount != nil {
			fmt.Printf("Sample rows: %d\n", *saveResult.RowCount)
		}
	}
}