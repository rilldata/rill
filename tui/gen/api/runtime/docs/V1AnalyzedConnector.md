# V1AnalyzedConnector

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | Option<**String**> |  | [optional]
**driver** | Option<[**models::V1ConnectorDriver**](v1ConnectorDriver.md)> |  | [optional]
**config** | Option<**std::collections::HashMap<String, String>**> |  | [optional]
**preset_config** | Option<**std::collections::HashMap<String, String>**> |  | [optional]
**project_config** | Option<**std::collections::HashMap<String, String>**> |  | [optional]
**env_config** | Option<**std::collections::HashMap<String, String>**> |  | [optional]
**provision** | Option<**bool**> |  | [optional]
**provision_args** | Option<[**serde_json::Value**](.md)> |  | [optional]
**has_anonymous_access** | Option<**bool**> |  | [optional]
**used_by** | Option<[**Vec<models::V1ResourceName>**](v1ResourceName.md)> |  | [optional]
**error_message** | Option<**String**> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


