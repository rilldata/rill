# AdminServiceGetIFrameRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**branch** | Option<**String**> | Branch to embed. If not set, the production branch is used. | [optional]
**ttl_seconds** | Option<**i64**> | TTL for the iframe's access token. If not set, defaults to 24 hours. | [optional]
**user_id** | Option<**String**> | If set, will use the attributes of the user with this ID. | [optional]
**user_email** | Option<**String**> | If set, will generate attributes corresponding to a user with this email. | [optional]
**attributes** | Option<[**serde_json::Value**](.md)> | If set, will use the provided attributes outright. | [optional]
**r#type** | Option<**String**> | Type of resource to embed. If not set, defaults to \"rill.runtime.v1.Explore\". | [optional]
**kind** | Option<**String**> | Deprecated: Alias for `type`. | [optional]
**resource** | Option<**String**> | Name of the resource to embed. This should identify a resource that is valid for embedding, such as a dashboard or component. | [optional]
**theme** | Option<**String**> | Theme to use for the embedded resource. | [optional]
**navigation** | Option<**bool**> | Navigation denotes whether navigation between different resources should be enabled in the embed. | [optional]
**state** | Option<**String**> | Blob containing UI state for rendering the initial embed. Not currently supported. | [optional]
**query** | Option<**std::collections::HashMap<String, String>**> | DEPRECATED: Additional parameters to set outright in the generated URL query. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


