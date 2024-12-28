# RuntimeServiceCreateTriggerRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**resources** | Option<[**Vec<models::V1ResourceName>**](v1ResourceName.md)> | Resources to trigger. See RefreshTriggerSpec for details. | [optional]
**models** | Option<[**Vec<models::V1RefreshModelTrigger>**](v1RefreshModelTrigger.md)> | Models to trigger. Unlike resources, this supports advanced configuration of the refresh trigger. | [optional]
**parser** | Option<**bool**> | Parser is a convenience flag to trigger the global project parser. Triggering the project parser ensures a pull of the repository and a full parse of all files. | [optional]
**all_sources_models** | Option<**bool**> | Convenience flag to trigger all sources and models. | [optional]
**all_sources_models_full** | Option<**bool**> | Convenience flag to trigger all sources and models. Will trigger models with RefreshModelTrigger.full set to true. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


