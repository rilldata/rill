# V1ExploreSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**display_name** | Option<**String**> |  | [optional]
**description** | Option<**String**> |  | [optional]
**metrics_view** | Option<**String**> |  | [optional]
**dimensions** | Option<**Vec<String>**> | Dimensions to show. If `dimensions_selector` is set, this will only be set in `state.valid_spec`. | [optional]
**dimensions_selector** | Option<[**models::V1FieldSelector**](v1FieldSelector.md)> |  | [optional]
**measures** | Option<**Vec<String>**> | Measures to show. If `measures_selector` is set, this will only be set in `state.valid_spec`. | [optional]
**measures_selector** | Option<[**models::V1FieldSelector**](v1FieldSelector.md)> |  | [optional]
**theme** | Option<**String**> |  | [optional]
**embedded_theme** | Option<[**models::V1ThemeSpec**](v1ThemeSpec.md)> |  | [optional]
**time_ranges** | Option<[**Vec<models::V1ExploreTimeRange>**](v1ExploreTimeRange.md)> | List of selectable time ranges with comparison time ranges. If the list is empty, a default list should be shown. | [optional]
**time_zones** | Option<**Vec<String>**> | List of selectable time zones. If the list is empty, a default list should be shown. The values should be valid IANA location identifiers. | [optional]
**default_preset** | Option<[**models::V1ExplorePreset**](v1ExplorePreset.md)> |  | [optional]
**embeds_hide_pivot** | Option<**bool**> | If true, the pivot tab will be hidden when the explore is embedded. | [optional]
**security_rules** | Option<[**Vec<models::V1SecurityRule>**](v1SecurityRule.md)> | Security for the explore dashboard. These are not currently parsed from YAML, but will be derived from the parent metrics view. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


