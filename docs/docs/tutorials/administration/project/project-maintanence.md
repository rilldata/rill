---
title: "Project Maintanence"
description: Project Maintanence
sidebar_label: "Project Maintanence"
tags:
  - CLI
  - Administration
---

import ComingSoon from '@site/src/components/ComingSoon';

Changes to the project information, such as title and description is only modifable in the CLI. 

To change the name of your project:
```bash
rill project rename
? Select project to rename my-rill-tutorial
? Enter the New Project Name my-Rill-tutorial
? Do you want to rename the project "my-rill-tutorial" to "my-Rill-tutorial"? Yes
Renamed project
New web url is: https://ui.rilldata.io/Rill_Tutorial/my-Rill-tutorial
  NAME               PUBLIC   GITHUB                                        ORGANIZATION   
 ------------------ -------- --------------------------------------------- --------------- 
  my-Rill-tutorial   No       https://github.com/royendo/my-rill-tutorial   Rill_Tutorial
```

To change the description, branch, and public access:
```bash
rill project edit

? Select project my-rill-tutorial
? Enter the description A project that follows the Rill Tutorials
? Enter the production branch main
? Make project public No
```


## Status

The Status Page gives us an overview of all the components within Rill Cloud, including the underlying source and models. While you will not be able to make any direct changes, the Status page is a good place to start when dashboards are acting strange.

![img](/img/tutorials/203/status.png)

You'll see here that there's an option to connect to GitHub.
During our first deployment onto Rill Cloud, we opted for a one-time upload. By doing so, we are able to directly deploy the project without any further steps, but we lose out on a few powerful capablities that can enhance the user experience, such as version control.


### When a dashboard is failing to load

When a dashboard fails to load, you will see an `Error` in the UI. There are a few potential causes for a dashboard to fail to load, but the best place to start is the Status page. For example, you might see the following in the UI: 

![img](/img/tutorials/admin/failing-dashboard.png)

In order to understand why this is failing, you can navigate to the Status page and find the dashboard's error message:

![img](/img/tutorials/admin/failing-status-page.png)

In this case, we can find that the table, `staging_to_CH` does not exist! We can see that this table fails to create due to the following error:

```bash
connection: dial tcp 127.0.0.1:9000: connect: connection refused
```

Seeing as this is ClickHouse model, it is likely that the credentials or connections are not correct for this connection. 

Whether it's the source or the model that is erroring and causing the dashboard to fail, you may need to [check the credentials](credential-env-variable-management.md) back in Rill Developer.


### Incremental Models are failing 

Additionally, you may need to troubleshoot your incremental model's splits. As seen in the above image, our model, S3-incremental, is erroring with the following:

```bash
failed to sync splits: blob (code=Unknown): AccessDenied: Access Denied status code:...
```

Depending on the error, you might not be able to determine the cause of the issue and will need to return to the CLI to check each split. This can be done by running:

```bash
rill project splits <name_of_model> --project <your-project>
```

```
 KEY (50)                           DATA                                                                                                                                                                 EXECUTED ON            ELAPSED   ERROR  
 ---------------------------------- -------------------------------------------------------------------------------------------------------------------------------------------------------------------- ---------------------- --------- ------- 
  39b3f0a233c2ac07897fa07a1d823437   {"path":"github-analytics/Clickhouse/2018/03/commits_2018_03.parquet","uri":"s3://rilldata-public-s3/github-analytics/Clickhouse/2018/03/commits_2018_03.parquet"}   2024-08-28T04:50:06Z   1.216s           
  4fd7e315006050de98adf812fa830036   {"path":"github-analytics/Clickhouse/2018/04/commits_2018_04.parquet","uri":"s3://rilldata-public-s3/github-analytics/Clickhouse/2018/04/commits_2018_04.parquet"}   2024-08-28T04:50:08Z   1.268s           
  962cfd66eabc5edf8def23ef5397ddf6   {"path":"github-analytics/Clickhouse/2018/05/commits_2018_05.parquet","uri":"s3://rilldata-public-s3/github-analytics/Clickhouse/2018/05/commits_2018_05.parquet"}   2024-08-28T04:50:09Z   1.072s           
  da86ef9b082147b580145b6590e300e2   {"path":"github-analytics/Clickhouse/2018/06/commits_2018_06.parquet","uri":"s3://rilldata-public-s3/github-analytics/Clickhouse/2018/06/commits_2018_06.parquet"}   2024-08-28T04:50:10Z   1.104s           
```

In this case, the issue was a permission issue on the file in S3, and was resolved by setting the file to public within the S3 bucket.

In order to refresh the full model, you can run the following:

```bash
rill project refresh --model <your_model> --full
```

If you notice that it is only a specific split that is broken you can use the KEY to refresh that specific split.
```bash
rill project refresh --model <your_model> --split SPLIT_KEY
```

### Data is not up to date

While you may have set up source refresh to automatically ingest new data, seen on the last column of the status page, there might be times where you are unable to view the new data due to external factors or you have updated the underlying data and dont want to wait for the next refresh. In these cases, you will want to run a project refresh to ingest the data again.

```bash
rill project refresh --all --full
```
This triggers a refresh of the full project and you should see all the last refresh dates change. You can either check this in Rill Cloud or by running the following:

```bash
rill project status --project <your-project-id>
```

If you only want to refresh a specific component, you can do so with the required flag `--source` or `--model`.
```bash
rill project refresh --source <your_source> [--model <your-model>]
```


### Parse Errors



## Settings
An additional page for administrators to manage objects in Rill Cloud.

### Public URL Management

Along with the CLI, you can also view and manage the public URLs from the Settings page. As an administrator, you can re-copy the URL or delete the public URL.
![img](/img/tutorials/admin/settings-public-url.png)

