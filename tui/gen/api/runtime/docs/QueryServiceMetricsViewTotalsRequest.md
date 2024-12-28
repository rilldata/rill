# QueryServiceMetricsViewTotalsRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**measure_names** | Option<**Vec<String>**> |  | [optional]
**time_start** | Option<**String**> |  | [optional]
**time_end** | Option<**String**> |  | [optional]
**r#where** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**where_sql** | Option<**String**> | Optional. If both where and where_sql are set, both will be applied with an AND between them. | [optional]
**priority** | Option<**i32**> |  | [optional]
**filter** | Option<[**models::V1MetricsViewFilter**](v1MetricsViewFilter.md)> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


