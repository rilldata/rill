---
title: "1. Source to Dashboard on Rill Cloud in 6 Steps"
sidebar_label: "1. Launch Rill Developer"
position: 1
collapsed: false
sidebar_position: 1
tags:
  - OLAP:DuckDB
---
:::note prerequisites
You will need to [install Rill](https://docs.rilldata.com/home/install).
:::

By the end of this course, you will have a deployed instance of your project in Rill Cloud, and your [30 day 
trial](./launch) will start. This also prepares you for the [Advanced Features Course](../rill_advanced_features/overview.md) that builds on top of this project.

## Start Rill Developer

```yaml
rill start my-rill-tutorial
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