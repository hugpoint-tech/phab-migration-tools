# phab_tool.gitea_api.SettingsApi

All URIs are relative to */api/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**get_general_api_settings**](SettingsApi.md#get_general_api_settings) | **GET** /settings/api | Get instance&#39;s global settings for api
[**get_general_attachment_settings**](SettingsApi.md#get_general_attachment_settings) | **GET** /settings/attachment | Get instance&#39;s global settings for Attachment
[**get_general_repository_settings**](SettingsApi.md#get_general_repository_settings) | **GET** /settings/repository | Get instance&#39;s global settings for repositories
[**get_general_ui_settings**](SettingsApi.md#get_general_ui_settings) | **GET** /settings/ui | Get instance&#39;s global settings for ui


# **get_general_api_settings**
> GeneralAPISettings get_general_api_settings()

Get instance's global settings for api

### Example

* Api Key Authentication (AccessToken):
* Api Key Authentication (AuthorizationHeaderToken):
* Basic Authentication (BasicAuth):
* Api Key Authentication (SudoHeader):
* Api Key Authentication (SudoParam):
* Api Key Authentication (TOTPHeader):
* Api Key Authentication (Token):

```python
import time
import phab_tool.gitea_api
from phab_tool.gitea_api.api import settings_api
from phab_tool.gitea_api.model.general_api_settings import GeneralAPISettings
from pprint import pprint
# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = phab_tool.gitea_api.Configuration(
    host = "/api/v1"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: AccessToken
configuration.api_key['AccessToken'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['AccessToken'] = 'Bearer'

# Configure API key authorization: AuthorizationHeaderToken
configuration.api_key['AuthorizationHeaderToken'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['AuthorizationHeaderToken'] = 'Bearer'

# Configure HTTP basic authorization: BasicAuth
configuration = phab_tool.gitea_api.Configuration(
    username = 'YOUR_USERNAME',
    password = 'YOUR_PASSWORD'
)

# Configure API key authorization: SudoHeader
configuration.api_key['SudoHeader'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['SudoHeader'] = 'Bearer'

# Configure API key authorization: SudoParam
configuration.api_key['SudoParam'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['SudoParam'] = 'Bearer'

# Configure API key authorization: TOTPHeader
configuration.api_key['TOTPHeader'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['TOTPHeader'] = 'Bearer'

# Configure API key authorization: Token
configuration.api_key['Token'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['Token'] = 'Bearer'

# Enter a context with an instance of the API client
with phab_tool.gitea_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = settings_api.SettingsApi(api_client)

    # example, this endpoint has no required or optional parameters
    try:
        # Get instance's global settings for api
        api_response = api_instance.get_general_api_settings()
        pprint(api_response)
    except phab_tool.gitea_api.ApiException as e:
        print("Exception when calling SettingsApi->get_general_api_settings: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

[**GeneralAPISettings**](GeneralAPISettings.md)

### Authorization

[AccessToken](../README.md#AccessToken), [AuthorizationHeaderToken](../README.md#AuthorizationHeaderToken), [BasicAuth](../README.md#BasicAuth), [SudoHeader](../README.md#SudoHeader), [SudoParam](../README.md#SudoParam), [TOTPHeader](../README.md#TOTPHeader), [Token](../README.md#Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | GeneralAPISettings |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_general_attachment_settings**
> GeneralAttachmentSettings get_general_attachment_settings()

Get instance's global settings for Attachment

### Example

* Api Key Authentication (AccessToken):
* Api Key Authentication (AuthorizationHeaderToken):
* Basic Authentication (BasicAuth):
* Api Key Authentication (SudoHeader):
* Api Key Authentication (SudoParam):
* Api Key Authentication (TOTPHeader):
* Api Key Authentication (Token):

```python
import time
import phab_tool.gitea_api
from phab_tool.gitea_api.api import settings_api
from phab_tool.gitea_api.model.general_attachment_settings import GeneralAttachmentSettings
from pprint import pprint
# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = phab_tool.gitea_api.Configuration(
    host = "/api/v1"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: AccessToken
configuration.api_key['AccessToken'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['AccessToken'] = 'Bearer'

# Configure API key authorization: AuthorizationHeaderToken
configuration.api_key['AuthorizationHeaderToken'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['AuthorizationHeaderToken'] = 'Bearer'

# Configure HTTP basic authorization: BasicAuth
configuration = phab_tool.gitea_api.Configuration(
    username = 'YOUR_USERNAME',
    password = 'YOUR_PASSWORD'
)

# Configure API key authorization: SudoHeader
configuration.api_key['SudoHeader'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['SudoHeader'] = 'Bearer'

# Configure API key authorization: SudoParam
configuration.api_key['SudoParam'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['SudoParam'] = 'Bearer'

# Configure API key authorization: TOTPHeader
configuration.api_key['TOTPHeader'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['TOTPHeader'] = 'Bearer'

# Configure API key authorization: Token
configuration.api_key['Token'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['Token'] = 'Bearer'

# Enter a context with an instance of the API client
with phab_tool.gitea_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = settings_api.SettingsApi(api_client)

    # example, this endpoint has no required or optional parameters
    try:
        # Get instance's global settings for Attachment
        api_response = api_instance.get_general_attachment_settings()
        pprint(api_response)
    except phab_tool.gitea_api.ApiException as e:
        print("Exception when calling SettingsApi->get_general_attachment_settings: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

[**GeneralAttachmentSettings**](GeneralAttachmentSettings.md)

### Authorization

[AccessToken](../README.md#AccessToken), [AuthorizationHeaderToken](../README.md#AuthorizationHeaderToken), [BasicAuth](../README.md#BasicAuth), [SudoHeader](../README.md#SudoHeader), [SudoParam](../README.md#SudoParam), [TOTPHeader](../README.md#TOTPHeader), [Token](../README.md#Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | GeneralAttachmentSettings |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_general_repository_settings**
> GeneralRepoSettings get_general_repository_settings()

Get instance's global settings for repositories

### Example

* Api Key Authentication (AccessToken):
* Api Key Authentication (AuthorizationHeaderToken):
* Basic Authentication (BasicAuth):
* Api Key Authentication (SudoHeader):
* Api Key Authentication (SudoParam):
* Api Key Authentication (TOTPHeader):
* Api Key Authentication (Token):

```python
import time
import phab_tool.gitea_api
from phab_tool.gitea_api.api import settings_api
from phab_tool.gitea_api.model.general_repo_settings import GeneralRepoSettings
from pprint import pprint
# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = phab_tool.gitea_api.Configuration(
    host = "/api/v1"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: AccessToken
configuration.api_key['AccessToken'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['AccessToken'] = 'Bearer'

# Configure API key authorization: AuthorizationHeaderToken
configuration.api_key['AuthorizationHeaderToken'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['AuthorizationHeaderToken'] = 'Bearer'

# Configure HTTP basic authorization: BasicAuth
configuration = phab_tool.gitea_api.Configuration(
    username = 'YOUR_USERNAME',
    password = 'YOUR_PASSWORD'
)

# Configure API key authorization: SudoHeader
configuration.api_key['SudoHeader'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['SudoHeader'] = 'Bearer'

# Configure API key authorization: SudoParam
configuration.api_key['SudoParam'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['SudoParam'] = 'Bearer'

# Configure API key authorization: TOTPHeader
configuration.api_key['TOTPHeader'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['TOTPHeader'] = 'Bearer'

# Configure API key authorization: Token
configuration.api_key['Token'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['Token'] = 'Bearer'

# Enter a context with an instance of the API client
with phab_tool.gitea_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = settings_api.SettingsApi(api_client)

    # example, this endpoint has no required or optional parameters
    try:
        # Get instance's global settings for repositories
        api_response = api_instance.get_general_repository_settings()
        pprint(api_response)
    except phab_tool.gitea_api.ApiException as e:
        print("Exception when calling SettingsApi->get_general_repository_settings: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

[**GeneralRepoSettings**](GeneralRepoSettings.md)

### Authorization

[AccessToken](../README.md#AccessToken), [AuthorizationHeaderToken](../README.md#AuthorizationHeaderToken), [BasicAuth](../README.md#BasicAuth), [SudoHeader](../README.md#SudoHeader), [SudoParam](../README.md#SudoParam), [TOTPHeader](../README.md#TOTPHeader), [Token](../README.md#Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | GeneralRepoSettings |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **get_general_ui_settings**
> GeneralUISettings get_general_ui_settings()

Get instance's global settings for ui

### Example

* Api Key Authentication (AccessToken):
* Api Key Authentication (AuthorizationHeaderToken):
* Basic Authentication (BasicAuth):
* Api Key Authentication (SudoHeader):
* Api Key Authentication (SudoParam):
* Api Key Authentication (TOTPHeader):
* Api Key Authentication (Token):

```python
import time
import phab_tool.gitea_api
from phab_tool.gitea_api.api import settings_api
from phab_tool.gitea_api.model.general_ui_settings import GeneralUISettings
from pprint import pprint
# Defining the host is optional and defaults to /api/v1
# See configuration.py for a list of all supported configuration parameters.
configuration = phab_tool.gitea_api.Configuration(
    host = "/api/v1"
)

# The client must configure the authentication and authorization parameters
# in accordance with the API server security policy.
# Examples for each auth method are provided below, use the example that
# satisfies your auth use case.

# Configure API key authorization: AccessToken
configuration.api_key['AccessToken'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['AccessToken'] = 'Bearer'

# Configure API key authorization: AuthorizationHeaderToken
configuration.api_key['AuthorizationHeaderToken'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['AuthorizationHeaderToken'] = 'Bearer'

# Configure HTTP basic authorization: BasicAuth
configuration = phab_tool.gitea_api.Configuration(
    username = 'YOUR_USERNAME',
    password = 'YOUR_PASSWORD'
)

# Configure API key authorization: SudoHeader
configuration.api_key['SudoHeader'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['SudoHeader'] = 'Bearer'

# Configure API key authorization: SudoParam
configuration.api_key['SudoParam'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['SudoParam'] = 'Bearer'

# Configure API key authorization: TOTPHeader
configuration.api_key['TOTPHeader'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['TOTPHeader'] = 'Bearer'

# Configure API key authorization: Token
configuration.api_key['Token'] = 'YOUR_API_KEY'

# Uncomment below to setup prefix (e.g. Bearer) for API key, if needed
# configuration.api_key_prefix['Token'] = 'Bearer'

# Enter a context with an instance of the API client
with phab_tool.gitea_api.ApiClient(configuration) as api_client:
    # Create an instance of the API class
    api_instance = settings_api.SettingsApi(api_client)

    # example, this endpoint has no required or optional parameters
    try:
        # Get instance's global settings for ui
        api_response = api_instance.get_general_ui_settings()
        pprint(api_response)
    except phab_tool.gitea_api.ApiException as e:
        print("Exception when calling SettingsApi->get_general_ui_settings: %s\n" % e)
```


### Parameters
This endpoint does not need any parameter.

### Return type

[**GeneralUISettings**](GeneralUISettings.md)

### Authorization

[AccessToken](../README.md#AccessToken), [AuthorizationHeaderToken](../README.md#AuthorizationHeaderToken), [BasicAuth](../README.md#BasicAuth), [SudoHeader](../README.md#SudoHeader), [SudoParam](../README.md#SudoParam), [TOTPHeader](../README.md#TOTPHeader), [Token](../README.md#Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details

| Status code | Description | Response headers |
|-------------|-------------|------------------|
**200** | GeneralUISettings |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

