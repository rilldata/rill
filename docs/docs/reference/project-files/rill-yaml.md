---
title: Project YAML
sidebar_label: Project YAML
sidebar_position: 40
---

The `rill.yaml` file contains metadata about your project.

## Properties

- _**`title`**_ — the name of your project which will be displayed in the upper left hand corner
- _**`compiler`**_ — the Rill project compiler version compatible with your project files (currently only supports: `rill-beta`)
- _**`mock_users`**_ — a list of mock users to test against dashboard [security policies](../../develop/security). For each mock user, possible attributes include:
  - _**`email`**_ — the mock user's email _(required)_
  - _**`name`**_ — the mock user's name
  - _**`admin`**_ — whether or not the mock user is an admin
 
## Project wide defaults

In `rill.yaml` you can specify project wide defaults that will be applied for all project files within a project.  

The individual project files will inherit any defaults that has been specified in `rill.yaml`. If the same property would be set in both `rill.yaml` and a specific project file the local setting in the project file would win. See the documentation for each individual project file for available properties.

The top level property is the project file name in pluralis such as `models`, `dashboards` and `sources`.  
Example:
```
title: My project
dashboards:
  first_day_of_week: 7
  available_time_zones:
    - America/New_York
models:
  materialize: true
