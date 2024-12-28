# V1MetricsViewComparisonRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**instance_id** | Option<**String**> |  | [optional]
**metrics_view_name** | Option<**String**> |  | [optional]
**dimension** | Option<[**models::V1MetricsViewAggregationDimension**](v1MetricsViewAggregationDimension.md)> |  | [optional]
**measures** | Option<[**Vec<models::V1MetricsViewAggregationMeasure>**](v1MetricsViewAggregationMeasure.md)> |  | [optional]
**comparison_measures** | Option<**Vec<String>**> |  | [optional]
**sort** | Option<[**Vec<models::V1MetricsViewComparisonSort>**](v1MetricsViewComparisonSort.md)> |  | [optional]
**time_range** | Option<[**models::V1TimeRange**](v1TimeRange.md)> |  | [optional]
**comparison_time_range** | Option<[**models::V1TimeRange**](v1TimeRange.md)> |  | [optional]
**r#where** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**where_sql** | Option<**String**> | Optional. If both where and where_sql are set, both will be applied with an AND between them. | [optional]
**having** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**having_sql** | Option<**String**> | Optional. If both having and having_sql are set, both will be applied with an AND between them. | [optional]
**aliases** | Option<[**Vec<models::V1MetricsViewComparisonMeasureAlias>**](v1MetricsViewComparisonMeasureAlias.md)> |  | [optional]
**limit** | Option<**String**> |  | [optional]
**offset** | Option<**String**> |  | [optional]
**priority** | Option<**i32**> |  | [optional]
**exact** | Option<**bool**> |  | [optional]
**filter** | Option<[**models::V1MetricsViewFilter**](v1MetricsViewFilter.md)> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


