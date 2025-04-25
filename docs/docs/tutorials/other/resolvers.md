---
sidebar_label: "Rill's Resolvers"
sidebar_position: 10
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
---

# Understanding Resolvers in Rill

Resolvers are a fundamental concept in Rill that represent logic that produces output data. They're used to evaluate [API requests](/integrate/custom-apis/), [alerts](/explore/alerts/), and other data-driven features in your application. This tutorial will explain what resolvers are and how to use the three main types in Rill: SQL, metrics_sql, and API resolvers.

## What are Resolvers?

A resolver has two levels of configuration:
1. **Static properties** - Usually declared in advance (e.g., in YAML for a custom API)
2. **Dynamic arguments** - Provided just prior to execution (e.g., in an API request)

For example, a SQL resolver has a static property for the SQL query template and dynamic arguments for query parameters injected at runtime.

## SQL Resolvers

SQL resolvers execute SQL queries against your data connectors. They're perfect for custom queries that don't fit the metrics model or when you need more complex data transformations.

### Example Usage

```yaml
# Define a custom API with a SQL resolver
type: API

sql: |
  SELECT 
    id, 
    name, 
    email, 
    created_at 
  FROM users 
    WHERE id = '{{args.user_id}}'
```

When calling this API, you would pass `user_id` as an argument:

```
.../runtime/api/my-api?user_id=abc
```

The SQL resolver supports templating in the SQL string to inject user attributes and args into the query. This makes it powerful for creating dynamic, parameterized queries.

## Metrics SQL Resolvers

Metrics SQL resolvers are specialized SQL resolvers for working with metrics views. They provide a way to write SQL against your metrics data while preserving the time-series nature of your metrics.

### Example Usage

```yaml
type: API

metrics_sql: |
  SELECT 
    dom AS domain,
    SUM(revenue) AS total_revenue
  FROM "ad_bids"
  WHERE pub = {{args.publisher}}
  GROUP BY 1
  ORDER BY 2 DESC
```

When calling this API, you would pass `publisher` as an argument:

```
.../runtime/api/my-api?publisher=abc
```

## API Resolvers

API resolvers act as proxies to other resolvers, allowing you to create reusable components that can be chained together. They're useful for creating abstraction layers in your data architecture.

### Example Usage

```yaml
type: API

api: user_details # References another API
args:
  user_id: 123    # Static args will override user-provided args
```

This example creates a new API that calls an existing `user_details` API with pre-defined arguments. This is useful for creating specialized views of your data or for composing complex queries from simpler components.

## Alerts! 

The examples show various ways to use these resolvers in APIs, but these can also be used in other components of Rill such as Alerts. Similar to how you define the resolver in an API, in alerts, you would do so under the `data` parameter.

```
data:
  metrics_sql: select measure from metrics_view
  #api: name_of_api
  #sql: select column from table
```

## Security Considerations

It's important to understand the security implications when working with different resolver types:

### Security Policy Scope
While Rill allows you to define [security policies](/manage/security) on your metrics views to grant access to certain groups, users, or based on a mapping file, these security policies **only apply to metrics resolvers**. When using SQL or other resolver types in APIs, these security policies do not automatically extend to protect your data.

### Resolver-Specific Security Implications

1. **SQL Resolvers**: 
   - Have no inherent access control beyond what you implement
   - Can expose raw data from any table the connector has access to
   - Should be carefully constructed to prevent SQL injection via proper parameter handling

2. **Metrics SQL Resolvers**: 
   - Benefit from metrics view security policies
   - Still need careful management of SQL injection risks 
   - Admin-only restriction applies to arbitrary queries via `/metrics-sql` API

3. **API Resolvers**: 
   - Inherit security characteristics of the underlying resolver they call
   - Include protection against infinite recursion loops
   - Do not add additional security on top of the referenced API

### Best Practices

- Validate and sanitize all user inputs before using them in resolvers
- Implement authorization checks in your API handlers for non-metrics resolvers
- Consider using API resolvers to wrap SQL resolvers with additional security checks
- Avoid exposing sensitive data in error messages
- Use parameterized queries rather than direct string interpolation

## Conclusion

Resolvers are a powerful abstraction in Rill that let you define how data should be processed and retrieved. By combining SQL, metrics_sql, and API resolvers, you can create flexible data pipelines that meet your specific needs while maintaining security and performance.

For more information, consult the Rill API reference documentation or experiment with the built-in APIs in your Rill instance. 
