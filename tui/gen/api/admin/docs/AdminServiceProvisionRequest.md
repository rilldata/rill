# AdminServiceProvisionRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**r#type** | Option<**String**> | Type of resource to provision. | [optional]
**name** | Option<**String**> | Name of the resource to provision. It forms a unique key together with deployment and type, which is used to de-duplicate provision requests. | [optional]
**args** | Option<[**serde_json::Value**](.md)> | Arguments for the provisioner call. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


