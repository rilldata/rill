# V1GetGithubUserStatusResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**has_access** | Option<**bool**> |  | [optional]
**grant_access_url** | Option<**String**> |  | [optional]
**access_token** | Option<**String**> |  | [optional]
**account** | Option<**String**> |  | [optional]
**user_installation_permission** | Option<[**models::V1GithubPermission**](v1GithubPermission.md)> |  | [optional]
**organization_installation_permissions** | Option<[**std::collections::HashMap<String, models::V1GithubPermission>**](v1GithubPermission.md)> |  | [optional]
**organizations** | Option<**Vec<String>**> | DEPRECATED: Use organization_installation_permissions instead. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


