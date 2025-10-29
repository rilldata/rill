---
title: Salesforce
description: Connect to data in a Salesforce org using the Bulk API
sidebar_label: Salesforce
sidebar_position: 65
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

[Salesforce](https://www.salesforce.com/) is a leading cloud-based Customer Relationship Management (CRM) platform designed to help businesses connect with and understand their customers better. It offers a comprehensive suite of applications focused on sales, customer service, marketing automation, analytics, and application development. Salesforce enables organizations of all sizes to build stronger relationships with their customers through personalized experiences, streamlined communication, and predictive insights. Rill can ingest data from Salesforce as a source by utilizing the Bulk API, which requires a Salesforce username and password (and, in some cases, a token, depending on the org configuration) to authenticate against a Salesforce org.

<img src='/img/build/connectors/data-sources/salesforce.png' class='rounded-gif' style={{width: '75%', display: 'block', margin: '0 auto'}}/>
<br />


## Local credentials

When using Rill Developer on your local machine, you will need to provide your credentials via a connector file. We would recommend not using plain text to create your file and instead use the `.env` file. For more details on your connector, see [connector YAML](/reference/project-files/connectors#salesforce) for more details.

:::tip Updating the project environmental variable

If you've already deployed to Rill Cloud, you can either [push/pull the credential](/manage/project-management/variables-and-credentials#pushing-and-pulling-credentials-to--from-rill-cloud-via-the-cli) from the CLI with:
```
rill env push
rill env pull
```
:::

Alternatively, you can include the credentials directly in the underlying source YAML by adding the `username` and `password` parameters. For example, your source YAML may contain the following properties (these can also be configured through the UI during source creation):
```yaml
type: "model"
connector: "salesforce"
endpoint: "login.salesforce.com"
username: "user@example.com"
password: "MyPasswordMyToken"
soql: "SELECT Id, Name, CreatedDate FROM Opportunity"
sobject: "Opportunity"
```

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](/build/connectors/credentials/#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials you have set locally for this source.

:::

## Cloud deployment

Once a project with a Salesforce source has been deployed, Rill requires you to explicitly provide the credentials using the following command:

```
rill env configure
```


:::note

Leave the `key` and `client_id` fields blank unless you are using JWT (described in the next section [below](#jwt)).

:::

### JWT

Authentication using JWT instead of a password is also supported. Set `client_id` to the **Client Id** (also known as the _Consumer Key_) of the Connected App to use, and set `key` to contain the PEM-formatted private key to use for signing.
