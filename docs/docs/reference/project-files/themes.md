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

_[object]_ - Color palette for the theme 

  - **`primary`** - _[string]_ - Primary color 

  - **`secondary`** - _[string]_ - Secondary color 

### `light`

_[object]_ - Light theme color configuration 

  - **`primary`** - _[string]_ - Primary color for light theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

  - **`secondary`** - _[string]_ - Secondary color for light theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

  - **`variables`** - _[object]_ - Custom CSS variables for light theme 

### `dark`

_[object]_ - Dark theme color configuration 

  - **`primary`** - _[string]_ - Primary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

  - **`secondary`** - _[string]_ - Secondary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats. 

  - **`variables`** - _[object]_ - Custom CSS variables for dark theme 

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
# Example: Basic theme with light and dark mode colors
type: theme
light:
    primary: "#4F46E5" # Indigo-600
    secondary: "#8B5CF6" # Purple-500
dark:
    primary: "#818CF8" # Indigo-400
    secondary: "#A78BFA" # Purple-400
```

```yaml
# Example: Advanced theme with custom color palettes
type: theme
light:
    primary: "#14B8A6" # Teal
    secondary: "#10B981" # Emerald
    variables:
        color-sequential-1: "hsl(180deg 80% 95%)"
        color-sequential-5: "hsl(180deg 80% 50%)"
        color-sequential-9: "hsl(180deg 80% 25%)"
        color-qualitative-1: "hsl(156deg 56% 52%)"
        color-qualitative-2: "hsl(27deg 100% 65%)"
dark:
    primary: "2DD4BF"
    secondary: "34D399"
    variables:
        color-sequential-1: "hsl(180deg 40% 30%)"
        color-sequential-5: "hsl(180deg 50% 50%)"
        color-sequential-9: "hsl(180deg 60% 70%)"
```