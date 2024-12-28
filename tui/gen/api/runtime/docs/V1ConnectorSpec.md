# V1ConnectorSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**driver** | Option<**String**> |  | [optional]
**properties** | Option<**std::collections::HashMap<String, String>**> |  | [optional]
**templated_properties** | Option<**Vec<String>**> |  | [optional]
**provision** | Option<**bool**> |  | [optional]
**provision_args** | Option<[**serde_json::Value**](.md)> |  | [optional]
**properties_from_variables** | Option<**std::collections::HashMap<String, String>**> | DEPRECATED: properties_from_variables stores properties whose value is a variable. NOTE : properties_from_variables and properties both should be used to get all properties. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


