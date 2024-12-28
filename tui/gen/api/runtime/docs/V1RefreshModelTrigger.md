# V1RefreshModelTrigger

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**model** | Option<**String**> | The model to refresh. | [optional]
**full** | Option<**bool**> | If true, the current table and state will be dropped before refreshing. For non-incremental models, this is equivalent to a normal refresh. | [optional]
**partitions** | Option<**Vec<String>**> | Keys of specific partitions to refresh. | [optional]
**all_errored_partitions** | Option<**bool**> | If true, it will refresh all partitions that errored on their last execution. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


