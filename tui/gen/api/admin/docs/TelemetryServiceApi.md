# \TelemetryServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**telemetry_service_record_events**](TelemetryServiceApi.md#telemetry_service_record_events) | **POST** /v1/telemetry/events | RecordEvents sends a batch of telemetry events. The events must conform to the schema described in rill/runtime/pkg/activity/README.md.



## telemetry_service_record_events

> serde_json::Value telemetry_service_record_events(body)
RecordEvents sends a batch of telemetry events. The events must conform to the schema described in rill/runtime/pkg/activity/README.md.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1RecordEventsRequest**](V1RecordEventsRequest.md) |  | [required] |

### Return type

[**serde_json::Value**](serde_json::Value.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

