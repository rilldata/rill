---
title: "Deploy to Cloud!"
sidebar_label: "Deploy To Cloud"
sidebar_position: 4
hide_table_of_contents: false
---

## Time to share our dashboard!

At this point, we would normally be ready to ship our dashboard off to Rill Cloud and share the dashboard. Just one problem, and quite a big one if you use a local ClickHouse server to run some testing. Since this binary is locally running on your machine, if you try to deploy as is, you will not be able to connect to the ClickHouse database. 

If you're already using ClickHouse Cloud, [skip to the deploy](#now-we-can-deploy)!

### Let's modify the ClickHouse Database credentials.

If you haven't already, we will need to go to ClickHouse's page and create a [Clickhouse Cloud](https://clickhouse.com/cloud) account. They have a 30 day free trial, so it shouldn't cost you a penny!


Now that you've set it up, we can modify the credentials based on the `connect` page.

Please remove the other clickhouse entries from your `.env` file and add the following, replacing with your account information.
```
connector.clickhouse.dsn="https://<hostname>:<port>?username=<username>&password=<password>&secure=true&skip_verify=true"
```

You will need to head back to [Conecting to ClickHouse](components/r_ch_4) to reinstall the uk_price_paid dashboard so that when you deploy to Cloud your dashboard is still there!


### Now we can deploy!

Sorry about that, now we can deploy our project to Rill Cloud!