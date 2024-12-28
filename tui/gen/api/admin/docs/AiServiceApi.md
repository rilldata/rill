# \AiServiceApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**a_i_service_complete**](AiServiceApi.md#a_i_service_complete) | **POST** /v1/ai/complete | Complete sends the messages of a chat to the AI and asks it to generate a new message.



## a_i_service_complete

> models::V1CompleteResponse a_i_service_complete(body)
Complete sends the messages of a chat to the AI and asks it to generate a new message.

### Parameters


Name | Type | Description  | Required | Notes
------------- | ------------- | ------------- | ------------- | -------------
**body** | [**V1CompleteRequest**](V1CompleteRequest.md) |  | [required] |

### Return type

[**models::V1CompleteResponse**](v1CompleteResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

