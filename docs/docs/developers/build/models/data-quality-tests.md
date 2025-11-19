---
title: Data Quality Tests
sidebar_label: Data Quality Tests
sidebar_position: 14
---

## Overview

Data quality tests allow you to define automated checks that run whenever your models refresh. These tests ensure your data meets expectations and catch issues early in your pipeline.

Tests are defined in your model's YAML file using the `tests:` property. Each test runs a SQL query against your model's output. If the query returns any rows, the test fails and the error is recorded in the model's state.

## When to Use Data Quality Tests

Data quality tests are useful for:

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
  - name: No Null Campaign ID
    assert: campaign_id IS NOT NULL

  # Range validation
  - name: Valid Bid Price
    assert: bid_price >= 0 AND bid_price <= 100

  # Value constraints
  - name: Valid Ad Status
    assert: status IN ('active', 'paused', 'completed', 'draft')

  # Multiple conditions
  - name: Valid Impression Count
    assert: impressions >= 0 AND impressions <= 1000000000
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
  - name: Minimum Impression Count
    sql: SELECT 'Too few impressions' WHERE (SELECT COUNT(*) FROM model) < 1000

  # Aggregate validation
  - name: Positive Total Spend
    sql: SELECT 'Negative spend detected' WHERE (SELECT SUM(spend) FROM model) < 0

  # Duplicate detection
  - name: No Duplicate Impression IDs
    sql: |
      SELECT impression_id, COUNT(*) as count
      FROM model
      GROUP BY impression_id
      HAVING COUNT(*) > 1

  # Referential integrity
  - name: Valid Campaign References
    sql: |
      SELECT i.campaign_id
      FROM model i
      LEFT JOIN campaigns c ON i.campaign_id = c.id
      WHERE c.id IS NULL

  # Data completeness
  - name: All Expected Ad Formats Present
    sql: |
      SELECT missing_format FROM (
        VALUES ('banner'), ('video'), ('native')
      ) AS expected(missing_format)
      WHERE missing_format NOT IN (SELECT DISTINCT ad_format FROM model)
```

## Complete Example

Here's a comprehensive example showing various validation patterns:

```yaml
# models/ad_impressions.yaml
type: model
sql: SELECT * FROM raw_impressions

tests:
  # Basic null checks
  - name: Impression ID Not Null
    assert: impression_id IS NOT NULL

  - name: Campaign ID Not Null
    assert: campaign_id IS NOT NULL

  # Range validations
  - name: Valid Bid Price
    assert: bid_price > 0 AND bid_price < 100

  - name: Valid Impression Date
    assert: impression_date >= '2020-01-01' AND impression_date <= CURRENT_DATE

  # Status validation
  - name: Valid Campaign Status
    assert: status IN ('active', 'paused', 'completed', 'draft')

  # Uniqueness check
  - name: No Duplicate Impression IDs
    sql: |
      SELECT impression_id, COUNT(*) as count
      FROM ad_impressions
      GROUP BY impression_id
      HAVING COUNT(*) > 1

  # Row count validation
  - name: Minimum Impressions Present
    sql: SELECT 'Too few impressions' WHERE (SELECT COUNT(*) FROM ad_impressions) < 1000

  # Aggregate validation
  - name: Positive Total Spend
    sql: SELECT 'Negative spend' WHERE (SELECT SUM(spend) FROM ad_impressions) < 0

  # Complex business logic
  - name: Clicks Must Have Impressions
    sql: SELECT * FROM ad_impressions WHERE clicks > 0 AND impressions = 0

  - name: Click Timestamp After Impression
    sql: SELECT * FROM ad_impressions WHERE click_timestamp < impression_timestamp

  # Referential integrity
  - name: Valid Campaign References
    sql: |
      SELECT i.campaign_id
      FROM ad_impressions i
      LEFT JOIN campaigns c ON i.campaign_id = c.campaign_id
      WHERE c.campaign_id IS NULL
```

## Test Execution

Tests are executed automatically when your model is refreshed:

1. **Model Refresh** - The model's SQL query runs and produces output
2. **Test Execution** - Each test query runs against the model's output table
3. **Result Recording** - Test failures are recorded in the model state
4. **Model Status** - The model remains available even if tests fail

### Test Behavior

- Tests run **after** successful model refresh
- A failing test does **not** prevent the model from being available for queries
- Test results are stored in the model's state and visible in the [Rill logs](/reference/cli/project/logs)
- All tests run independently - one failure doesn't stop other tests
- Tests can reference the model's output using the model name

## Viewing Test Results

Test results are stored in the model state and visible in:

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
- name: No Null Campaign IDs
- name: Spend Within Valid Range
- name: All Impressions Have Valid Status

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
  - name: Impression ID Not Null
    assert: impression_id IS NOT NULL

  - name: Campaign ID Not Null
    assert: campaign_id IS NOT NULL

  # Range validations
  - name: Valid Bid Price
    assert: bid_price > 0

  - name: Valid Date Range
    assert: impression_date >= '2020-01-01'
```

### Understanding Assert vs SQL Syntax

**Assert Syntax** - Define conditions that should be true for all rows:
- You write: `assert: value > 0`
- Rill converts this to: `SELECT * FROM model WHERE NOT (value > 0)`
- Tests **pass** if no rows are returned (all rows satisfy the condition)
- Tests **fail** if any rows are returned (violations found)

**SQL Syntax** - Write custom queries that return failing rows:
- You write: `sql: SELECT * FROM model WHERE value <= 0`
- Your query should explicitly return rows that represent failures
- Tests **pass** if the query returns an empty result set
- Tests **fail** if the query returns any rows

:::tip Key Difference
With `assert`, you define what should be **true**. With `sql`, you query for what is **wrong**.
:::

### Choosing Between Assert and SQL

**Use Assert when:**
- Testing row-level conditions
- Checking simple constraints
- The logic is straightforward
- You want Rill to handle the "NOT" logic for you

**Use SQL when:**
- Testing aggregate values (COUNT, SUM, AVG)
- Checking relationships between tables
- Implementing complex validation logic
- You need more control over the error message
- You prefer to explicitly write the failure query

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
sql: SELECT * FROM raw_impressions WHERE impression_timestamp > (SELECT MAX(impression_timestamp) FROM {{ .self }})

tests:
  - name: No Null Impression Timestamps
    assert: impression_timestamp IS NOT NULL

  - name: No Future Impression Timestamps
    assert: impression_timestamp <= CURRENT_TIMESTAMP
```

The tests validate the **entire model output**, not just the newly added rows.

## Working with Partitioned Models

Tests also work with partitioned models:

```yaml
type: model
incremental: true
partitions:
  sql: SELECT DISTINCT impression_date FROM raw_impressions

sql: SELECT * FROM raw_impressions WHERE impression_date = '{{ .partition.impression_date }}'

tests:
  - name: Valid Impression Dates
    assert: impression_date IS NOT NULL

  - name: Partition Date Matches
    assert: impression_date = '{{ .partition.impression_date }}'
```

Tests run against the complete model output after all partitions are processed.

## Examples

### Checking for Duplicates

```yaml
- name: No Duplicate Impression IDs
  sql: |
    SELECT impression_id, COUNT(*) as count
    FROM model
    GROUP BY impression_id
    HAVING COUNT(*) > 1
```

### Validating Relationships

```yaml
- name: Valid Campaign References
  sql: |
    SELECT i.campaign_id
    FROM model i
    LEFT JOIN campaigns c ON i.campaign_id = c.id
    WHERE c.id IS NULL
```

### Checking Completeness

```yaml
- name: All Expected Ad Formats Present
  sql: |
    SELECT expected_format
    FROM (VALUES ('banner'), ('video'), ('native')) AS expected(expected_format)
    WHERE expected_format NOT IN (SELECT DISTINCT ad_format FROM model)
```

### Aggregate Validations

```yaml
- name: CTR Within Expected Range
  sql: |
    SELECT 'CTR out of range' as error
    WHERE (SELECT SUM(clicks) * 1.0 / NULLIF(SUM(impressions), 0) FROM model) > 0.5
```

### Date Range Checks

```yaml
- name: Valid Impression Date Range
  assert: impression_date >= '2020-01-01' AND impression_date <= CURRENT_DATE

- name: Impressions Within Last Year
  assert: impression_date >= CURRENT_DATE - INTERVAL '1 year'
```

:::warning Limitations

- Tests only run when the model is refreshed
- Failing tests do not prevent the model from being queryable
- Tests cannot modify data - they are read-only validations
- Test queries should complete reasonably quickly to avoid long refresh times

:::