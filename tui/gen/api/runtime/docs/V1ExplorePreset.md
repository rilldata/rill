# V1ExplorePreset

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**dimensions** | Option<**Vec<String>**> | Dimensions to show. If `dimensions_selector` is set, this will only be set in `state.valid_spec`. | [optional]
**dimensions_selector** | Option<[**models::V1FieldSelector**](v1FieldSelector.md)> |  | [optional]
**measures** | Option<**Vec<String>**> | Measures to show. If `measures_selector` is set, this will only be set in `state.valid_spec`. | [optional]
**measures_selector** | Option<[**models::V1FieldSelector**](v1FieldSelector.md)> |  | [optional]
**r#where** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**time_range** | Option<**String**> | Time range for the explore. It corresponds to the `range` property of the explore's `time_ranges`. If not found in `time_ranges`, it should be added to the list. | [optional]
**timezone** | Option<**String**> |  | [optional]
**time_grain** | Option<**String**> |  | [optional]
**select_time_range** | Option<**String**> |  | [optional]
**comparison_mode** | Option<[**models::V1ExploreComparisonMode**](v1ExploreComparisonMode.md)> |  | [optional]
**compare_time_range** | Option<**String**> |  | [optional]
**comparison_dimension** | Option<**String**> | If comparison_mode is EXPLORE_COMPARISON_MODE_DIMENSION, this indicates the dimension to use. | [optional]
**view** | Option<[**models::V1ExploreWebView**](v1ExploreWebView.md)> |  | [optional]
**explore_sort_by** | Option<**String**> |  | [optional]
**explore_sort_asc** | Option<**bool**> |  | [optional]
**explore_sort_type** | Option<[**models::V1ExploreSortType**](v1ExploreSortType.md)> |  | [optional]
**explore_expanded_dimension** | Option<**String**> |  | [optional]
**time_dimension_measure** | Option<**String**> |  | [optional]
**time_dimension_chart_type** | Option<**String**> |  | [optional]
**time_dimension_pin** | Option<**bool**> |  | [optional]
**pivot_rows** | Option<**Vec<String>**> |  | [optional]
**pivot_cols** | Option<**Vec<String>**> |  | [optional]
**pivot_sort_by** | Option<**String**> |  | [optional]
**pivot_sort_asc** | Option<**bool**> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


