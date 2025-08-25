---
title: "Query-Time Dimension Joins"
description: "Use lookup dimensions to enrich your metrics view data with reference information at query time"
sidebar_label: "Query-Time Joins"
sidebar_position: 55
---

Query-time joins or more simply, lookup dimensions, provide a powerful way to incorporate reference tables into your dashboard without requiring upstream SQL transformations. This feature enables data enrichment with reference information at query time, eliminating the need for complex ETL processes.

:::note

The lookup function is dependent on the type of [OLAP engine](/connect/olap) that you are using for your project.

:::

## ClickHouse Lookups

ClickHouse provides powerful dictionary functions for lookup operations through a several different dictGet functions. Note that the table needs to be defined as a dictionary. 

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

## Druid Lookups

For those looking to add ID to name mappings with Druid (as an OLAP engine), you can utilize expressions in your **Dimension** settings. Simply use the lookup function and provide the name of the lookup and ID, i.e. `lookup(city_id, 'cities')`. Be sure to include the lookup table name in single quotes.

```yaml
dimensions:
  - name: city_name
    display_name: "Cities"
    expression: lookup(city_id, 'cities')
    description: "City names from lookup table"
```

## DuckDB/MotherDuck Lookups

Unfortunately, DuckDB does not have a dimension lookup **function** and instead relies on upstream modeling to join the lookup table and create a [one big table](/build/models/#one-big-table-and-dashboarding) for dashboarding. Take a look through [DuckDB docs](https://duckdb.org/docs/stable/sql/introduction) for further information!

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