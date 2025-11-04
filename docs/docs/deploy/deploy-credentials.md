---
title:  Deployment Credential Considerations
description: Configuring credentials for your deployed project on Rill Cloud
sidebar_label: Configure Deployment Credentials
sidebar_position: 10
---

:::tip Dev/prod setup

We recommend reviewing the [Dev/Prod Setup](/build/connectors/templating) documentation before deploying your project to Rill Cloud to ensure that your local testing and production environments have been separated. 

This will ensure that your shared dashboard will be decoupled from your dev environment, and you can further develop your dashboards locally without worrying about data availability. 

:::

When deploying a project, credentials that have been defined in your `.env` file will be automatically passed into your Rill Cloud project. However, for [remote sources](/build/connectors) that are dynamically retrieving your credentials via the CLI, such as S3 and GCS, you will need to ensure that these are [defined in the .env file](/manage/project-management/variables-and-credentials#credentials-naming-schema). 


[Local credentials](/build/connectors/credentials#setting-credentials-for-rill-developer) are used by Rill Developer to connect to sources from your local machine, while [deployment credentials](/deploy/deploy-credentials#configure-environmental-variables-and-credentials-for-rill-cloud) are what is used by Rill Cloud for production workloads. There are a [few ways to set up credentials in Rill Developer](/build/connectors/credentials/#setting-credentials-for-rill-developer), however you will need to ensure that they are set up in your `.env` file for a seamless experience.



## Configure Environmental Variables and Credentials for Rill Cloud 

If you have defined your connector's credentials in your `.env` file, these will be deployed along with your project. You should see the credentials in [your project's settings page.](/manage/project-management/variables-and-credentials#modifying-variables-and-credentials-via-the-settings-page)

<img src = '/img/tutorials/admin/env-var-ui.png' class='rounded-gif' />
<br />


If not, after deploying to Rill Cloud, you can run the following in the CLI to configure all of your required credentials: `rill env configure`. When running this command, Rill will detect any connectors that are being used by the project and prompt you to fill in the required fields. When completed, this will be pushed to your Rill Cloud Deployment and automatically refresh the required objects. Once completed, you will see these in your project's environmental variables settings page. 


```bash
$rill env configure
Finish deploying your project by providing access to the connectors. Rill requires credentials for the following connectors:

 - your connectors here (used by models and sources)

Configuring connector "bigquery":
...

Updated project variables
```
## Service Accounts
:::info Separating development and production credentials

As a general best practice, it is strongly recommended to use service accounts and dedicated service credentials for projects deployed to Rill Cloud, especially when used in a production capacity. This is covered in more detail in our [Dev/Prod Setup documentation](/build/connectors/templating).

:::

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

When using service accounts, it is very likely that different or even personal credentials are being used in local development (i.e., Rill Developer). Therefore, it is worth double-checking that the correct credentials are being used or set before [syncing credentials](/build/connectors/credentials#pulling-credentials-and-variables-from-a-deployed-project-on-rill-cloud) between your local instance of [Rill Developer and Rill Cloud](/get-started/concepts/cloud-vs-developer) using the `rill env push` and `rill env pull` commands respectively.

:::