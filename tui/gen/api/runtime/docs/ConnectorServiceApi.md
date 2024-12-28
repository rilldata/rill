# \ConnectorServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**connector_service_big_query_list_datasets**](ConnectorServiceApi.md#connector_service_big_query_list_datasets) | **GET** /v1/bigquery/datasets | BigQueryListDatasets list all datasets in a bigquery project
[**connector_service_big_query_list_tables**](ConnectorServiceApi.md#connector_service_big_query_list_tables) | **GET** /v1/bigquery/tables | BigQueryListTables list all tables in a bigquery project:dataset
[**connector_service_gcs_get_credentials_info**](ConnectorServiceApi.md#connector_service_gcs_get_credentials_info) | **GET** /v1/gcs/credentials_info | GCSGetCredentialsInfo returns metadata for the given bucket.
[**connector_service_gcs_list_buckets**](ConnectorServiceApi.md#connector_service_gcs_list_buckets) | **GET** /v1/gcs/buckets | GCSListBuckets lists buckets accessible with the configured credentials.
[**connector_service_gcs_list_objects**](ConnectorServiceApi.md#connector_service_gcs_list_objects) | **GET** /v1/gcs/bucket/{bucket}/objects | GCSListObjects lists objects for the given bucket.
[**connector_service_olap_get_table**](ConnectorServiceApi.md#connector_service_olap_get_table) | **GET** /v1/connectors/olap/table | OLAPGetTable returns metadata about a table or view in an OLAP
[**connector_service_olap_list_tables**](ConnectorServiceApi.md#connector_service_olap_list_tables) | **GET** /v1/olap/tables | OLAPListTables list all tables across all databases on motherduck
[**connector_service_s3_get_bucket_metadata**](ConnectorServiceApi.md#connector_service_s3_get_bucket_metadata) | **GET** /v1/s3/bucket/{bucket}/metadata | S3GetBucketMetadata returns metadata for the given bucket.
[**connector_service_s3_get_credentials_info**](ConnectorServiceApi.md#connector_service_s3_get_credentials_info) | **GET** /v1/s3/credentials_info | S3GetCredentialsInfo returns metadata for the given bucket.
[**connector_service_s3_list_buckets**](ConnectorServiceApi.md#connector_service_s3_list_buckets) | **GET** /v1/s3/buckets | S3ListBuckets lists buckets accessible with the configured credentials.
[**connector_service_s3_list_objects**](ConnectorServiceApi.md#connector_service_s3_list_objects) | **GET** /v1/s3/bucket/{bucket}/objects | S3ListBuckets lists objects for the given bucket.



## connector_service_big_query_list_datasets

> models::V1BigQueryListDatasetsResponse connector_service_big_query_list_datasets(instance_id, connector, page_size, page_token)
BigQueryListDatasets list all datasets in a bigquery project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1BigQueryListDatasetsResponse**](v1BigQueryListDatasetsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_big_query_list_tables

> models::V1BigQueryListTablesResponse connector_service_big_query_list_tables(instance_id, connector, dataset, page_size, page_token)
BigQueryListTables list all tables in a bigquery project:dataset

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |
**dataset** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1BigQueryListTablesResponse**](v1BigQueryListTablesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_gcs_get_credentials_info

> models::V1GcsGetCredentialsInfoResponse connector_service_gcs_get_credentials_info(instance_id, connector)
GCSGetCredentialsInfo returns metadata for the given bucket.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |

### Return type

[**models::V1GcsGetCredentialsInfoResponse**](v1GCSGetCredentialsInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_gcs_list_buckets

> models::V1GcsListBucketsResponse connector_service_gcs_list_buckets(instance_id, connector, page_size, page_token)
GCSListBuckets lists buckets accessible with the configured credentials.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1GcsListBucketsResponse**](v1GCSListBucketsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_gcs_list_objects

> models::V1GcsListObjectsResponse connector_service_gcs_list_objects(bucket, instance_id, connector, page_size, page_token, prefix, start_offset, end_offset, delimiter)
GCSListObjects lists objects for the given bucket.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**bucket** | **String** |  | [required] |
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |
**prefix** | Option<**String**> |  |  |
**start_offset** | Option<**String**> |  |  |
**end_offset** | Option<**String**> |  |  |
**delimiter** | Option<**String**> |  |  |

### Return type

[**models::V1GcsListObjectsResponse**](v1GCSListObjectsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_olap_get_table

> models::V1OlapGetTableResponse connector_service_olap_get_table(instance_id, connector, database, database_schema, table)
OLAPGetTable returns metadata about a table or view in an OLAP

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |
**database** | Option<**String**> |  |  |
**database_schema** | Option<**String**> |  |  |
**table** | Option<**String**> |  |  |

### Return type

[**models::V1OlapGetTableResponse**](v1OLAPGetTableResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_olap_list_tables

> models::V1OlapListTablesResponse connector_service_olap_list_tables(instance_id, connector, search_pattern)
OLAPListTables list all tables across all databases on motherduck

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> | Connector to list tables from. |  |
**search_pattern** | Option<**String**> | Optional search pattern to filter tables by. Has the same syntax and behavior as ILIKE in SQL. If the connector supports schema/database names, it searches against both the plain table name and the fully qualified table name. |  |

### Return type

[**models::V1OlapListTablesResponse**](v1OLAPListTablesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_s3_get_bucket_metadata

> models::V1S3GetBucketMetadataResponse connector_service_s3_get_bucket_metadata(bucket, instance_id, connector)
S3GetBucketMetadata returns metadata for the given bucket.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**bucket** | **String** |  | [required] |
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |

### Return type

[**models::V1S3GetBucketMetadataResponse**](v1S3GetBucketMetadataResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_s3_get_credentials_info

> models::V1S3GetCredentialsInfoResponse connector_service_s3_get_credentials_info(instance_id, connector)
S3GetCredentialsInfo returns metadata for the given bucket.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |

### Return type

[**models::V1S3GetCredentialsInfoResponse**](v1S3GetCredentialsInfoResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_s3_list_buckets

> models::V1S3ListBucketsResponse connector_service_s3_list_buckets(instance_id, connector, page_size, page_token)
S3ListBuckets lists buckets accessible with the configured credentials.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1S3ListBucketsResponse**](v1S3ListBucketsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## connector_service_s3_list_objects

> models::V1S3ListObjectsResponse connector_service_s3_list_objects(bucket, instance_id, connector, page_size, page_token, region, prefix, start_after, delimiter)
S3ListBuckets lists objects for the given bucket.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**bucket** | **String** |  | [required] |
**instance_id** | Option<**String**> |  |  |
**connector** | Option<**String**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |
**region** | Option<**String**> |  |  |
**prefix** | Option<**String**> |  |  |
**start_after** | Option<**String**> |  |  |
**delimiter** | Option<**String**> |  |  |

### Return type

[**models::V1S3ListObjectsResponse**](v1S3ListObjectsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

