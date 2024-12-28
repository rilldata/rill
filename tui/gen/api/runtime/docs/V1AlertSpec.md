# V1AlertSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**display_name** | Option<**String**> |  | [optional]
**trigger** | Option<**bool**> |  | [optional]
**refresh_schedule** | Option<[**models::V1Schedule**](v1Schedule.md)> |  | [optional]
**watermark_inherit** | Option<**bool**> | If true, will use the lowest watermark of its refs instead of the trigger time. | [optional]
**intervals_iso_duration** | Option<**String**> |  | [optional]
**intervals_limit** | Option<**i32**> |  | [optional]
**intervals_check_unclosed** | Option<**bool**> |  | [optional]
**timeout_seconds** | Option<**i64**> |  | [optional]
**query_name** | Option<**String**> |  | [optional]
**query_args_json** | Option<**String**> |  | [optional]
**resolver** | Option<**String**> |  | [optional]
**resolver_properties** | Option<[**serde_json::Value**](.md)> |  | [optional]
**query_for_user_id** | Option<**String**> |  | [optional]
**query_for_user_email** | Option<**String**> |  | [optional]
**query_for_attributes** | Option<[**serde_json::Value**](.md)> |  | [optional]
**notify_on_recover** | Option<**bool**> |  | [optional]
**notify_on_fail** | Option<**bool**> |  | [optional]
**notify_on_error** | Option<**bool**> |  | [optional]
**renotify** | Option<**bool**> |  | [optional]
**renotify_after_seconds** | Option<**i64**> |  | [optional]
**notifiers** | Option<[**Vec<models::V1Notifier>**](v1Notifier.md)> |  | [optional]
**annotations** | Option<**std::collections::HashMap<String, String>**> |  | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


