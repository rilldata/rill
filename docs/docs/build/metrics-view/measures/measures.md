---
title: "Define Your Measures"
description: "Learn how to create and configure measures for quantitative analysis and business insights in your metrics views"
---

Measures are the quantitative metrics that power your dashboards and reports. They represent numeric calculations derived from your underlying data through SQL aggregation functions and expressions. These measures transform raw data into meaningful business insights, such as total revenue, average order value, or customer count.

Measures are the "how much" and "how many" of your data. They provide the numerical foundation for your analysis, enabling you to:

- **Quantify performance**: Track key business metrics like revenue, growth, and efficiency
- **Support decision-making**: Provide concrete numbers for business decisions


## Supported SQL Functions

* Standard SQL [numeric](https://www.w3schools.com/sql/sql_operators.asp) operators and functions
* Common SQL [aggregates](https://www.w3schools.com/sql/sql_aggregate_functions.asp): `AVG`, `COUNT`, `MAX`, `MIN`, `SUM`, `STDDEV`, `VARIANCE`
* Advanced aggregates (engine-dependent): `APPROX_COUNT_DISTINCT`, `APPROX_QUANTILE`, `STDDEV_POP`, `STDDEV_SAMP`, `VAR_POP`, `VAR_SAMP`
* [Filtered aggregates](https://www.postgresql.org/docs/current/sql-expressions.html#SYNTAX-AGGREGATES) can be used to filter the set of rows fed to the aggregate functions (syntax may vary by engine)

:::info Engine-specific SQL dialects

Different OLAP engines support varying SQL dialects and functions. While standard SQL functions work across engines, some advanced features may be engine-specific. For engine-specific documentation, see:

- **DuckDB**: [DuckDB SQL documentation](https://duckdb.org/docs/sql/introduction.html)
- **ClickHouse**: [ClickHouse SQL documentation](https://clickhouse.com/docs/en/sql-reference)
- **Druid**: [Druid SQL documentation](https://druid.apache.org/docs/latest/querying/sql.html)
- **Pinot**: [Pinot SQL documentation](https://docs.pinot.apache.org/users/user-guide-query/pinot-query-language)

:::

As an example, if you have a table of sales events with the sales price and customer ID, you could calculate the following measures with these aggregates and expressions:
* Number of sales: `COUNT(*)`
* Total revenue: `SUM(sales_price)` 
* Revenue per customer: `CAST(SUM(sales_price) AS FLOAT)/CAST(COUNT(DISTINCT customer_id) AS FLOAT)`
* Number of orders with order value more than $100: `COUNT(*) FILTER (WHERE order_val > 100)` (syntax may vary by engine)


Explore these advanced capabilities to enhance your measures:

- **[Measure Formatting](/build/metrics-view/measures/measures-formatting)** - Learn how to format and display your measures effectively
- **[Case Statements and Filters](/build/metrics-view/measures/case-statements)** - Use conditional logic and filtering in your measures
- **[Referencing Measures](/build/metrics-view/measures/referencing)** - Reference and combine existing measures in your calculations
- **[Quantiles](/build/metrics-view/measures/quantiles)** - Calculate percentiles and quantiles for statistical analysis
- **[Fixed Measures](/build/metrics-view/measures/fixed-measures)** - Create measures with fixed values and constants
- **[Window Functions](/build/metrics-view/measures/windows)** - Apply window functions for advanced analytical operations
