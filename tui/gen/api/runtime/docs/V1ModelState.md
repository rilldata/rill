# V1ModelState

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**executor_connector** | Option<**String**> | executor_connector is the ModelExecutor that produced the model's result. | [optional]
**result_connector** | Option<**String**> | result_connector is the connector where the model's result is stored. | [optional]
**result_properties** | Option<[**serde_json::Value**](.md)> | result_properties are returned by the executor and contains metadata about the result. | [optional]
**result_table** | Option<**String**> | result_table contains the model's result table for SQL models. It is a convenience field that can also be derived from result_properties. | [optional]
**spec_hash** | Option<**String**> | spec_hash is a hash of those parts of the spec that affect the model's result. | [optional]
**refs_hash** | Option<**String**> | refs_hash is a hash of the model's refs current state. It is used to determine if the model's refs have changed. | [optional]
**refreshed_on** | Option<**String**> | refreshed_on is the time the model was last executed. | [optional]
**incremental_state** | Option<[**serde_json::Value**](.md)> | incremental_state contains the result of the most recent invocation of the model's incremental state resolver. | [optional]
**incremental_state_schema** | Option<[**models::V1StructType**](v1StructType.md)> |  | [optional]
**partitions_model_id** | Option<**String**> | partitions_model_id is a randomly generated ID used to store the model's partitions in the CatalogStore. | [optional]
**partitions_have_errors** | Option<**bool**> | partitions_have_errors is true if one or more partitions failed to execute. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


