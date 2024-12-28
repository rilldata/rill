# V1ModelSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**refresh_schedule** | Option<[**models::V1Schedule**](v1Schedule.md)> |  | [optional]
**timeout_seconds** | Option<**i64**> |  | [optional]
**incremental** | Option<**bool**> |  | [optional]
**incremental_state_resolver** | Option<**String**> |  | [optional]
**incremental_state_resolver_properties** | Option<[**serde_json::Value**](.md)> |  | [optional]
**partitions_resolver** | Option<**String**> |  | [optional]
**partitions_resolver_properties** | Option<[**serde_json::Value**](.md)> |  | [optional]
**partitions_watermark_field** | Option<**String**> |  | [optional]
**partitions_concurrency_limit** | Option<**i64**> |  | [optional]
**input_connector** | Option<**String**> |  | [optional]
**input_properties** | Option<[**serde_json::Value**](.md)> |  | [optional]
**stage_connector** | Option<**String**> | stage_connector is optional. | [optional]
**stage_properties** | Option<[**serde_json::Value**](.md)> |  | [optional]
**output_connector** | Option<**String**> |  | [optional]
**output_properties** | Option<[**serde_json::Value**](.md)> |  | [optional]
**trigger** | Option<**bool**> |  | [optional]
**trigger_full** | Option<**bool**> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


