---
title: Model Validation
sidebar_label: Model Validation
sidebar_position: 14
---

## Overview

Model validation allows you to define data quality tests that run automatically when your models refresh. Tests help ensure your data meets expectations and catch issues early in your data pipeline.

Tests are defined in your model's YAML file using the `tests:` property. Each test runs a SQL query against your model's output. If the query returns any rows, the test fails and the error is recorded in the model's state.

## When to Use Model Validation

Model validation is useful for:

- **Data Quality Checks** - Verify that your data meets business rules and constraints
- **Schema Validation** - Ensure expected columns exist and have correct data types
- **Referential Integrity** - Check relationships between tables
- **Range Validation** - Verify numeric values fall within expected ranges
- **Null Checks** - Ensure required fields don't contain null values
- **Uniqueness Constraints** - Verify primary keys and unique columns
- **Business Logic** - Validate complex business rules and calculations

## Defining Tests

Tests are defined in your model's YAML file under the `tests:` property. Each test requires:
- `name` - A descriptive name for the test
- Either `assert` or `sql` - The validation logic

### Basic Syntax

```yaml
type: model
sql: SELECT * FROM my_source

tests:
  - name: Test Name
    assert: column > 0  # OR
    sql: SELECT * FROM model WHERE condition_is_bad
```

## Test Types

### Assert Tests

Assert tests use a WHERE clause condition that should be true for all rows. If any row fails the assertion (the condition is false), the test fails.

**Syntax:**
```yaml
- name: Test Name
  assert: <condition>
```

The assertion is internally converted to:
```sql
SELECT * FROM model WHERE NOT (<condition>)
```

**Examples:**

```yaml
tests:
  # Check for null values
  - name: No Null Values
    assert: user_id IS NOT NULL

  # Range validation
  - name: Valid Age Range
    assert: age >= 0 AND age <= 120

  # Value constraints
  - name: Valid Status
    assert: status IN ('active', 'inactive', 'pending')

  # Multiple conditions
  - name: Valid Revenue
    assert: revenue >= 0 AND revenue <= 1000000
```

### SQL Tests

SQL tests use a complete SQL query that should return zero rows. If any rows are returned, the test fails.

**Syntax:**
```yaml
- name: Test Name
  sql: SELECT * FROM model WHERE <bad_condition>
```

**Examples:**

```yaml
tests:
  # Row count validation
  - name: Minimum Row Count
    sql: SELECT 'Too few rows' WHERE (SELECT COUNT(*) FROM model) < 100

  # Aggregate validation
  - name: Total Revenue Check
    sql: SELECT 'Revenue mismatch' WHERE (SELECT SUM(revenue) FROM model) < 0

  # Duplicate detection
  - name: No Duplicate Users
    sql: |
      SELECT user_id, COUNT(*) as count
      FROM model
      GROUP BY user_id
      HAVING COUNT(*) > 1

  # Referential integrity
  - name: Valid Customer References
    sql: |
      SELECT o.customer_id
      FROM model o
      LEFT JOIN customers c ON o.customer_id = c.id
      WHERE c.id IS NULL

  # Data completeness
  - name: All Expected Categories Present
    sql: |
      SELECT missing_category FROM (
        VALUES ('A'), ('B'), ('C')
      ) AS expected(missing_category)
      WHERE missing_category NOT IN (SELECT DISTINCT category FROM model)
```

## Complete Example

Here's a comprehensive example showing various validation patterns:

```yaml
# models/orders.yaml
type: model
sql: SELECT * FROM raw_orders

tests:
  # Basic null checks
  - name: Order ID Not Null
    assert: order_id IS NOT NULL

  - name: Customer ID Not Null
    assert: customer_id IS NOT NULL

  # Range validations
  - name: Valid Order Amount
    assert: amount > 0 AND amount < 1000000

  - name: Valid Order Date
    assert: order_date >= '2020-01-01' AND order_date <= CURRENT_DATE

  # Status validation
  - name: Valid Order Status
    assert: status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled')

  # Uniqueness check
  - name: No Duplicate Order IDs
    sql: |
      SELECT order_id, COUNT(*) as count
      FROM orders
      GROUP BY order_id
      HAVING COUNT(*) > 1

  # Row count validation
  - name: Minimum Orders Present
    sql: SELECT 'Too few orders' WHERE (SELECT COUNT(*) FROM orders) < 10

  # Aggregate validation
  - name: Positive Total Revenue
    sql: SELECT 'Negative revenue' WHERE (SELECT SUM(amount) FROM orders) < 0

  # Complex business logic
  - name: Shipped Orders Have Ship Date
    sql: SELECT * FROM orders WHERE status = 'shipped' AND ship_date IS NULL

  - name: Delivered Orders After Order Date
    sql: SELECT * FROM orders WHERE delivered_date < order_date

  # Referential integrity
  - name: Valid Customer References
    sql: |
      SELECT o.customer_id
      FROM orders o
      LEFT JOIN customers c ON o.customer_id = c.customer_id
      WHERE c.customer_id IS NULL
```

## Test Execution

Tests are executed automatically when your model is refreshed:

1. **Model Refresh** - The model's SQL query runs and produces output
2. **Test Execution** - Each test query runs against the model's output table
3. **Result Recording** - Test failures are recorded in the model state
4. **Model Status** - The model remains available even if tests fail

### Test Behavior

- Tests run **after** the model data is materialized
- A failing test does **not** prevent the model from being available for queries
- Test results are stored in the model's state and visible in the Rill UI
- All tests run independently - one failure doesn't stop other tests
- Tests can reference the model's output using the model name

## Viewing Test Results

Test results are stored in the model state and visible in:

- **Rill UI** - View test status in the model details page
- **Model State** - Access via the runtime API in the `test_errors` field
- **Logs** - Test failures are logged during model reconciliation

When a test fails, the error message includes:
- Test name
- Number of rows that failed the test
- Sample of the failing rows (if applicable)

## Best Practices

### Test Naming

Use descriptive names that clearly indicate what's being validated:

```yaml
# Good names
- name: No Null Customer IDs
- name: Revenue Within Valid Range
- name: All Orders Have Valid Status

# Less clear names
- name: Test 1
- name: Check Data
- name: Validation
```

### Test Organization

Group related tests together and add comments:

```yaml
tests:
  # Null checks
  - name: Order ID Not Null
    assert: order_id IS NOT NULL

  - name: Customer ID Not Null
    assert: customer_id IS NOT NULL

  # Range validations
  - name: Valid Amount
    assert: amount > 0

  - name: Valid Date Range
    assert: order_date >= '2020-01-01'
```

### Choosing Between Assert and SQL

**Use Assert when:**
- Testing row-level conditions
- Checking simple constraints
- The logic is straightforward

**Use SQL when:**
- Testing aggregate values (COUNT, SUM, AVG)
- Checking relationships between tables
- Implementing complex validation logic
- You need more control over the error message

### Performance Considerations

- Tests add time to your model refresh cycle
- Complex tests with joins or aggregations can be expensive
- Consider the trade-off between test coverage and refresh time
- Use indexes on columns referenced in test queries when possible

### Testing Strategy

**Start with critical validations:**
1. Null checks on required fields
2. Uniqueness constraints
3. Referential integrity
4. Range validations

**Add domain-specific tests:**
1. Business rules
2. Data quality checks
3. Completeness validations

**Monitor and iterate:**
1. Review test failures regularly
2. Add tests when data issues are discovered
3. Remove or update tests that are no longer relevant

## Working with Incremental Models

Tests work with incremental models and run after each incremental refresh:

```yaml
type: model
incremental: true
sql: SELECT * FROM new_data WHERE timestamp > (SELECT MAX(timestamp) FROM {{ .self }})

tests:
  - name: No Null Timestamps
    assert: timestamp IS NOT NULL

  - name: No Future Timestamps
    assert: timestamp <= CURRENT_TIMESTAMP
```

The tests validate the **entire model output**, not just the newly added rows.

## Working with Partitioned Models

Tests also work with partitioned models:

```yaml
type: model
incremental: true
partitions:
  sql: SELECT DISTINCT date FROM source_data

sql: SELECT * FROM source_data WHERE date = '{{ .partition.date }}'

tests:
  - name: Valid Dates
    assert: date IS NOT NULL

  - name: Partition Date Matches
    assert: date = '{{ .partition.date }}'
```

Tests run against the complete model output after all partitions are processed.

## Common Patterns

### Checking for Duplicates

```yaml
- name: No Duplicate Keys
  sql: |
    SELECT key, COUNT(*) as count
    FROM model
    GROUP BY key
    HAVING COUNT(*) > 1
```

### Validating Relationships

```yaml
- name: Valid Foreign Keys
  sql: |
    SELECT m.foreign_key
    FROM model m
    LEFT JOIN other_table o ON m.foreign_key = o.id
    WHERE o.id IS NULL
```

### Checking Completeness

```yaml
- name: All Expected Values Present
  sql: |
    SELECT expected_value
    FROM (VALUES ('A'), ('B'), ('C')) AS expected(expected_value)
    WHERE expected_value NOT IN (SELECT DISTINCT value FROM model)
```

### Aggregate Validations

```yaml
- name: Total Matches Expected
  sql: |
    SELECT 'Total mismatch' as error
    WHERE ABS((SELECT SUM(amount) FROM model) - 1000000) > 0.01
```

### Date Range Checks

```yaml
- name: Valid Date Range
  assert: date >= '2020-01-01' AND date <= CURRENT_DATE

- name: Dates Within Last Year
  assert: date >= CURRENT_DATE - INTERVAL '1 year'
```

## Limitations

- Tests only run when the model is refreshed
- Failing tests do not prevent the model from being queryable
- Tests cannot modify data - they are read-only validations
- Test queries should complete reasonably quickly to avoid long refresh times

## Related Topics

- [Models 101](/build/models/models-101) - Introduction to models in Rill
- [Incremental Models](/build/models/incremental-models) - Build incremental models
- [Partitioned Models](/build/models/partitioned-models) - Work with partitioned data
- [Data Refresh](/build/models/data-refresh) - Schedule model refreshes
