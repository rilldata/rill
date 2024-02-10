---
title: Salesforce
description: Connect to data in a Salesforce org using the Bulk API
sidebar_label: Salesforce
sidebar_position: 80
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## How to configure credentials in Rill

Rill utilizes a Salesforce username along with a password (and token, depending
on org configuration) to authenticate against a Salesforce org.

### Configure credentials for local development

When working on a local project, you have the option to specify credentials when running Rill using the `--variable` flag.
An example of using this syntax in terminal:
```
rill start --variable connector.salesforce.username="user@example.com" --variable connector.salesforce.password="MyPasswordMyToken"
```

Alternatively, you can include the credentials string directly in the source code by adding the `username` and `password` parameters. 
An example of a source using this approach:
```
type: "salesforce"
endpoint: "login.salesforce.com"
username: "user@example.com"
password: "MyPasswordMyToken"
soql: "SELECT Id, Name, CreatedDate FROM Opportunity"
sobject: "Opportunity"
```
This approach is less recommended because it places the connection string (which may contain sensitive information like passwords) in the source file, which is committed to Git. For more information, please refer to the documentation on [sources](../../reference/project-files/index.md).

### Configure credentials for deployments on Rill Cloud

Once a project having a Salesforce source has been deployed using `rill deploy`, Rill requires you to explicitly provide the credentials using following command:
```
rill env configure
```
Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.

Leave `key` and `client_id` blank unless using JWT as described below.

### JWT

Authentication using JWT instead of a password is also supported by setting
`client_id` to the Client Id (also known as Consumer Key) of the Connected App
to use, and setting `key` to contain the PEM-formatted private key to use for
signing.
