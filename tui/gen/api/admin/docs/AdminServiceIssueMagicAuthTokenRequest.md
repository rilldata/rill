# AdminServiceIssueMagicAuthTokenRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ttl_minutes** | Option<**String**> | TTL for the token in minutes. Set to 0 for no expiry. Defaults to no expiry. | [optional]
**resource_type** | Option<**String**> | Type of resource to grant access to. Currently only supports \"rill.runtime.v1.Explore\". | [optional]
**resource_name** | Option<**String**> | Name of the resource to grant access to. | [optional]
**filter** | Option<[**models::V1Expression**](v1Expression.md)> |  | [optional]
**fields** | Option<**Vec<String>**> | Optional list of fields to limit access to. If empty, no field access rule will be added. This will be translated to a rill.runtime.v1.SecurityRuleFieldAccess, which currently applies to dimension and measure names for explores and metrics views. | [optional]
**state** | Option<**String**> | Optional state to store with the token. Can be fetched with GetCurrentMagicAuthToken. | [optional]
**display_name** | Option<**String**> | Optional display name to store with the token. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


