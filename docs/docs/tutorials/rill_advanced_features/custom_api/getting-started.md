---
title: "Custom APIs"
description:  "Getting Started with API"
sidebar_label: "Getting Started"
sidebar_position: 12
---

## Getting started with custom APIs

Our [custom APIs](https://docs.rilldata.com/integrate/custom-api#accessing-custom-apis) support both GET and POST requests to the API endpoint that exist under the `/apis/` folder in your project.

GET:
```bash
curl https://admin.rilldata.com/v1/orgs/<org-name>/projects/<project-name>/runtime/api/<api-name>[?query-args] \
-H "Authorization: Bearer <token>"
```

POST:
```bash
curl -X POST https://admin.rilldata.com/v1/orgs/<org-name>/projects/<project-name>/runtime/api/<api-name>[?query-args] \
-H "Authorization: Bearer <token>"
```

In order to send the request to the API endpoint, you will need to setup the token via the CLI. There are two types of bearer tokens:

- Service Account token
- User Token


### Accessing custom APIs
Let's create both a **service account token** and a **user token** to see the difference in the responses.

Navigate back to the CLI to run the following command:
```bash
rill service create my-api-service
Created service "my-api-service" in org "Rill_Learn".
Access token: rill_svc_<RANDOM_STRING>
```

Once this is created, you can use the service access token to create a user token. Using [dashboard access policies](https://docs.rilldata.com/manage/security), we can add the following to our [advanced_metrics_view_explore.yaml](https://docs.rilldata.com/tutorials/advanced_developer/advanced-dashboard.md) file.
```yaml
security:
  access: "{{ .user.admin }} AND '{{ .user.domain }}' == 'rilldata.com'"
  ```

This access policy gives access to the dashboard for admins who's email domain is rilldata.com. For the user token below, please select an email for a user that is [a viewer](https://docs.rilldata.com/tutorials/administration/user-management) to the project, `my-rill-tutorial`.

```bash
curl -X POST https://admin.rilldata.com/v1/organizations/<ORG_NAME>/projects/<PROJECT_NAME>/credentials \
-H "Authorization: Bearer rill_svc_<RANDOM_STRING>" --data-raw '{
  "user_email": "email@domaim.com"
}'
```

This will return a reponse which you will need to extract the accessToken.

