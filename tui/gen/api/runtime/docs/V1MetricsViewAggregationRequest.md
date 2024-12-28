# V1MetricsViewAggregationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  | [optional]
**metrics_view** | Option<**String**> |  | [optional]
**dimensions** | Option<[**Vec<models::V1MetricsViewAggregationDimension>**](v1MetricsViewAggregationDimension.md)> |  | [optional]
**measures** | Option<[**Vec<models::V1MetricsViewAggregationMeasure>**](v1MetricsViewAggregationMeasure.md)> |  | [optional]
**sort** | Option<[**Vec<models::V1MetricsViewAggregationSort>**](v1MetricsViewAggregationSort.md)> |  | [optional]
**time_range** | Option<[**models::V1TimeRange**](v1TimeRange.md)> |  | [optional]
**comparison_time_range** | Option<[**models::V1TimeRange**](v1TimeRange.md)> |  | [optional]
**time_start** | Option<**String**> |  | [optional]
**time_end** | Option<**String**> |  | [optional]
**pivot_on** | Option<**Vec<String>**> |  | [optional]
**aliases** | Option<[**Vec<models::V1MetricsViewComparisonMeasureAlias>**](v1MetricsViewComparisonMeasureAlias.md)> |  | [optional]
**r#where** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**where_sql** | Option<**String**> | Optional. If both where and where_sql are set, both will be applied with an AND between them. | [optional]
**having** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**having_sql** | Option<**String**> | Optional. If both having and having_sql are set, both will be applied with an AND between them. | [optional]
**limit** | Option<**String**> |  | [optional]
**offset** | Option<**String**> |  | [optional]
**priority** | Option<**i32**> |  | [optional]
**filter** | Option<[**models::V1MetricsViewFilter**](v1MetricsViewFilter.md)> |  | [optional]
**exact** | Option<**bool**> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


