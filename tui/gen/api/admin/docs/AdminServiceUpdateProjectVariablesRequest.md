# AdminServiceUpdateProjectVariablesRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**environment** | Option<**String**> | Environment to set the variables for. If empty, the variable(s) will be used as defaults for all environments. | [optional]
**variables** | Option<**std::collections::HashMap<String, String>**> | New variable values. It is NOT NECESSARY to pass all variables, existing variables not included in the request will be left unchanged. | [optional]
**unset_variables** | Option<**Vec<String>**> | Variables to delete. | [optional]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


