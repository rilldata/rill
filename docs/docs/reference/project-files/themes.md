---
title: Theme YAML
sidebar_label: Theme YAML
sidebar_position: 41
---

In your Rill project directory, create a `<theme_name>.yaml` file in any directory containing `kind: theme`. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud.

To apply that theme to a dashboard, add `default_theme: <name of theme>` to the yaml file for that dashboard. Alternatively, you can add this to the end of the URL in your browser: `?theme=<name of theme>`

## Properties

_**`kind`**_ - is mandatory and must be `theme`. _(required)_

_**`colors`**_ - used to override the dashboard colors.
  - _**`primary`**_ - overrides the primary blue color in the dashboard. Can have any hex (without the # character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. Note that the hue of the input colors is used for variants but the saturation and lightness is copied over from the [blue color palette](https://tailwindcss.com/docs/customizing-colors).
  - _**`secondary`**_ - overrides the secondary color in the dashboard. Applies to the loading spinner only as of now. Can have any hex (without the # character), [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

## Example
You can copy this directly into your <theme_name>.yaml file:
```yaml
kind: theme
colors:
  primary: crimson 
  secondary: lime 
```
