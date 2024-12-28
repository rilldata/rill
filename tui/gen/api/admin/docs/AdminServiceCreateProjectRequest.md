# AdminServiceCreateProjectRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | Option<**String**> |  | [optional]
**description** | Option<**String**> |  | [optional]
**public** | Option<**bool**> |  | [optional]
**provisioner** | Option<**String**> |  | [optional]
**prod_olap_driver** | Option<**String**> |  | [optional]
**prod_olap_dsn** | Option<**String**> |  | [optional]
**prod_slots** | Option<**String**> |  | [optional]
**subpath** | Option<**String**> |  | [optional]
**prod_branch** | Option<**String**> |  | [optional]
**github_url** | Option<**String**> | github_url is set for projects whose project files are stored in github. This is set to a github repo url. Either github_url or archive_asset_id should be set. | [optional]
**archive_asset_id** | Option<**String**> | archive_asset_id is set for projects whose project files are not stored in github but are managed by rill. | [optional]
**prod_version** | Option<**String**> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


