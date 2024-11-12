---
title: "Incremental Models"
description:  "Start with basics"
sidebar_label: "Basic Incremental Models"
sidebar_position: 1
---
Before enabling incremental on the model, let's take a look at the following model YAML file, now.yaml. You'll notice that this is not our typical SQL model where we can write SQL statements into the textfile and have these automatically run. Instead, model YAML files requires a bit more setup such as defining the type: model, and the sql: parameter.
## Getting Started 

```yaml
# This model outputs the current time every time it is refreshed.
type: model
refresh:
  cron: 0 0 * * *
sql: SELECT now() AS inserted_on
```


To understand what this is doing, let's go ahead and select the refresh button as seen in the screenshot below. This button performs the same command as the below in the CLI.

```bash
rill project refresh --model now --local
```
:::note
Since we're using Rill Developer, we will need to add the `--local` flag to the refresh commands or else this will refresh the project on Rill Cloud!
:::

After the model refreshes, you should see the inserted_on value change.

![img](/img/tutorials/302/now.png)


## Enable Increments on our Model 

As mentioned previously, the `incremental: true` tells Rill that this model is an incremental model. You will see that the UI changes slightly when this is enabled. Not only will you be able to full refresh, but also incrementally refresh.

```yaml
type: model
refresh:
  cron: 0 0 * * *
sql: SELECT now() AS inserted_on
incremental: true
```

After adding the following, what's different?

When selecting the refrehs button, a new drop down appears. In this case, we have the choice to incremental refresh or full refresh.

![img](/img/tutorials/302/now-incremental.png)

Instead of overwriting the same row, we are now appending the new values of now() into the table. With the refresh enabled to run at midnight on every night, you should see the amount of rows increase each day at midnight UTC. 

Running the incremental refresh is the same as the following command in the CLI:

```
rill project refresh --model now_incremental --local
```

If you want to perform a full refresh you'll need to add the `--full` flag.

```
rill project refresh --model now_incremental --local --full
```


Next, let's take a moment to review `states:`. 


## States in Models

Lastly, we can add a `state:` key that allows us to manually define some sort of state .


```yaml
#filename: incremental_state.yaml
type: model
refresh:
  cron: 0 0 * * *
sql: SELECT {{ if incremental }} {{ .state.max_val }} + 1 {{ else }} 0 {{ end}} AS val, now() AS inserted_on
state:
  sql: SELECT MAX(val) as max_val FROM incremental_state
incremental: true
```
:::note
In more realistic cases, we could select the MAX(time_stamp) which will grab the latest time_stamp that your current model includes. Then, based on this it would incrementally refresh your model to only insert the new data. However, keep in mind that any 'old' data that gets added outside of Rill would not be detected.
:::
Along with the inserted_on column, we are also creating a val column that defaults to 0. Then on each run, if incremental, increases this value. The state retrieves the max value of 'val' as max_val.

```SQL
SELECT 
{{ if incremental }} 
        {{ .state.max_val }} + 1 {{ else }} 0 
{{ end}}
```

![img](/img/tutorials/302/now-state.png)



import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />