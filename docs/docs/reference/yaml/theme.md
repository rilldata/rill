---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 39
---

In your Rill project directory, create a `<theme_name>.yaml` file in any directory containing `type: theme`. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud.

To apply that theme to a dashboard, add `default_theme: <name of theme>` to the yaml file for that dashboard. Alternatively, you can add this to the end of the URL in your browser: `?theme=<name of theme>`


## Properties

### `type`

_[string]_ - Refers to the resource type and must be `theme` _(required)_

### `colors`

_[object]_ - Used to override the dashboard colors. Either primary or secondary color must be provided. _(required)_

  - **`primary`** - _[string]_ - Overrides the primary blue color in the dashboard. Can have any hex (without the '#' character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. Note that the hue of the input colors is used for variants but the saturation and lightness is copied over from the [blue color palette](https://tailwindcss.com/docs/customizing-colors). 

  - **`secondary`** - _[string]_ - Overrides the secondary color in the dashboard. Applies to the loading spinner only as of now. Can have any hex (without the '#' character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

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