---
title: Salesforce
description: Connect to data in a Salesforce org using the Bulk API
sidebar_label: Salesforce
sidebar_position: 65
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Salesforce](https://www.salesforce.com/) is a leading cloud-based Customer Relationship Management (CRM) platform. Rill ingests data from Salesforce by issuing SOQL queries against the Bulk API. Authentication uses the Salesforce OAuth 2.0 endpoints exposed by a Connected App (or, for the client credentials and JWT flows, an External Client App).

The Salesforce connector follows the same shape as other warehouse connectors: credentials live in a connector file under `connectors/`, and each query is its own model file under `models/`.

## Authentication

The Salesforce connector supports three OAuth flows. All three require a [Connected App](https://help.salesforce.com/s/articleView?id=sf.connected_app_overview.htm) in your Salesforce org; the client credentials and JWT flows also accept an [External Client App](https://help.salesforce.com/s/articleView?id=xcloud.ecapps_intro.htm). The flow you choose determines which other fields you need to provide.

| Flow | Required fields |
| --- | --- |
| Username / Password (OAuth) | `username`, `password`, `client_id`, `client_secret` |
| Client Credentials | `client_id`, `client_secret` |
| JWT Bearer | `username`, `client_id`, `key` |

The connector picks a flow based on which credentials are populated: JWT wins when `key` is set; otherwise a `username` plus `password` selects the OAuth password flow; otherwise a `client_secret` selects the client credentials flow.

:::note SOAP login is deprecated

Earlier versions of this connector used the SOAP login endpoint when a username and password were supplied. Salesforce is decommissioning the SOAP login endpoint, so the connector now uses the OAuth password flow instead. This requires the Connected App's Client Secret in addition to the existing Client ID.

Note that the OAuth password flow only works with a Connected App; External Client Apps do not support it. The client credentials and JWT flows work with either.

:::

### Connector file

Place your credentials in a connector file at `connectors/<name>.yaml`. Reference secret values from `.env`.

#### Username / Password (OAuth)

```yaml
type: connector
driver: salesforce

endpoint: login.salesforce.com
username: user@example.com
password: "{{ .env.connector.salesforce.password }}"
client_id: "<Client ID>"
client_secret: "{{ .env.connector.salesforce.client_secret }}"
```

#### Client Credentials

```yaml
type: connector
driver: salesforce

endpoint: login.salesforce.com
client_id: "<Client ID>"
client_secret: "{{ .env.connector.salesforce.client_secret }}"
```

#### JWT Bearer

```yaml
type: connector
driver: salesforce

endpoint: login.salesforce.com
username: user@example.com
client_id: "<Client ID>"
key: "{{ .env.connector.salesforce.key }}"
```

PEM keys contain newlines, which break `.env` parsing if stored raw. The UI's file picker base64-encodes the uploaded key automatically; when hand-editing `.env`, base64-encode the PEM file yourself:

```sh
base64 < key.pem | tr -d '\n' >> .env
```

Raw PEM written inline in the connector YAML (without an `.env` reference) is also accepted.

## Models

A Salesforce model file references the connector by name and supplies the SOQL query plus the SObject the query reads from. The Bulk API needs the SObject to create the job, so it is set explicitly — Rill does not parse it from the SOQL.

```yaml
type: model
materialize: true

connector: salesforce

soql: |
  SELECT Id, Name, CreatedDate
  FROM Opportunity
sobject: Opportunity

output:
  connector: duckdb
```

Use `soql:` for the query. `sql:` is also accepted as an alias for parity with other warehouse drivers (the connector explorer in the UI writes the query into `sql:`). Add `queryAll: true` to include soft-deleted records.

## Local credentials

When using Rill Developer on your local machine, provide credentials via a connector file as shown above. Keep secrets in `.env` rather than the connector YAML. See [connector YAML](/reference/project-files/connectors) for more details.

:::tip Updating the project environmental variable

If you've already deployed to Rill Cloud, you can either [push/pull the credential]( /guide/administration/project-settings/variables-and-credentials#pushing-and-pulling-credentials-to--from-rill-cloud-via-the-cli) from the CLI with:
```
rill env push
rill env pull
```
:::

## Deploy to Rill Cloud

When deploying a project to Rill Cloud, Rill requires you to explicitly provide Salesforce credentials used in your project. See the [connector YAML reference docs](/reference/project-files/connectors) for more information.

If you subsequently add sources that require new credentials (or if you simply entered the wrong credentials during the initial deploy), update them by pushing the `Deploy` button or by running:
```
rill env push
```
