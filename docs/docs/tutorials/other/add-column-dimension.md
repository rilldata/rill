---
title: "How do I add a new dimension to my dashboard?"
sidebar_label: "Add/Modify Dimension to Dashboard via YAML"
sidebar_position: 10
hide_table_of_contents: false
tags:
  - OLAP:DuckDB
  - OLAP:ClickHouse
---

Changes to the underlying data may require modification to your SQL / Dashboard in order to start visualizing your new data. 

You will need to add the dimension/s back in Rill Developer and [push the changes to Rill Cloud](/tutorials/rill_advanced_features/advanced_developer/update-rill-cloud) when you're ready.

## Rill Developer

Depending on whether your dashboard is created directly from your source or via a model (check if your dashboard as a `table:` or `model:` it it's YAML), you will need to check the corresponding file to make the changes.

### Checking the Sources

You need to check that the new column is being ingested into Rill. If you have a select * statement, go ahead and select the `refresh` button and confirm that the new column is listed. If not, add the new column into your select statement and select `refresh`.

If your dashboard is created directly from the source, navigate to [adding the new dimension to the Dashboard](#adding-the-new-dimension-to-the-dashboard). If not, continue to the model to make changes.

![source](/img/tutorials/other/source-new-dimension.png)

---

### Dashboard created from a Model

After confirming that the sources have ingested the new data, you can [modify the model to include these new dimensions / measures](https://docs.rilldata.com/build/models/). If you need to make any transformations, you can do so here or [in the dashboard layer](https://docs.rilldata.com/build/dashboards/expressions).


![model](/img/tutorials/other/model-new-dimension.png)

---
### Adding the new Dimension to the Dashboard

Finally, you can add the dimension / measure to the dashboard. Notice in the right panel, your newly created dimension or metric can be seen in the right panel. 

![dashboard](/img/tutorials/other/dashboard-new-dimension.png)

## Rill Cloud

### Pushing Changes
Once you have finished making all the changes to your dashboard you can [push the new changes to Rill Cloud](/tutorials/rill_advanced_features/advanced_developer/update-rill-cloud) either via the UI by selecting `Update` or via GitHub by pushing the changes to your repository.

<img src = '/img/tutorials/other/redeploy.gif' class='rounded-gif' />
<br />

### Refreshing Sources
If you are not seeing the new dimension in your dashboard, please refresh the source or model.

```bash
rill project refresh --source <your_source> 
rill project refresh --model <your_source> --full
```