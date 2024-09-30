---
title: "Incremental Models"
description:  "Start with basics"
sidebar_label: "Basic Incremental Models"
sidebar_position: 1
---
Before enabing incremental on the model, let's take a look at the following model YAML file, now.yaml. 
### Sample YAML 
```yaml
# This model outputs the current time every time it is refreshed.
type: model
refresh:
  cron: 0 0 * * *
sql: SELECT now() AS inserted_on
```

To understand what this is doing, let's try to run some `rill project refresh` commands locally from the CLI. (Mention something here about new UI coming for project refresh and will update ASAP.) Since we're using Rill Developer, we will need to add the `--local` flag to the refresh commands


```bash
rill project refresh --model now --local
```

When running the above command, you should see that the inserted_on column value changes to the value of now(). 

![img](/img/tutorials/302/now.png)


### Enable Increments on our Model 

As mentioned previously, the `incremental: true` tells Rill that this model is incremental. 

```yaml
type: model
refresh:
  cron: 0 0 * * *
sql: SELECT now() AS inserted_on
incremental: true
```

After adding the following, let's run the same command. What's different?

![img](/img/tutorials/302/now-incremental.png)


Instead of overwriting the same row, we are now appending the new values of now() into the table. Next, let's take a moment to review Splits. With the refresh enabled to run at midnight on every night, you should see the amount of rows increase each day at midnight UTC.


### States in Models

Lastly, we can add a `state:` key that allows us to manually define some sort of state to use as our incrementing key.


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

Along with the inserted_on column, we are also creating a val column that defaults to 0. Then on each run, if incremental, increases this value. The state retrieves the max value of 'val' as max_val.

```SQL
SELECT 
{{ if incremental }} 
        {{ .state.max_val }} + 1 {{ else }} 0 
{{ end}}
```

![img](/img/tutorials/302/now-state.png)