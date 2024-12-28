# \QueryServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**query_service_column_cardinality**](QueryServiceApi.md#query_service_column_cardinality) | **GET** /v1/instances/{instanceId}/queries/column-cardinality/tables/{tableName} | Get cardinality for a column
[**query_service_column_descriptive_statistics**](QueryServiceApi.md#query_service_column_descriptive_statistics) | **GET** /v1/instances/{instanceId}/queries/descriptive-statistics/tables/{tableName} | Get basic stats for a numeric column like min, max, mean, stddev, etc
[**query_service_column_null_count**](QueryServiceApi.md#query_service_column_null_count) | **GET** /v1/instances/{instanceId}/queries/null-count/tables/{tableName} | Get the number of nulls in a column
[**query_service_column_numeric_histogram**](QueryServiceApi.md#query_service_column_numeric_histogram) | **GET** /v1/instances/{instanceId}/queries/numeric-histogram/tables/{tableName} | Get the histogram for values in a column
[**query_service_column_rollup_interval**](QueryServiceApi.md#query_service_column_rollup_interval) | **POST** /v1/instances/{instanceId}/queries/rollup-interval/tables/{tableName} | ColumnRollupInterval returns the minimum time granularity (as well as the time range) for a specified timestamp column
[**query_service_column_rug_histogram**](QueryServiceApi.md#query_service_column_rug_histogram) | **GET** /v1/instances/{instanceId}/queries/rug-histogram/tables/{tableName} | Get outliers for a numeric column
[**query_service_column_time_grain**](QueryServiceApi.md#query_service_column_time_grain) | **GET** /v1/instances/{instanceId}/queries/smallest-time-grain/tables/{tableName} | Estimates the smallest time grain present in the column
[**query_service_column_time_range**](QueryServiceApi.md#query_service_column_time_range) | **GET** /v1/instances/{instanceId}/queries/time-range-summary/tables/{tableName} | Get the time range summaries (min, max) for a column
[**query_service_column_time_series**](QueryServiceApi.md#query_service_column_time_series) | **POST** /v1/instances/{instanceId}/queries/timeseries/tables/{tableName} | Generate time series for the given measures (aggregation expressions) along with the sparkline timeseries
[**query_service_column_top_k**](QueryServiceApi.md#query_service_column_top_k) | **POST** /v1/instances/{instanceId}/queries/topk/tables/{tableName} | Get TopK elements from a table for a column given an agg function agg function and k are optional, defaults are count(*) and 50 respectively
[**query_service_export**](QueryServiceApi.md#query_service_export) | **POST** /v1/instances/{instanceId}/queries/export | Export builds a URL to download the results of a query as a file.
[**query_service_export_report**](QueryServiceApi.md#query_service_export_report) | **POST** /v1/instances/{instanceId}/reports/{report}/export | ExportReport builds a URL to download the results of a query as a file.
[**query_service_metrics_view_aggregation**](QueryServiceApi.md#query_service_metrics_view_aggregation) | **POST** /v1/instances/{instanceId}/queries/metrics-views/{metricsView}/aggregation | MetricsViewAggregation is a generic API for running group-by/pivot queries against a metrics view.
[**query_service_metrics_view_comparison**](QueryServiceApi.md#query_service_metrics_view_comparison) | **POST** /v1/instances/{instanceId}/queries/metrics-views/{metricsViewName}/compare-toplist | MetricsViewComparison returns a toplist containing comparison data of another toplist (same dimension/measure but a different time range). Returns a toplist without comparison data if comparison time range is omitted.
[**query_service_metrics_view_rows**](QueryServiceApi.md#query_service_metrics_view_rows) | **POST** /v1/instances/{instanceId}/queries/metrics-views/{metricsViewName}/rows | MetricsViewRows returns the underlying model rows matching a metrics view time range and filter(s).
[**query_service_metrics_view_schema**](QueryServiceApi.md#query_service_metrics_view_schema) | **GET** /v1/instances/{instanceId}/queries/metrics-views/{metricsViewName}/schema | MetricsViewSchema Get the data types of measures and dimensions
[**query_service_metrics_view_search**](QueryServiceApi.md#query_service_metrics_view_search) | **POST** /v1/instances/{instanceId}/queries/metrics-views/{metricsViewName}/search | MetricsViewSearch Get the data types of measures and dimensions
[**query_service_metrics_view_time_range**](QueryServiceApi.md#query_service_metrics_view_time_range) | **POST** /v1/instances/{instanceId}/queries/metrics-views/{metricsViewName}/time-range-summary | MetricsViewTimeRange Get the time range summaries (min, max) for time column in a metrics view
[**query_service_metrics_view_time_series**](QueryServiceApi.md#query_service_metrics_view_time_series) | **POST** /v1/instances/{instanceId}/queries/metrics-views/{metricsViewName}/timeseries | MetricsViewTimeSeries returns time series for the measures in the metrics view. It's a convenience API for querying a metrics view.
[**query_service_metrics_view_toplist**](QueryServiceApi.md#query_service_metrics_view_toplist) | **POST** /v1/instances/{instanceId}/queries/metrics-views/{metricsViewName}/toplist | Deprecated - use MetricsViewComparison instead. MetricsViewToplist returns the top dimension values of a metrics view sorted by one or more measures. It's a convenience API for querying a metrics view.
[**query_service_metrics_view_totals**](QueryServiceApi.md#query_service_metrics_view_totals) | **POST** /v1/instances/{instanceId}/queries/metrics-views/{metricsViewName}/totals | MetricsViewTotals returns totals over a time period for the measures in a metrics view. It's a convenience API for querying a metrics view.
[**query_service_query**](QueryServiceApi.md#query_service_query) | **POST** /v1/instances/{instanceId}/query | Query runs a SQL query against the instance's OLAP datastore.
[**query_service_query_batch**](QueryServiceApi.md#query_service_query_batch) | **POST** /v1/instances/{instanceId}/query/batch | Batch request with different queries
[**query_service_resolve_component**](QueryServiceApi.md#query_service_resolve_component) | **POST** /v1/instances/{instanceId}/queries/components/{component}/resolve | ResolveComponent resolves the data and renderer for a Component resource.
[**query_service_table_cardinality**](QueryServiceApi.md#query_service_table_cardinality) | **GET** /v1/instances/{instanceId}/queries/table-cardinality/tables/{tableName} | TableCardinality returns row count
[**query_service_table_columns**](QueryServiceApi.md#query_service_table_columns) | **POST** /v1/instances/{instanceId}/queries/columns-profile/tables/{tableName} | TableColumns returns column profiles
[**query_service_table_rows**](QueryServiceApi.md#query_service_table_rows) | **GET** /v1/instances/{instanceId}/queries/rows/tables/{tableName} | TableRows returns table rows



## query_service_column_cardinality

> models::V1ColumnCardinalityResponse query_service_column_cardinality(instance_id, table_name, connector, database, database_schema, column_name, priority)
Get cardinality for a column

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** | Required | [required] |
**table_name** | **String** | Required | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**column_name** | Option<**String**> | Required |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1ColumnCardinalityResponse**](v1ColumnCardinalityResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_descriptive_statistics

> models::V1ColumnDescriptiveStatisticsResponse query_service_column_descriptive_statistics(instance_id, table_name, connector, database, database_schema, column_name, priority)
Get basic stats for a numeric column like min, max, mean, stddev, etc

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** | Required | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**column_name** | Option<**String**> | Required |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1ColumnDescriptiveStatisticsResponse**](v1ColumnDescriptiveStatisticsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_null_count

> models::V1ColumnNullCountResponse query_service_column_null_count(instance_id, table_name, connector, database, database_schema, column_name, priority)
Get the number of nulls in a column

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** | Required | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**column_name** | Option<**String**> | Required |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1ColumnNullCountResponse**](v1ColumnNullCountResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_numeric_histogram

> models::V1ColumnNumericHistogramResponse query_service_column_numeric_histogram(instance_id, table_name, connector, database, database_schema, column_name, histogram_method, priority)
Get the histogram for values in a column

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** |  | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**column_name** | Option<**String**> |  |  |
**histogram_method** | Option<**String**> |  |  |[default to HISTOGRAM_METHOD_UNSPECIFIED]
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1ColumnNumericHistogramResponse**](v1ColumnNumericHistogramResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_rollup_interval

> models::V1ColumnRollupIntervalResponse query_service_column_rollup_interval(instance_id, table_name, body)
ColumnRollupInterval returns the minimum time granularity (as well as the time range) for a specified timestamp column

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** | Required | [required] |
**body** | [**QueryServiceColumnRollupIntervalRequest**](QueryServiceColumnRollupIntervalRequest.md) |  | [required] |

### Return type

[**models::V1ColumnRollupIntervalResponse**](v1ColumnRollupIntervalResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_rug_histogram

> models::V1ColumnRugHistogramResponse query_service_column_rug_histogram(instance_id, table_name, connector, database, database_schema, column_name, priority)
Get outliers for a numeric column

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** |  | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**column_name** | Option<**String**> |  |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1ColumnRugHistogramResponse**](v1ColumnRugHistogramResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_time_grain

> models::V1ColumnTimeGrainResponse query_service_column_time_grain(instance_id, table_name, connector, database, database_schema, column_name, priority)
Estimates the smallest time grain present in the column

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** | Required | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**column_name** | Option<**String**> | Required |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1ColumnTimeGrainResponse**](v1ColumnTimeGrainResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_time_range

> models::V1ColumnTimeRangeResponse query_service_column_time_range(instance_id, table_name, connector, database, database_schema, column_name, priority)
Get the time range summaries (min, max) for a column

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** |  | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**column_name** | Option<**String**> |  |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1ColumnTimeRangeResponse**](v1ColumnTimeRangeResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_time_series

> models::V1ColumnTimeSeriesResponse query_service_column_time_series(instance_id, table_name, body)
Generate time series for the given measures (aggregation expressions) along with the sparkline timeseries

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** | Required | [required] |
**body** | [**QueryServiceColumnTimeSeriesRequest**](QueryServiceColumnTimeSeriesRequest.md) |  | [required] |

### Return type

[**models::V1ColumnTimeSeriesResponse**](v1ColumnTimeSeriesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_column_top_k

> models::V1ColumnTopKResponse query_service_column_top_k(instance_id, table_name, body)
Get TopK elements from a table for a column given an agg function agg function and k are optional, defaults are count(*) and 50 respectively

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** | Required | [required] |
**body** | [**QueryServiceColumnTopKRequest**](QueryServiceColumnTopKRequest.md) |  | [required] |

### Return type

[**models::V1ColumnTopKResponse**](v1ColumnTopKResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_export

> models::V1ExportResponse query_service_export(instance_id, body)
Export builds a URL to download the results of a query as a file.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**QueryServiceExportRequest**](QueryServiceExportRequest.md) |  | [required] |

### Return type

[**models::V1ExportResponse**](v1ExportResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_export_report

> models::V1ExportReportResponse query_service_export_report(instance_id, report, body)
ExportReport builds a URL to download the results of a query as a file.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**report** | **String** |  | [required] |
**body** | [**QueryServiceExportReportRequest**](QueryServiceExportReportRequest.md) |  | [required] |

### Return type

[**models::V1ExportReportResponse**](v1ExportReportResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_aggregation

> models::V1MetricsViewAggregationResponse query_service_metrics_view_aggregation(instance_id, metrics_view, body)
MetricsViewAggregation is a generic API for running group-by/pivot queries against a metrics view.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view** | **String** | Required | [required] |
**body** | [**QueryServiceMetricsViewAggregationRequest**](QueryServiceMetricsViewAggregationRequest.md) |  | [required] |

### Return type

[**models::V1MetricsViewAggregationResponse**](v1MetricsViewAggregationResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_comparison

> models::V1MetricsViewComparisonResponse query_service_metrics_view_comparison(instance_id, metrics_view_name, body)
MetricsViewComparison returns a toplist containing comparison data of another toplist (same dimension/measure but a different time range). Returns a toplist without comparison data if comparison time range is omitted.

ie. comparsion toplist: | measure1_base | measure1_previous   | measure1__delta_abs | measure1__delta_rel | dimension | |---------------|---------------------|---------------------|--------------------|-----------| | 2             | 2                   | 0                   | 0                  | Safari    | | 1             | 0                   | 1                   | N/A                | Chrome    | | 0             | 4                   | -4                  | -1.0               | Firefox   |  ie. toplist: | measure1 | measure2 | dimension | |----------|----------|-----------| | 2        | 45       | Safari    | | 1        | 350      | Chrome    | | 0        | 25       | Firefox   |

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view_name** | **String** |  | [required] |
**body** | [**RequestMessageForQueryServiceMetricsViewComparison**](RequestMessageForQueryServiceMetricsViewComparison.md) |  | [required] |

### Return type

[**models::V1MetricsViewComparisonResponse**](v1MetricsViewComparisonResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_rows

> models::V1MetricsViewRowsResponse query_service_metrics_view_rows(instance_id, metrics_view_name, body)
MetricsViewRows returns the underlying model rows matching a metrics view time range and filter(s).

ie. without granularity | column1 | column2 | dimension | |---------|---------|-----------| | 2       | 2       | Safari    | | 1       | 0       | Chrome    | | 0       | 4       | Firefox   |  ie. with granularity | timestamp__day0      | column1 | column2 | dimension | |----------------------|---------|---------|-----------| | 2022-01-01T00:00:00Z | 2       | 2       | Safari    | | 2022-01-01T00:00:00Z | 1       | 0       | Chrome    | | 2022-01-01T00:00:00Z | 0       | 4       | Firefox   |

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view_name** | **String** |  | [required] |
**body** | [**QueryServiceMetricsViewRowsRequest**](QueryServiceMetricsViewRowsRequest.md) |  | [required] |

### Return type

[**models::V1MetricsViewRowsResponse**](v1MetricsViewRowsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_schema

> models::V1MetricsViewSchemaResponse query_service_metrics_view_schema(instance_id, metrics_view_name, priority)
MetricsViewSchema Get the data types of measures and dimensions

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view_name** | **String** |  | [required] |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1MetricsViewSchemaResponse**](v1MetricsViewSchemaResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_search

> models::V1MetricsViewSearchResponse query_service_metrics_view_search(instance_id, metrics_view_name, body)
MetricsViewSearch Get the data types of measures and dimensions

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view_name** | **String** |  | [required] |
**body** | [**QueryServiceMetricsViewSearchRequest**](QueryServiceMetricsViewSearchRequest.md) |  | [required] |

### Return type

[**models::V1MetricsViewSearchResponse**](v1MetricsViewSearchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_time_range

> models::V1MetricsViewTimeRangeResponse query_service_metrics_view_time_range(instance_id, metrics_view_name, body)
MetricsViewTimeRange Get the time range summaries (min, max) for time column in a metrics view

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view_name** | **String** |  | [required] |
**body** | [**QueryServiceMetricsViewTimeRangeRequest**](QueryServiceMetricsViewTimeRangeRequest.md) |  | [required] |

### Return type

[**models::V1MetricsViewTimeRangeResponse**](v1MetricsViewTimeRangeResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_time_series

> models::V1MetricsViewTimeSeriesResponse query_service_metrics_view_time_series(instance_id, metrics_view_name, body)
MetricsViewTimeSeries returns time series for the measures in the metrics view. It's a convenience API for querying a metrics view.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view_name** | **String** |  | [required] |
**body** | [**QueryServiceMetricsViewTimeSeriesRequest**](QueryServiceMetricsViewTimeSeriesRequest.md) |  | [required] |

### Return type

[**models::V1MetricsViewTimeSeriesResponse**](v1MetricsViewTimeSeriesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_toplist

> models::V1MetricsViewToplistResponse query_service_metrics_view_toplist(instance_id, metrics_view_name, body)
Deprecated - use MetricsViewComparison instead. MetricsViewToplist returns the top dimension values of a metrics view sorted by one or more measures. It's a convenience API for querying a metrics view.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view_name** | **String** |  | [required] |
**body** | [**DeprecatedUseMetricsViewComparisonRequestWithoutAComparisonTimeRange**](DeprecatedUseMetricsViewComparisonRequestWithoutAComparisonTimeRange.md) |  | [required] |

### Return type

[**models::V1MetricsViewToplistResponse**](v1MetricsViewToplistResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_metrics_view_totals

> models::V1MetricsViewTotalsResponse query_service_metrics_view_totals(instance_id, metrics_view_name, body)
MetricsViewTotals returns totals over a time period for the measures in a metrics view. It's a convenience API for querying a metrics view.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**metrics_view_name** | **String** |  | [required] |
**body** | [**QueryServiceMetricsViewTotalsRequest**](QueryServiceMetricsViewTotalsRequest.md) |  | [required] |

### Return type

[**models::V1MetricsViewTotalsResponse**](v1MetricsViewTotalsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_query

> models::V1QueryResponse query_service_query(instance_id, body)
Query runs a SQL query against the instance's OLAP datastore.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**QueryServiceQueryRequest**](QueryServiceQueryRequest.md) |  | [required] |

### Return type

[**models::V1QueryResponse**](v1QueryResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_query_batch

> models::StreamResultOfV1QueryBatchResponse query_service_query_batch(instance_id, body)
Batch request with different queries

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**QueryServiceQueryBatchRequest**](QueryServiceQueryBatchRequest.md) |  | [required] |

### Return type

[**models::StreamResultOfV1QueryBatchResponse**](Stream_result_of_v1QueryBatchResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_resolve_component

> models::V1ResolveComponentResponse query_service_resolve_component(instance_id, component, body)
ResolveComponent resolves the data and renderer for a Component resource.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** | Instance ID | [required] |
**component** | **String** | Component name | [required] |
**body** | [**QueryServiceResolveComponentRequest**](QueryServiceResolveComponentRequest.md) |  | [required] |

### Return type

[**models::V1ResolveComponentResponse**](v1ResolveComponentResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_table_cardinality

> models::V1TableCardinalityResponse query_service_table_cardinality(instance_id, table_name, connector, database, database_schema, priority)
TableCardinality returns row count

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** | Required | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1TableCardinalityResponse**](v1TableCardinalityResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_table_columns

> models::V1TableColumnsResponse query_service_table_columns(instance_id, table_name, connector, database, database_schema, priority)
TableColumns returns column profiles

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** |  | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1TableColumnsResponse**](v1TableColumnsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## query_service_table_rows

> models::V1TableRowsResponse query_service_table_rows(instance_id, table_name, connector, database, database_schema, limit, priority)
TableRows returns table rows

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**table_name** | **String** |  | [required] |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**limit** | Option<**i32**> |  |  |
**priority** | Option<**i32**> |  |  |

### Return type

[**models::V1TableRowsResponse**](v1TableRowsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

