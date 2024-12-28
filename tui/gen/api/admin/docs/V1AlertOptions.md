# V1AlertOptions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**display_name** | Option<**String**> |  | [optional]
**interval_duration** | Option<**String**> |  | [optional]
**resolver** | Option<**String**> |  | [optional]
**resolver_properties** | Option<[**serde_json::Value**](.md)> |  | [optional]
**query_name** | Option<**String**> | DEPRECATED: Use resolver and resolver_properties instead. | [optional]
**query_args_json** | Option<**String**> | DEPRECATED: Use resolver and resolver_properties instead. | [optional]
**metrics_view_name** | Option<**String**> |  | [optional]
**renotify** | Option<**bool**> |  | [optional]
**renotify_after_seconds** | Option<**i64**> |  | [optional]
**email_recipients** | Option<**Vec<String>**> |  | [optional]
**slack_users** | Option<**Vec<String>**> |  | [optional]
**slack_channels** | Option<**Vec<String>**> |  | [optional]
**slack_webhooks** | Option<**Vec<String>**> |  | [optional]
**web_open_path** | Option<**String**> | Annotation for the subpath of <UI host>/org/project to open for the report. | [optional]
**web_open_state** | Option<**String**> | Annotation for the base64-encoded UI state to open for the report. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


