---
title: Deployment Credentials
description: Configuring credentials for your deployed project on Rill Cloud
sidebar_label: Deployment Credentials
sidebar_position: 11
---
## Overview

When deploying a project to Rill Cloud, credentials will need to be **separately** specified and passed in for Rill Cloud to connect to and ingest data from [remote sources](/reference/connectors/connectors.md). [Local credentials](../build/credentials/credentials.md#setting-credentials-for-rill-developer) will be used by Rill Developer to connect to sources from your local machine, typically for local devlopment and modeling, while [deployment credentials](../build/credentials/credentials.md#setting-credentials-for-a-rill-cloud-project) are what is used by Rill Cloud for production workloads. For more details about using and setting deployment credentials, please see our [configuring credentials](../build/credentials/credentials.md) page and the respective [connectors](/reference/connectors/connectors.md) or [OLAP engine](/reference/olap-engines/olap-engines.md) page.

:::info Separating development and production credentials

As a general best practice, it is strongly recommended to use service accounts and dedicated service credentials for projects deployed to Rill Cloud, especially when used in a production capacity. This will be covered in more detail in the following section below.

:::

## Service Accounts

Service accounts are non-human user accounts that provide an identity for processes or services running on a server to interact with external resources, such as databases, APIs, and cloud services. Unlike personal user accounts, service accounts are intended for use by software applications or automated tools and do not require interactive login. In the context of Rill, service accounts are credentials that should be used for projects deployed to Rill Cloud.

### Why are service accounts important?

Using service accounts for production workflows and pipelines is a general best practice for several reasons:

1. **Improved Security Posture**: Service accounts are specialized accounts used specifically by applications, as opposed to human users, to interact with data sources and other services. They help in implementing the principle of _least privilege_ by restricting permissions to only what is necessary for the application to function. This minimizes the potential damage when an account or set of credentials are compromised.

2. **Auditing and Monitoring**: Using service accounts makes it easier to audit and monitor access and activities. Since these accounts are used exclusively by applications (in this case Rill), any data access or actions performed can be traced back to the specific application, simplifying the process of identifying unusual or unauthorized activities.

3. **Credential Management**: Service accounts facilitate better management of credentials. For instance, when humans manage and share credentials, there's a higher risk of credentials being leaked or mishandled. Service accounts can be managed programmatically, reducing human error and improving security through automated rotations and stricter access controls.

4. **Scalability and Automation**: As organizations grow and deploy more applications, the use of service accounts allows for scalable and automated access management. It's easier to programmatically control access for multiple service accounts across different environments and services, fitting well into infrastructure as code (IaC) practices. Similarly, this ensures deployed projects (on Rill Cloud) don't share credentials with other applications.

5. **Compliance and Governance**: For regulatory compliance, using service accounts helps enforce data governance policies by ensuring that access is granted _according to the defined roles and responsibilities of the application_. It also aids in ensuring that data handling complies with policies and regulations since access patterns and permissions are clearly defined and can be audited.

6. **Reliability**: Applications using service accounts are <u>less likely</u> to experience downtime due to credential changes. Unlike user accounts, which may have passwords that expire or change frequently, service accounts can be configured with long-lived credentials or managed identity solutions that automatically handle authentication, reducing the risk of disruptions. Similarly, service accounts are not tied to the lifecycle of employee accounts, which help ensure that automated processes aren't potentially disrupted by personnel changes (such as offramping an employee).

### Service Account Best Practices
- _Secure Storage of Credentials_: Store service account credentials securely, using encrypted storage solutions and access controls to prevent unauthorized access.
- _Regular Rotation of Credentials_: Regularly update service account passwords and keys to reduce the risk of compromise.
- _Minimum Necessary Permissions_: Grant **only** the permissions necessary for the specific tasks the service account needs to perform, and review permissions regularly to adapt to changes in application functionality.
- _Monitoring and Logging_: Depending on your organization's security needs, you can also consider implementing monitoring and logging of all access and actions taken by service accounts to detect and respond to anomalous activities promptly.

:::warning Be careful of overriding local credentials and/or pushing the wrong credentials to Rill Cloud

When using service accounts, it is very likely that different or even personal credentials are being used in local development (i.e. Rill Developer). Therefore, it is worth double checking that the correct credentials are being used or set before [syncing credentials](../build/credentials/credentials.md#pushing-and-pulling-credentials-to--from-rill-cloud) between your local instance of [Rill Developer and Rill Cloud](../build/connect/connect.md#rill-developer-vs-rill-cloud) using the `rill env push` and `rill env pull` commands respectively

:::