# V1MetricsViewTimeSeriesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  | [optional]
**metrics_view_name** | Option<**String**> |  | [optional]
**measure_names** | Option<**Vec<String>**> |  | [optional]
**time_start** | Option<**String**> |  | [optional]
**time_end** | Option<**String**> |  | [optional]
**time_granularity** | Option<[**models::V1TimeGrain**](v1TimeGrain.md)> |  | [optional]
**r#where** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**where_sql** | Option<**String**> | Optional. If both where and where_sql are set, both will be applied with an AND between them. | [optional]
**having** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**having_sql** | Option<**String**> | Optional. If both having and having_sql are set, both will be applied with an AND between them. | [optional]
**time_zone** | Option<**String**> |  | [optional]
**priority** | Option<**i32**> |  | [optional]
**filter** | Option<[**models::V1MetricsViewFilter**](v1MetricsViewFilter.md)> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


