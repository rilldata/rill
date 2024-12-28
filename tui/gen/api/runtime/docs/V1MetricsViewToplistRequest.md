# V1MetricsViewToplistRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  | [optional]
**metrics_view_name** | Option<**String**> |  | [optional]
**dimension_name** | Option<**String**> |  | [optional]
**measure_names** | Option<**Vec<String>**> |  | [optional]
**time_start** | Option<**String**> |  | [optional]
**time_end** | Option<**String**> |  | [optional]
**limit** | Option<**String**> |  | [optional]
**offset** | Option<**String**> |  | [optional]
**sort** | Option<[**Vec<models::V1MetricsViewSort>**](v1MetricsViewSort.md)> |  | [optional]
**r#where** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**where_sql** | Option<**String**> |  | [optional]
**having** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**having_sql** | Option<**String**> |  | [optional]
**priority** | Option<**i32**> |  | [optional]
**filter** | Option<[**models::V1MetricsViewFilter**](v1MetricsViewFilter.md)> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


