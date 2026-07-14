---
title: "Query-Time Dimension Joins"
description: "Use lookup dimensions to enrich your metrics view data with reference information at query time"
sidebar_label: "Query-Time Joins"
sidebar_position: 55
---

Query-time joins or more simply, lookup dimensions, provide a powerful way to incorporate reference tables into your dashboard without requiring upstream SQL transformations. This feature enables data enrichment with reference information at query time, eliminating the need for complex ETL processes.

:::note

The lookup function is dependent on the type of [OLAP engine](/developers/build/connectors/olap) that you are using for your project.

:::

## ClickHouse Lookups

ClickHouse provides powerful dictionary functions for lookup operations through `dictGet` functions. The table needs to be defined as a dictionary.

### Using `expression`

You can use `dictGet` directly in a dimension's `expression`:

```yaml
dimensions:
  - name: user_email
    display_name: "User Email"
    # Query-time lookup in the 'users' dictionary
    expression: dictGet('users', 'email', user_id)
    description: "User email from users dictionary"
    
  - name: product_category
    display_name: "Product Category" 
    expression: dictGet('products', 'category', product_id)
    description: "Product category information"
    
  # With default fallback
  - name: user_email_safe
    display_name: "User Email (Safe)"
    expression: dictGetOrDefault('users', 'email', user_id, 'unknown@example.com')
    description: "User email with fallback for missing users"
    
  # Multiple attributes from same dictionary
  - name: user_full_name
    display_name: "User Full Name"
    expression: concat(dictGet('users', 'first_name', user_id), ' ', dictGet('users', 'last_name', user_id))
    description: "Concatenated full name from user dictionary"
```

This approach is straightforward but has a performance drawback: when filtering by a lookup dimension, ClickHouse must evaluate `dictGet` for every row before applying the filter, which prevents it from using indexes on the key column.

### Using `lookup_*` properties

The `lookup_*` properties provide a declarative way to define lookup dimensions. Rill automatically rewrites filter queries to use an optimized subquery pattern, enabling ClickHouse to leverage indexes on key columns.

```yaml
dimensions:
  - name: country_name
    display_name: "Country"
    # column in the fact table to match with the dictionary key
    column: country_code
    lookup_table: country_dict
    lookup_key_column: code
    lookup_value_column: name

  # With a default value for missing keys
  - name: region
    display_name: "Region"
    # column in the fact table to match with the dictionary key
    column: country_code
    lookup_table: country_dict
    lookup_key_column: code
    lookup_value_column: region
    lookup_default_expression: "'Unknown'"
```

The properties are:
- **`lookup_table`** — the dictionary name. Use `database.dictionary_name` if the dictionary is in a non-default database (e.g. `mydb.my_dict`).
- **`lookup_key_column`** — the primary key column in the dictionary that matches the fact table's `column`.
- **`lookup_value_column`** — the attribute column in the dictionary to display.
- **`lookup_default_expression`** _(optional)_ — a SQL expression used as a fallback when no match is found (maps to `dictGetOrDefault`).

#### How the query optimization works

When you filter on a lookup dimension defined with `expression: dictGet(...)`, the generated query looks like:

```sql
SELECT ... FROM fact_table
WHERE dictGet('dict', 'name', publisher_id) IN ('Alice', 'Bob')
```

ClickHouse must call `dictGet` on every row to evaluate the filter, making it unable to use the index on `publisher_id`.

With `lookup_*` properties, Rill rewrites the filter to a subquery pattern:

```sql
SELECT ... FROM fact_table
WHERE publisher_id IN (
  SELECT code FROM dict WHERE name IN ('Alice', 'Bob')
)
```

This allows ClickHouse to first resolve the small set of matching keys from the dictionary, then use the index on `publisher_id` to efficiently scan only the relevant rows in the fact table.

:::tip Optimizing GROUP BY with INJECTIVE
If the key-to-value mapping in your dictionary is 1:1 (e.g. every `code` maps to exactly one `name`), you can additionally mark the attribute columns as `INJECTIVE` in the ClickHouse dictionary definition. This tells ClickHouse that the mapping preserves uniqueness, allowing it to optimize `GROUP BY` queries by grouping on the key column instead of evaluating `dictGet` for every group. Combined with `lookup_*` properties, this gives you optimized performance for both filtering (via Rill's subquery rewrite) and grouping (via ClickHouse's INJECTIVE optimization).
:::

## Druid Lookups

For those looking to add ID to name mappings with Druid (as an OLAP engine), you can utilize expressions in your **Dimension** settings. Simply use the `lookup` function and provide the name of the lookup and ID, i.e. `lookup(city_id, 'cities')`. Be sure to include the lookup table name in single quotes.

```yaml
dimensions:
  - name: city_name
    display_name: "Cities"
    expression: lookup(city_id, 'cities')
    description: "City names from lookup table"
```

## DuckDB/MotherDuck Lookups

Unfortunately, DuckDB does not have a dimension lookup **function** and instead relies on upstream modeling to join the lookup table and create a [one big table](/developers/build/models/models-101#one-big-table-and-dashboarding) for dashboarding. Take a look through [DuckDB docs](https://duckdb.org/docs/stable/sql/introduction) for further information!

```sql
SELECT 
    o.*,
    u.user_name
FROM orders o
LEFT JOIN users u ON u.email = o.email;
```


Alternatively, you can use DuckDB's `map` function to create a lookup table by mapping values from one column to another. This approach creates an in-memory mapping that can be referenced without using a joining SQL:

```yaml
 - expression: (SELECT map(list(email), list(user_name)) FROM users_dataset)[email]
    name: user_name
```


## Performance Implications

While joins can kill the performance of [OLAP engines](/developers/build/connectors/olap), lookups (key-value pairs) are common to reduce data size and improve query speeds. Lookups can be done during ingestion time (a static lookup to enrich the source data) or at query time (dynamic lookups).

**Static Lookups** 

Static lookups are lookups that are ingested at processing time. When a record is being processed, if a match is found between the record and lookup's key, the lookup's corresponding value at that moment in time is extracted and carbon-copied into your OLAP engine for the records it processed.

Static lookups are best suited for:

- Dimensions with values that require a historical record for how they have changed over time
- Values that are never expected to change (leverage Dynamic Lookups if the values are expected to change)
- Extremely large lookups (hundreds of thousands of records or >50MB lookup file) to improve query performance

Customers typically store lookup values in S3 or GCS, and the lookup file is then updated by customers as needed and consumed by ETL logic.

**Dynamic Lookups**

Since static lookups transform and store the data permanently, any changes to the mapping would require reprocessing the entire dataset to ensure consistency. To address the case when values in a lookup are expected to change with time, we developed dynamic lookups. Dynamic lookups, also known as Query Time Lookups, are lookups that are retrieved at query time, as opposed to being used at ingestion time.

Benefits of dynamic lookups include:

- Historical continuity for dimensions that change frequently without reprocessing the entire dataset
- Time savings, because there is no dataset reprocessing required to complete the update
- Dynamic lookups are kept separate from the dataset. Thus, any human errors introduced in the lookup do not impact the underlying dataset
- Ability for users to create new dimension tables from metadata associated with a dimension table. For example, account ownership can change during the course of a quarter. In such cases, a dynamic lookup ensures that the most current ownership information is always available without requiring data reprocessing.