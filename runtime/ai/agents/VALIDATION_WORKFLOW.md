# Model Validation Workflow

This document describes the comprehensive validation workflow implemented in the agent architecture.

## Architecture Overview

```
User Request
     ↓
ProjectEditor Agent (Orchestrator)
     ↓
ModelEditor Agent (Context-aware editing)
     ↓
ModelValidator Agent (Comprehensive validation)
     ↓
File saved ONLY if validation passes
```

## Agent Responsibilities

### 1. **ProjectEditor Agent**
- **Role**: Orchestrator and request router
- **Responsibilities**:
  - Analyze user requests and determine operation type
  - Route to appropriate specialized agents
  - Coordinate multi-agent workflows
  - Manage project-level file operations

### 2. **ModelEditor Agent**
- **Role**: Context-aware SQL editing specialist
- **Responsibilities**:
  - Read existing model content for context
  - Apply modifications while preserving structure and intent
  - Hand off to ModelValidator for validation
  - Save files only after successful validation

### 3. **ModelValidator Agent** ⭐ **NEW**
- **Role**: Comprehensive SQL validation specialist
- **Responsibilities**:
  - **Syntax Validation**: DuckDB syntax correctness
  - **Execution Validation**: Query can run without errors
  - **Quality Validation**: Best practices and coding standards
  - **Performance Analysis**: Optimization opportunities
  - **File Protection**: Prevent saving of invalid models

### 4. **SyntheticDataAgent**
- **Role**: Test data generation specialist
- **Responsibilities**:
  - Generate realistic business data using DuckDB SQL
  - Validate generated SQL for correctness
  - Create domain-specific synthetic datasets

## Validation Workflow

### Step 1: User Request Processing
```
User: "Edit the sales model to add discount calculations"
↓
ProjectEditor: Detects edit operation, routes to ModelEditor
```

### Step 2: Context-Aware Editing
```
ModelEditor:
1. Reads existing sales_model.sql content
2. Creates context-aware prompt with original SQL
3. AI agent applies modifications preserving structure
4. Generates updated SQL
```

### Step 3: Comprehensive Validation
```
ModelValidator performs:
✅ Syntax Validation (DuckDB parsing)
✅ Execution Validation (sample query execution)
✅ Quality Assessment (coding standards, best practices)
✅ Performance Analysis (optimization suggestions)
```

### Step 4: Conditional File Saving
```
IF validation passes:
  ✅ Save updated model file
  ✅ Return success with validation details
ELSE:
  ❌ Do NOT save file
  ❌ Return failure with specific issues
```

## Validation Categories

### 1. **Syntax Validation**
- SQL syntax correctness using DuckDB parser
- Proper table and column references
- Correct function usage and data types
- **Result**: Pass/Fail (blocking)

### 2. **Execution Validation**
- Query executes without runtime errors
- Reasonable execution time limits
- Sample data validation (up to 100 rows)
- **Result**: Pass/Fail (blocking)

### 3. **Quality Validation** (Scoring: 0-100)
- Consistent naming conventions (snake_case)
- Proper indentation and formatting
- Avoiding `SELECT *` in production queries
- Meaningful column aliases and comments
- **Result**: Quality score with warnings

### 4. **Performance Analysis**
- Efficient join strategies
- Index-friendly query patterns
- Appropriate WHERE clause usage
- Cross join and subquery optimization
- **Result**: Suggestions for improvement

## Validation Results

### Success Response
```json
{
  "handoff_successful": true,
  "agent_used": "ModelEditorAgent",
  "model_name": "sales_model",
  "validation_passed": true,
  "file_saved": true,
  "validation_result": {
    "quality_score": 85,
    "issues": [],
    "warnings": ["Consider adding LIMIT clause"],
    "suggestions": ["Add index on date column"]
  }
}
```

### Failure Response
```json
{
  "handoff_successful": true,
  "agent_used": "ModelEditorAgent", 
  "model_name": "sales_model",
  "validation_passed": false,
  "file_saved": false,
  "validation_result": {
    "quality_score": 45,
    "issues": [
      {
        "type": "syntax_error",
        "severity": "error", 
        "message": "Column 'invalid_col' does not exist"
      }
    ],
    "warnings": ["Using SELECT * impacts performance"],
    "suggestions": ["Specify explicit column names"]
  }
}
```

## Web Interface Experience

### Successful Edit
```
Model: sales_model
Successfully handed off to ModelEditorAgent

✅ Validation: PASSED
💾 File: SAVED

Updated SQL:
SELECT 
    order_id,
    customer_id,
    amount * (1 - COALESCE(discount_percentage, 0)/100) AS final_amount
FROM orders
```

### Failed Edit
```
Model: sales_model  
Successfully handed off to ModelEditorAgent

❌ Validation: FAILED
⚠️  File: NOT SAVED (validation failed)

Updated SQL:
[SQL content that failed validation]
```

## Benefits

### 1. **Quality Assurance**
- No invalid SQL models can be saved
- Consistent coding standards enforcement
- Performance optimization guidance

### 2. **Developer Experience**
- Immediate feedback on SQL quality
- Actionable suggestions for improvement
- Prevention of runtime errors in production

### 3. **Reliability**
- Multi-layer validation prevents data corruption
- Context-aware editing preserves business logic
- Specialized agents provide domain expertise

### 4. **Maintainability**
- Clear separation of concerns
- Extensible validation rules
- Comprehensive error reporting

## Usage Examples

### Edit with Validation
```go
// User request automatically triggers full workflow
result, err := projectEditor.ProcessUserRequestWithContext(ctx, 
    "Edit the revenue model to include tax calculations")

// Result includes validation details and file save status
```

### Direct Validation
```go
// Direct validation without editing
validationResult, err := validator.ValidateModel(ctx, "model_name", sqlContent)

// Save only if validation passes
if validationResult.ValidationPassed {
    err = validator.saveModelFile(modelName, sqlContent)
}
```

## Future Enhancements

### Planned Features
- **Schema Validation**: Check against actual database schemas
- **Data Quality Rules**: Custom business logic validation
- **Performance Benchmarking**: Query execution time analysis
- **Version Control Integration**: Track validation history
- **Custom Rule Engine**: User-defined validation rules

### Extensibility
- Additional validation agents for specific domains
- Pluggable validation rule system
- Integration with external SQL analyzers
- Custom quality metrics and scoring

The validation workflow ensures that all edited models meet high standards for correctness, performance, and maintainability while providing actionable feedback to improve SQL quality.