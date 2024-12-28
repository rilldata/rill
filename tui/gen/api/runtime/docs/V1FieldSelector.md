# V1FieldSelector

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**invert** | Option<**bool**> | Invert the result such that all fields *except* the selected fields are returned. | [optional]
**all** | Option<**bool**> | Select all fields. | [optional]
**fields** | Option<[**models::V1StringListValue**](v1StringListValue.md)> |  | [optional]
**regex** | Option<**String**> | Select fields by a regular expression. | [optional]
**duckdb_expression** | Option<**String**> | Select fields by a DuckDB SQL SELECT expression. For example \"* EXCLUDE (city)\". | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


