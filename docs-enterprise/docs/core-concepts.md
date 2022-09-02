---
title: "🧰 Core Concepts"
slug: "core-concepts"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Rill Cloud Console to manage data; Rill Explore to analyze & share"/>
Rill enables you to leverage all the power of Druid with a serverless cloud service that is simple, secure, and elastic. Rill is designed to fit into your existing analytics ecosystem. You can read from a wide variety of streamed or batch data sources such as Kafka and Big Query, and you can perform analytics using industry standard tools such as Tableau and Looker.

As such, this section introduces the core concepts related to managing your datasources, querying your data and terminology around application integration.

## Rill Developer

Rill Developer is our open source tool that makes it effortless to transform your datasets with SQL. Rill Developer follows a few guiding principles:

  * no more data analysis "side-quests" – helps you build intuition about your dataset through automatic profiling
  * no "run query" button required – responds to each keystroke by re-profiling the resulting dataset
  * works with your local datasets – imports and exports Parquet and CSV
  * feels good to use – powered by Sveltekit & DuckDB = conversation-fast, not wait-ten-seconds-for-result-set fast 

Learn more at [Rill Developer's Github page](https://github.com/rilldata/rill-developer).

## Rill Cloud Console (RCC)

Rill Cloud Console, RCC, is Rill's console for your Apache Druid cloud database service. This console would be used by admins and technical teams to manage your dataset, ingest data, or query data directly. For existing users, login to RCC at [app.rilldata.com](https://app.rilldata.com). From the console, you'll be able to create and access your team workspaces and from a workspace, a user will see all of the datasets available in that workspace.  

A quick overview of the main components of RCC will help you get started.

### Organization

When your company subscribes to Rill, it is assigned an organization name. Your organization will have one or more administrators. 

### Workspaces

A workspace is a virtual space for a team of users. User permissions are granted at the workspace level. When a user joins a workspace, they have access to all of the datasets in that workspace. 

Workspaces are the perimeter for security. For example, if a workspace is granted permission to read from a BigQuery project, all users in the workspace with Editor privilege will be able to import data from that BigQuery project.

### Datasets

A dataset includes any data that imported from an external datasource such as Google BigQuery, AWS, or a Kafka stream. The import is done to Druid via a process called ingestion. This ingestion process applies optimizations to your data and to the Druid configuration of your data to support fast queries. 

For example, if your original data is at the millisecond, the import may aggregate to the minute, hour or day. Or you may remove high cardinality columns or apply aggregation approximations such as HLL or Sketch to provide aggregative counts of those high cardinality columns. It will also specify segmentation of the data across the physical Druid cluster to optimize access. 

All of these actions are specified at ingestion time, and, if you are in the RCC viewing a dataset, these actions have already been applied to your data. 

### User Types

Within RCC, Users may have "Admin", "Viewer" or "Editor" privilege.

As an Admin, you can create workspaces and invite users to those workspaces. Alternately, you may whitelist everyone in a particular email domain so that they can log in to Rill without an invitation. Admins may also modify workspace logos and set up security features like API keys & Service Accounts.

As an Editor you'll additionally be able to create new datasets, loading from common warehouses such as BigQuery or Amazon S3 or from streaming solutions such as Kafka.

As a Viewer you'll have fast query access via the Druid SQL Console, command line or programmatic APIs, or visualization tools like Tableau. 

### Querying Data

Once your dataset has been created you can query it through a variety of interfaces. 

From within RCC, you can click on Druid Console to query the data from the interactive Druid SQL Console. You can find extensive details on [query concepts and best practices](https://druid.apache.org/docs/latest/querying/querying.html) via the Apache Druid docs.

## Rill Explore
### Dashboards

Rill also provides access to Rill Explore - an easy-to-use interface designed specifically for operational analytics focused on ad hoc data exploration. For existing users, login to Explore at [dash.rilldata.com](https://dash.rilldata.com). For more details on Explore, [visit our Explore docs section](/getting-started).

Dashboards may be defined on any combination of dimensions and metrics (including derived calculations between metrics) within your dataset. Layout is customizable and includes time series, topN, bar chart and heat map views.

Dashboards can be shared internally, to external users or may be [embedded within your own application](/embedding-explore). Many customers take advantage of Parent/Child dashboard relationships for [external facing dashboards](/create-an-external-dashboard) - creating a single view of data that is inherited to each child dashboard which is then filtered by a specific subset of criteria.

### User Types

Within Explore, Users may have "Admin", "Member" or "Guest" privilege. Admins have full control over all dashboard set-up, permissions and users while Members (internal to the company users) would only be able to see their dashboards. Guest users are external to your company domain.

### Security Policies

Each user is tied to a dashboard (or set of dashboards) via a Security Policy. More information on setting up access can be found in [the permissions guide](/admin-security).

### Alerts, Scheduled Exports and Bookmarks

Multiple capabilities exist in Explore to improve your workflow. You can set alerts on common metrics and filter sets for automated notifications, set regular exports of data to your inbox for reporting, or create bookmarks that are saved views of common analyses. More information on each area is available in the [Explore Getting Started guide](/getting-started).
## External Applications

Beyond native console querying, you can connect other BI tools and applications to your dataset to create fast and interactive dashboards and reports. More details can be found in our [application integration docs](/authenticating-integrated-applications). 

You can also query your dataset through API access - this will allow you to connect your own internal tools directly to your Druid database. More details can be found in the [Apache Druid API documentation](https://druid.apache.org/docs/latest/operations/api-reference.html).