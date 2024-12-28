# V1RefreshTriggerSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**resources** | Option<[**Vec<models::V1ResourceName>**](v1ResourceName.md)> | Resources to refresh. The refreshable types are sources, models, alerts, reports, and the project parser. If a model is specified, a normal incremental refresh is triggered. Use the \"models\" field to trigger other kinds of model refreshes. | [optional]
**models** | Option<[**Vec<models::V1RefreshModelTrigger>**](v1RefreshModelTrigger.md)> | Models to refresh. These are specified separately to enable more fine-grained configuration. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


