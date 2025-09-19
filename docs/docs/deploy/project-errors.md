---
title: Managing Project Errors
description: Explore and configure alerts for project errors
sidebar_label: Managing Project Errors
sidebar_position: 25
---

Rill projects can go into an error state for many reasons, such as a malformed YAML file, missing credentials for a connector, or a breaking change in a data type.
Regardless of the error, Rill Cloud takes various steps to surface, manage, and contain errors:

- **Visibility:** Admins will always be able to view the project status at the individual resource level within Rill Cloud using the `Status` tab or by using the [rill project status](/reference/cli/project/status.md) CLI command.
- **Isolation:** Rill Cloud will handle errors at the individual resource level. For example, if a dashboard falls into an error state or fails to reconcile, all other dashboards should remain available. 
- **Fallback:** Rill Cloud will attempt to fall back to the most recent valid state when possible. For example, if the underlying model for a dashboard fails to build, the dashboard will keep serving from the most recent valid state.

There are times where you'll need to check downstream objects to find the true cause of an issue. The surfaced error might look like a dashboard is not displaying but the root cause of the issue is a model timeout. You'll need to check the project's status page to ensure you are looking at the root cause.

## Resource Level Errors

If you have already created YAML alerts, when deploying to Rill Cloud for the first time, you'll get a notification if there are any [resource level errors](/build/alerts#project-status-alerts). 

You'll receive something similar to the below via email or slack, depending on your notify settings.
```
Project resource Status Alert
Your alert triggered for Thu, 06 Feb 2025 23:28:00 UTC. The first row that matched your alert criteria is:
• error: dependency error: resource "commits_metrics" (rill.runtime.v1.MetricsView) has an error
• name: commits_explore
• status: Idle
• type: Explore
```

This will give you a good idea of what object has an issue, and you can browse the [status page](/manage/project-management#checking-deployment-status) for more information.

## Credentials Errors

If this is a first deployment or your credentials expired, your source models may fail to load. Some common error messages are:
```
HTTP 401: Unauthorized
HTTP 403: Forbidden

Insufficient privileges to operate on table 'SCHEMA.TABLE'
Invalid username or password

Access Denied: BigQuery BigQuery: Permission denied for table
Invalid credentials provided
```
If these or similar errors are seen in your project's [status page](/manage/project-management/variables-and-credentials), you will need to make changes to your credentials. This can be done in your [project's setting page]/manage/project-management/variables-and-credentials) or via the CLI running `rill env configure`.


## Model Errors

Model errors typically occur when there are issues with data processing, SQL syntax, or data type mismatches. Most of these errors will have surfaced in Rill Developer but a few possible issues after deploying to Rill Cloud are:

1. **Prod vs dev environment** - The prod parameters in your YAML files dont exist.
2. **Timeouts, OOM** - Assuming the data in prod is a lot larger, timeouts and OOM issues can occur. [Contact us](/contact) if you see any related error messages.

To troubleshoot model errors:
1. Check the model's status in the project status page
2. Review the error message for specific details about what failed
3. Verify that all referenced tables and views exist and are accessible in prod environment 
   1. Run the following from the CLI to see list of tables and size usage: `rill project tables --project <project_nane>`
4. Verify Credentials are correct for prod environment.

If after going through the above steps, you are still unable to resolve the issue, [contact us!](/contact)

## Metrics View / Dashboard Errors

Metrics view and dashboard errors often stem from issues with the underlying models or configuration problems. Common issues include:

- **Model Dependencies:** Dashboards failing because their underlying models have errors
- **Missing Dimensions/Measures:** References to fields that don't exist in the underlying model

To resolve metrics view and dashboard errors:

1. Verify that all referenced models are building successfully
2. Check that dimensions and measures reference valid fields