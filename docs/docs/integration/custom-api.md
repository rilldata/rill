---
title: "Custom API Integration"
description: How to integrate custom APIs with your application
sidebar_label: "Custom API Integration"
sidebar_position: 20
---

Rill exposes [custom APIs](../develop/custom-apis/index.md) you have created under `apis` folder as HTTP endpoints 
at `https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/<api-name>`.

## Accessing custom APIs
You need to do a POST request to the API endpoint with A bearer token in the `Authorization` header.
    
```bash
curl -X POST https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/runtime/api/<api-name>[?query-args] \
-H "Authorization: Bearer <token>"
```

There are two types of bearer tokens that you can use to access the custom APIs:
1. **Service Account Token**: You can use a service account token to access the custom APIs.
    Read more about [Service Account Tokens](../reference/cli/service). 

    :::note
    1. Service accounts have full access to all APIs so if there are security policies defined for metrics view being used in metrics_sql API, they will not be enforced.
    2. Also since there is no user context available, this means if the api being called uses `{{ .user.<attr> }}` in SQL templating, it will fail.
    :::
    
2. **User Token**: You can use a user token to access the custom API when you want user context and enforce security policies defined for the [metrics view](../develop/metrics-dashboard.md) being used in the `metrics_sql` API.
    To get user token you need to perform a handshake with Rill's [credentials API](https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/credentials) using a service account token. Example:
    
    ```bash
    curl -X POST https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/credentials \
    -H "Authorization: Bearer <service-account-token>"
   --data-raw '{
      "user_email":"<user-email>"
    }'
    ``` 
   The API accepts the following parameters:
    | Parameter | Description                                                                                                                                                                                    | Required                         |
    | --- |------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------|
    | user_email | The email of the user for which token is being asked for                                                                                                                                                | No (either this or `attributes`) |
    | attributes | Json payload containing user attributes used for enforcing policies. When using this make sure to pass all the attributes used in your security policy like `email`, `domain` and `admin`| No (either this or `user_email`) |
    | ttl_seconds | The time to live for the iframe URL                                                                                                                                                            | No (Default: 86400)              |
