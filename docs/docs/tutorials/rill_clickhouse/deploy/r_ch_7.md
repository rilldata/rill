---
title: Deploy to Rill Cloud
sidebar_label: 'via CLI'
sidebar_position: 9
hide_table_of_contents: false
---
## Deploy via CLI

Let's head back to the Terminal to run the following, you will be prompted to login/signup if you have not already done so. Please follow the instructions in the UI then return back to the CLI when you are finished.

```
rill deploy


No git remote was found.
You can connect to Github or use one-time uploads to deploy your project.
? Do you want to use one-time uploads? (Y/n) 
```

You'll see two options appear in the UI:

1. Deploying with GitHub
2. One-time upload.

Depending if you want to deploy with GitHub vs upload it once, please select Y or N.

```
? Do you want to use one-time uploads? Yes
Using org "Rill_Learn".

Starting upload.
All files uploaded successfully.

Created project "Rill_tutorials/my-rill-clickhouse". Use `rill project rename` to change name if required.


Could not access all connectors. Rill requires credentials for the following connectors:

 - clickhouse (used by uk_price_paid_dashboard and others)
 - s3 (used by sn-ch-incre and others)
 - snowflake (used by sn-ch-incre and others)

Run `rill env configure --project /my-rill-clickhouse"l` to provide credentials.

Your project can be accessed at: https://ui.rilldata.com/Rill_tutorials//my-rill-clickhouse"
Opening project in browser...
```
Since we have already added and modified the connection credentials in the .env file, you can run the following commands to push your credentials to Rill Cloud.

```
rill env push
```

Congratulations! You've successfully deployed your project to Rill Cloud!
