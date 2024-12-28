# \RuntimeServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**runtime_service_analyze_connectors**](RuntimeServiceApi.md#runtime_service_analyze_connectors) | **GET** /v1/instances/{instanceId}/connectors/analyze | AnalyzeConnectors scans all the project files and returns information about all referenced connectors.
[**runtime_service_analyze_variables**](RuntimeServiceApi.md#runtime_service_analyze_variables) | **GET** /v1/instances/{instanceId}/variables/analyze | AnalyzeVariables scans `Source`, `Model` and `Connector` resources in the catalog for use of an environment variable
[**runtime_service_create_directory**](RuntimeServiceApi.md#runtime_service_create_directory) | **POST** /v1/instances/{instanceId}/files/dir | CreateDirectory create a directory for the given path
[**runtime_service_create_instance**](RuntimeServiceApi.md#runtime_service_create_instance) | **POST** /v1/instances | CreateInstance creates a new instance
[**runtime_service_create_trigger**](RuntimeServiceApi.md#runtime_service_create_trigger) | **POST** /v1/instances/{instanceId}/trigger | CreateTrigger submits a refresh trigger, which will asynchronously refresh the specified resources. Triggers are ephemeral resources that will be cleaned up by the controller.
[**runtime_service_delete_file**](RuntimeServiceApi.md#runtime_service_delete_file) | **DELETE** /v1/instances/{instanceId}/files/entry | DeleteFile deletes a file from a repo
[**runtime_service_delete_instance**](RuntimeServiceApi.md#runtime_service_delete_instance) | **POST** /v1/instances/{instanceId} | DeleteInstance deletes an instance
[**runtime_service_edit_instance**](RuntimeServiceApi.md#runtime_service_edit_instance) | **PATCH** /v1/instances/{instanceId} | EditInstance edits an existing instance
[**runtime_service_generate_metrics_view_file**](RuntimeServiceApi.md#runtime_service_generate_metrics_view_file) | **POST** /v1/instances/{instanceId}/files/generate-metrics-view | GenerateMetricsViewFile generates a metrics view YAML file from a table in an OLAP database
[**runtime_service_generate_renderer**](RuntimeServiceApi.md#runtime_service_generate_renderer) | **POST** /v1/instances/{instanceId}/generate/renderer | GenerateRenderer generates a component renderer and renderer properties from a resolver and resolver properties
[**runtime_service_generate_resolver**](RuntimeServiceApi.md#runtime_service_generate_resolver) | **POST** /v1/instances/{instanceId}/generate/resolver | GenerateResolver generates resolver and resolver properties from a table or a metrics view
[**runtime_service_get_explore**](RuntimeServiceApi.md#runtime_service_get_explore) | **GET** /v1/instances/{instanceId}/resources/explore | GetExplore is a convenience RPC that combines looking up an Explore resource and its underlying MetricsView into one network call.
[**runtime_service_get_file**](RuntimeServiceApi.md#runtime_service_get_file) | **GET** /v1/instances/{instanceId}/files/entry | GetFile returns the contents of a specific file in a repo.
[**runtime_service_get_instance**](RuntimeServiceApi.md#runtime_service_get_instance) | **GET** /v1/instances/{instanceId} | GetInstance returns information about a specific instance
[**runtime_service_get_logs**](RuntimeServiceApi.md#runtime_service_get_logs) | **GET** /v1/instances/{instanceId}/logs | GetLogs returns recent logs from a controller
[**runtime_service_get_model_partitions**](RuntimeServiceApi.md#runtime_service_get_model_partitions) | **GET** /v1/instances/{instanceId}/models/{model}/partitions | GetModelPartitions returns the partitions of a model
[**runtime_service_get_resource**](RuntimeServiceApi.md#runtime_service_get_resource) | **GET** /v1/instances/{instanceId}/resource | GetResource looks up a specific catalog resource
[**runtime_service_health**](RuntimeServiceApi.md#runtime_service_health) | **GET** /v1/health | Health runs a health check on the runtime.
[**runtime_service_instance_health**](RuntimeServiceApi.md#runtime_service_instance_health) | **GET** /v1/health/instances/{instanceId} | InstanceHealth runs a health check on a specific instance.
[**runtime_service_issue_dev_jwt**](RuntimeServiceApi.md#runtime_service_issue_dev_jwt) | **POST** /v1/dev-jwt | IssueDevJWT issues a JWT for mimicking a user in local development.
[**runtime_service_list_connector_drivers**](RuntimeServiceApi.md#runtime_service_list_connector_drivers) | **GET** /v1/connectors/meta | ListConnectorDrivers returns a description of all the connector drivers registed in the runtime, including their configuration specs and the capabilities they support.
[**runtime_service_list_examples**](RuntimeServiceApi.md#runtime_service_list_examples) | **GET** /v1/examples | ListExamples lists all the examples embedded into binary
[**runtime_service_list_files**](RuntimeServiceApi.md#runtime_service_list_files) | **GET** /v1/instances/{instanceId}/files | ListFiles lists all the files matching a glob in a repo. The files are sorted by their full path.
[**runtime_service_list_instances**](RuntimeServiceApi.md#runtime_service_list_instances) | **GET** /v1/instances | ListInstances lists all the instances currently managed by the runtime
[**runtime_service_list_notifier_connectors**](RuntimeServiceApi.md#runtime_service_list_notifier_connectors) | **GET** /v1/instances/{instanceId}/connectors/notifiers | ListNotifierConnectors returns the names of all configured connectors that can be used as notifiers. This API is much faster than AnalyzeConnectors and can be called without admin-level permissions.
[**runtime_service_list_resources**](RuntimeServiceApi.md#runtime_service_list_resources) | **GET** /v1/instances/{instanceId}/resources | ListResources lists the resources stored in the catalog
[**runtime_service_ping**](RuntimeServiceApi.md#runtime_service_ping) | **GET** /v1/ping | Ping returns information about the runtime
[**runtime_service_put_file**](RuntimeServiceApi.md#runtime_service_put_file) | **POST** /v1/instances/{instanceId}/files/entry | PutFile creates or updates a file in a repo
[**runtime_service_rename_file**](RuntimeServiceApi.md#runtime_service_rename_file) | **POST** /v1/instances/{instanceId}/files/rename | RenameFile renames a file in a repo
[**runtime_service_unpack_empty**](RuntimeServiceApi.md#runtime_service_unpack_empty) | **POST** /v1/instances/{instanceId}/files/unpack-empty | UnpackEmpty unpacks an empty project
[**runtime_service_unpack_example**](RuntimeServiceApi.md#runtime_service_unpack_example) | **POST** /v1/instances/{instanceId}/files/unpack-example | UnpackExample unpacks an example project
[**runtime_service_watch_files**](RuntimeServiceApi.md#runtime_service_watch_files) | **GET** /v1/instances/{instanceId}/files/watch | WatchFiles streams repo file update events. It is not supported on all backends.
[**runtime_service_watch_logs**](RuntimeServiceApi.md#runtime_service_watch_logs) | **GET** /v1/instances/{instanceId}/logs/watch | WatchLogs streams new logs emitted from a controller
[**runtime_service_watch_resources**](RuntimeServiceApi.md#runtime_service_watch_resources) | **GET** /v1/instances/{instanceId}/resources/-/watch | WatchResources streams updates to catalog resources (including creation and deletion events)



## runtime_service_analyze_connectors

> models::V1AnalyzeConnectorsResponse runtime_service_analyze_connectors(instance_id)
AnalyzeConnectors scans all the project files and returns information about all referenced connectors.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |

### Return type

[**models::V1AnalyzeConnectorsResponse**](v1AnalyzeConnectorsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_analyze_variables

> models::V1AnalyzeVariablesResponse runtime_service_analyze_variables(instance_id)
AnalyzeVariables scans `Source`, `Model` and `Connector` resources in the catalog for use of an environment variable

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |

### Return type

[**models::V1AnalyzeVariablesResponse**](v1AnalyzeVariablesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_create_directory

> serde_json::Value runtime_service_create_directory(instance_id, body)
CreateDirectory create a directory for the given path

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RequestMessageForRuntimeServiceCreateDirectory**](RequestMessageForRuntimeServiceCreateDirectory.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_create_instance

> models::V1CreateInstanceResponse runtime_service_create_instance(body)
CreateInstance creates a new instance

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1CreateInstanceRequest**](V1CreateInstanceRequest.md) | Request message for RuntimeService.CreateInstance. See message Instance for field descriptions. | [required] |

### Return type

[**models::V1CreateInstanceResponse**](v1CreateInstanceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_create_trigger

> serde_json::Value runtime_service_create_trigger(instance_id, body)
CreateTrigger submits a refresh trigger, which will asynchronously refresh the specified resources. Triggers are ephemeral resources that will be cleaned up by the controller.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** | Instance to target. | [required] |
**body** | [**RuntimeServiceCreateTriggerRequest**](RuntimeServiceCreateTriggerRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_delete_file

> serde_json::Value runtime_service_delete_file(instance_id, path, force)
DeleteFile deletes a file from a repo

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**path** | Option<**String**> |  |  |
**force** | Option<**bool**> |  |  |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_delete_instance

> serde_json::Value runtime_service_delete_instance(instance_id, body)
DeleteInstance deletes an instance

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | **serde_json::Value** |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_edit_instance

> models::V1EditInstanceResponse runtime_service_edit_instance(instance_id, body)
EditInstance edits an existing instance

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RuntimeServiceEditInstanceRequest**](RuntimeServiceEditInstanceRequest.md) |  | [required] |

### Return type

[**models::V1EditInstanceResponse**](v1EditInstanceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_generate_metrics_view_file

> models::V1GenerateMetricsViewFileResponse runtime_service_generate_metrics_view_file(instance_id, body)
GenerateMetricsViewFile generates a metrics view YAML file from a table in an OLAP database

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RequestMessageForRuntimeServiceGenerateMetricsViewFile**](RequestMessageForRuntimeServiceGenerateMetricsViewFile.md) |  | [required] |

### Return type

[**models::V1GenerateMetricsViewFileResponse**](v1GenerateMetricsViewFileResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_generate_renderer

> models::V1GenerateRendererResponse runtime_service_generate_renderer(instance_id, body)
GenerateRenderer generates a component renderer and renderer properties from a resolver and resolver properties

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RuntimeServiceGenerateRendererRequest**](RuntimeServiceGenerateRendererRequest.md) |  | [required] |

### Return type

[**models::V1GenerateRendererResponse**](v1GenerateRendererResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_generate_resolver

> models::V1GenerateResolverResponse runtime_service_generate_resolver(instance_id, body)
GenerateResolver generates resolver and resolver properties from a table or a metrics view

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RuntimeServiceGenerateResolverRequest**](RuntimeServiceGenerateResolverRequest.md) |  | [required] |

### Return type

[**models::V1GenerateResolverResponse**](v1GenerateResolverResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_get_explore

> models::V1GetExploreResponse runtime_service_get_explore(instance_id, name)
GetExplore is a convenience RPC that combines looking up an Explore resource and its underlying MetricsView into one network call.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**name** | Option<**String**> |  |  |

### Return type

[**models::V1GetExploreResponse**](v1GetExploreResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_get_file

> models::V1GetFileResponse runtime_service_get_file(instance_id, path)
GetFile returns the contents of a specific file in a repo.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**path** | Option<**String**> |  |  |

### Return type

[**models::V1GetFileResponse**](v1GetFileResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_get_instance

> models::V1GetInstanceResponse runtime_service_get_instance(instance_id, sensitive)
GetInstance returns information about a specific instance

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**sensitive** | Option<**bool**> |  |  |

### Return type

[**models::V1GetInstanceResponse**](v1GetInstanceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_get_logs

> models::V1GetLogsResponse runtime_service_get_logs(instance_id, ascending, limit, level)
GetLogs returns recent logs from a controller

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**ascending** | Option<**bool**> |  |  |
**limit** | Option<**i32**> |  |  |
**level** | Option<**String**> |  |  |[default to LOG_LEVEL_UNSPECIFIED]

### Return type

[**models::V1GetLogsResponse**](v1GetLogsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_get_model_partitions

> models::V1GetModelPartitionsResponse runtime_service_get_model_partitions(instance_id, model, pending, errored, page_size, page_token)
GetModelPartitions returns the partitions of a model

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**model** | **String** |  | [required] |
**pending** | Option<**bool**> |  |  |
**errored** | Option<**bool**> |  |  |
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1GetModelPartitionsResponse**](v1GetModelPartitionsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_get_resource

> models::V1GetResourceResponse runtime_service_get_resource(instance_id, name_period_kind, name_period_name)
GetResource looks up a specific catalog resource

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**name_period_kind** | Option<**String**> |  |  |
**name_period_name** | Option<**String**> |  |  |

### Return type

[**models::V1GetResourceResponse**](v1GetResourceResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_health

> models::V1HealthResponse runtime_service_health()
Health runs a health check on the runtime.

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1HealthResponse**](v1HealthResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_instance_health

> models::V1InstanceHealthResponse runtime_service_instance_health(instance_id)
InstanceHealth runs a health check on a specific instance.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |

### Return type

[**models::V1InstanceHealthResponse**](v1InstanceHealthResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_issue_dev_jwt

> models::V1IssueDevJwtResponse runtime_service_issue_dev_jwt(body)
IssueDevJWT issues a JWT for mimicking a user in local development.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1IssueDevJwtRequest**](V1IssueDevJwtRequest.md) |  | [required] |

### Return type

[**models::V1IssueDevJwtResponse**](v1IssueDevJWTResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_list_connector_drivers

> models::V1ListConnectorDriversResponse runtime_service_list_connector_drivers()
ListConnectorDrivers returns a description of all the connector drivers registed in the runtime, including their configuration specs and the capabilities they support.

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1ListConnectorDriversResponse**](v1ListConnectorDriversResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_list_examples

> models::V1ListExamplesResponse runtime_service_list_examples()
ListExamples lists all the examples embedded into binary

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1ListExamplesResponse**](v1ListExamplesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_list_files

> models::V1ListFilesResponse runtime_service_list_files(instance_id, glob)
ListFiles lists all the files matching a glob in a repo. The files are sorted by their full path.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**glob** | Option<**String**> |  |  |

### Return type

[**models::V1ListFilesResponse**](v1ListFilesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_list_instances

> models::V1ListInstancesResponse runtime_service_list_instances(page_size, page_token)
ListInstances lists all the instances currently managed by the runtime

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**page_size** | Option<**i64**> |  |  |
**page_token** | Option<**String**> |  |  |

### Return type

[**models::V1ListInstancesResponse**](v1ListInstancesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_list_notifier_connectors

> models::V1ListNotifierConnectorsResponse runtime_service_list_notifier_connectors(instance_id)
ListNotifierConnectors returns the names of all configured connectors that can be used as notifiers. This API is much faster than AnalyzeConnectors and can be called without admin-level permissions.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |

### Return type

[**models::V1ListNotifierConnectorsResponse**](v1ListNotifierConnectorsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_list_resources

> models::V1ListResourcesResponse runtime_service_list_resources(instance_id, kind, path)
ListResources lists the resources stored in the catalog

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** | Instance to list resources from. | [required] |
**kind** | Option<**String**> | Filter by resource kind (optional). |  |
**path** | Option<**String**> | Filter by resource path (optional). |  |

### Return type

[**models::V1ListResourcesResponse**](v1ListResourcesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_ping

> models::V1PingResponse runtime_service_ping()
Ping returns information about the runtime

### Parameters

This endpoint does not need any parameter.

### Return type

[**models::V1PingResponse**](v1PingResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_put_file

> models::V1PutFileResponse runtime_service_put_file(instance_id, body)
PutFile creates or updates a file in a repo

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RequestMessageForRuntimeServicePutFile**](RequestMessageForRuntimeServicePutFile.md) |  | [required] |

### Return type

[**models::V1PutFileResponse**](v1PutFileResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_rename_file

> serde_json::Value runtime_service_rename_file(instance_id, body)
RenameFile renames a file in a repo

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RequestMessageForRuntimeServiceRenameFile**](RequestMessageForRuntimeServiceRenameFile.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_unpack_empty

> serde_json::Value runtime_service_unpack_empty(instance_id, body)
UnpackEmpty unpacks an empty project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RequestMessageForRuntimeServiceUnpackEmpty**](RequestMessageForRuntimeServiceUnpackEmpty.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_unpack_example

> serde_json::Value runtime_service_unpack_example(instance_id, body)
UnpackExample unpacks an example project

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**body** | [**RequestMessageForRuntimeServiceUnpackExample**](RequestMessageForRuntimeServiceUnpackExample.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_watch_files

> models::StreamResultOfV1WatchFilesResponse runtime_service_watch_files(instance_id, replay)
WatchFiles streams repo file update events. It is not supported on all backends.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**replay** | Option<**bool**> |  |  |

### Return type

[**models::StreamResultOfV1WatchFilesResponse**](Stream_result_of_v1WatchFilesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_watch_logs

> models::StreamResultOfV1WatchLogsResponse runtime_service_watch_logs(instance_id, replay, replay_limit, level)
WatchLogs streams new logs emitted from a controller

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**replay** | Option<**bool**> |  |  |
**replay_limit** | Option<**i32**> |  |  |
**level** | Option<**String**> |  |  |[default to LOG_LEVEL_UNSPECIFIED]

### Return type

[**models::StreamResultOfV1WatchLogsResponse**](Stream_result_of_v1WatchLogsResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)


## runtime_service_watch_resources

> models::StreamResultOfV1WatchResourcesResponse runtime_service_watch_resources(instance_id, kind, replay, level)
WatchResources streams updates to catalog resources (including creation and deletion events)

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**instance_id** | **String** |  | [required] |
**kind** | Option<**String**> |  |  |
**replay** | Option<**bool**> |  |  |
**level** | Option<**String**> |  |  |

### Return type

[**models::StreamResultOfV1WatchResourcesResponse**](Stream_result_of_v1WatchResourcesResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

