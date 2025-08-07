---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 40
---

In your Rill project directory, create a `<theme_name>.yaml` file in any directory containing `type: theme`. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud.

To apply that theme to a dashboard, add `default_theme: <name of theme>` to the yaml file for that dashboard. Alternatively, you can add this to the end of the URL in your browser: `?theme=<name of theme>`


## Properties

### `type`

_[string]_ - Refers to the resource type and must be `theme` _(required)_

### `colors`

_[object]_ - Color palette for the theme 

  - **`primary`** - _[string]_ - Primary color 

  - **`secondary`** - _[string]_ - Secondary color 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 

## Examples

```yaml
# Example: You can copy this directly into your <theme_name>.yaml file
type: theme

colors:
  primary: plum
  secondary: violet
```
