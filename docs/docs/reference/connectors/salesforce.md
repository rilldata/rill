---
title: Salesforce
description: Connect to data in a Salesforce org using the Bulk API
sidebar_label: Salesforce
sidebar_position: 11
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

Salesforce is a leading cloud-based Customer Relationship Management (CRM) platform designed to help businesses connect with and understand their customers better. It offers a comprehensive suite of applications focused on sales, customer service, marketing automation, analytics, and application development. Salesforce enables organizations of all sizes to build stronger relationships with their customers through personalized experiences, streamlined communication, and predictive insights. Rill is able to ingest data from Salesforce as a source by utilizing the Bulk API, which requires a Salesforce username along with a password (and in some cases, a token, depending on the org configuration) to authenticate against a Salesforce org.

![Connecting to Salesforce](/img/reference/connectors/salesforce/salesforce.png)

## Local credentials

When using Rill Developer on your local machine (i.e. `rill start`), you have the option to specify credentials when running Rill using the `--var` flag. For example, you could run the following command via the terminal (when starting Rill):
```bash
rill start --var connector.salesforce.username="user@example.com" --var connector.salesforce.password="MyPasswordMyToken"
```

Alternatively, you can also include the credentials directly in the underlying source YAML by adding the `username` and `password` parameters. For example, your source YAML may contain the following properties (also can be configured through the UI during source creation):
```yaml
type: "salesforce"
endpoint: "login.salesforce.com"
username: "user@example.com"
password: "MyPasswordMyToken"
soql: "SELECT Id, Name, CreatedDate FROM Opportunity"
sobject: "Opportunity"
```

:::warning Beware of committing credentials to Git

Outside of local development, this approach is not recommended because it places the connection string (which may contain sensitive information like passwords!) in the source YAML file, which will then be committed to Git.

:::

:::info Source Properties

For more information about available source properties / configurations, please refer to our reference documentation on [Source YAML](../../reference/project-files/index.md).

:::

:::tip Did you know?

If this project has already been deployed to Rill Cloud and credentials have been set for this source, you can use `rill env pull` to [pull these cloud credentials](../../build/credentials/credentials.md#rill-env-pull) locally (into your local `.env` file). Please note that this may override any credentials that you have set locally for this source.

:::

## Cloud deployment

Once a project having a Salesforce source has been deployed using `rill deploy`, Rill requires you to explicitly provide the credentials using the following command:

```
rill env configure
```


:::info

Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

:::

:::note

Leave the `key` and `client_id` fields blank unless using JWT (described in the next section [below](#jwt)).

:::

### JWT

Authentication using JWT instead of a password is also supported by setting
`client_id` to the **Client Id** (also known as _Consumer Key_) of the Connected App
to use, and setting `key` to contain the PEM-formatted private key to use for
signing.

:::tip Did you know?

If you've configured credentials locally already (in your `<RILL_HOME>/.home` file), you can use `rill env push` to [push these credentials](../../build/credentials/credentials.md#rill-env-push) to your Rill Cloud project. This will allow other users to retrieve / reuse the same credentials automatically by running `rill env pull`.

:::