# V1MetricsViewSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**connector** | Option<**String**> |  | [optional]
**database** | Option<**String**> |  | [optional]
**database_schema** | Option<**String**> |  | [optional]
**table** | Option<**String**> |  | [optional]
**model** | Option<**String**> | Name of the model the metrics view is based on. Either table or model should be set. | [optional]
**display_name** | Option<**String**> |  | [optional]
**description** | Option<**String**> |  | [optional]
**time_dimension** | Option<**String**> |  | [optional]
**smallest_time_grain** | Option<[**models::V1TimeGrain**](v1TimeGrain.md)> |  | [optional]
**watermark_expression** | Option<**String**> | Expression to evaluate a watermark for the metrics view. If not set, the watermark defaults to max(time_dimension). | [optional]
**dimensions** | Option<[**Vec<models::MetricsViewSpecDimensionV2>**](MetricsViewSpecDimensionV2.md)> |  | [optional]
**measures** | Option<[**Vec<models::MetricsViewSpecMeasureV2>**](MetricsViewSpecMeasureV2.md)> |  | [optional]
**security_rules** | Option<[**Vec<models::V1SecurityRule>**](v1SecurityRule.md)> |  | [optional]
**first_day_of_week** | Option<**i64**> | ISO 8601 weekday number to use as the base for time aggregations by week. Defaults to 1 (Monday). | [optional]
**first_month_of_year** | Option<**i64**> | Month number to use as the base for time aggregations by year. Defaults to 1 (January). | [optional]
**default_dimensions** | Option<**Vec<String>**> | List of selected dimensions by defaults. Deprecated: Now defined in the Explore resource. | [optional]
**default_measures** | Option<**Vec<String>**> | List of selected measures by defaults. Deprecated: Now defined in the Explore resource. | [optional]
**default_time_range** | Option<**String**> | Default time range for the dashboard. It should be a valid ISO 8601 duration string. Deprecated: Now defined in the Explore resource. | [optional]
**default_comparison_mode** | Option<[**models::MetricsViewSpecComparisonMode**](MetricsViewSpecComparisonMode.md)> |  | [optional]
**default_comparison_dimension** | Option<**String**> | If comparison mode is dimension then this determines which is the default dimension. Deprecated: Now defined in the Explore resource. | [optional]
**default_theme** | Option<**String**> | Default theme to apply. Deprecated: Now defined in the Explore resource. | [optional]
**available_time_ranges** | Option<[**Vec<models::MetricsViewSpecAvailableTimeRange>**](MetricsViewSpecAvailableTimeRange.md)> | List of available time ranges with comparison ranges that would replace the default list. Deprecated: Now defined in the Explore resource. | [optional]
**available_time_zones** | Option<**Vec<String>**> | Available time zones list preferred time zones using IANA location identifiers. Deprecated: Now defined in the Explore resource. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


