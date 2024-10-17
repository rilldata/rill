---
title: "1. Launch Rill Developer"
sidebar_label: "1. Launch Rill Developer"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:ClickHouse
---

:::note prerequisites
You will need to [install Rill](https://docs.rilldata.com/home/install).

You need access to either a [locally running ClickHouse Server](https://clickhouse.com/docs/en/install) or [ClickHouse Cloud](https://docs.rilldata.com/reference/olap-engines/clickhouse#connecting-to-clickhouse-cloud). We recommend using ClickHouse Cloud as this will make deploying to Rill Cloud easier. Please review the documentation, [here](https://docs.rilldata.com/reference/olap-engines/clickhouse).
:::
## Start Rill Developer

```yaml
rill start my-rill-clickhouse
```

After running the command, Rill Developer should automatically open in your default browser. If not, you can access it via the following url:

```
localhost:9009
``` 

You should see the folowing webpage appear. 

![my-rill-project](/img/tutorials/101/new-rill-project.png)
<br />

Let's go ahead and select `Start with an empty project`.

<details>
  <summary>Where am I in the terminal?</summary>
  
    You can use the `pwd` command to see which directory in the terminal you are. <br />
    If this is not where you'd like to make the directory use the `cd` command to change directories.

</details>


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />
