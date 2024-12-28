# V1MetricsViewRowsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  | [optional]
**metrics_view_name** | Option<**String**> |  | [optional]
**time_start** | Option<**String**> |  | [optional]
**time_end** | Option<**String**> |  | [optional]
**time_granularity** | Option<[**models::V1TimeGrain**](v1TimeGrain.md)> |  | [optional]
**r#where** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**sort** | Option<[**Vec<models::V1MetricsViewSort>**](v1MetricsViewSort.md)> |  | [optional]
**limit** | Option<**i32**> |  | [optional]
**offset** | Option<**String**> |  | [optional]
**priority** | Option<**i32**> |  | [optional]
**time_zone** | Option<**String**> |  | [optional]
**filter** | Option<[**models::V1MetricsViewFilter**](v1MetricsViewFilter.md)> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


