---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 39
---

In your Rill project directory, create a `<theme_name>.yaml` file in any directory containing `type: theme`. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud.

To apply that theme to a dashboard, add `theme: <name of theme>` to the yaml file for that dashboard. Alternatively, you can add this to the end of the URL in your browser: `?theme=<name of theme>`

:::tip Modern Theming with Light and Dark Modes
The new theming system supports separate `light` and `dark` mode customization with extensive CSS variable control. This allows you to define distinct color palettes for each mode and customize UI components, borders, backgrounds, and chart colors.
:::


## Properties

### `type`

_[string]_ - Refers to the resource type and must be `theme` _(required)_

### `colors`

_[object]_ - **DEPRECATED**: Use `light` and `dark` properties instead. Legacy color palette for the theme. Cannot be used together with `light` or `dark` properties. 

  - **`primary`** - _[string]_ - Primary color (hex, named color, or hsl format) 

  - **`secondary`** - _[string]_ - Secondary color (hex, named color, or hsl format) 

### `light`

_[object]_ - Color customization for light mode. Supports CSS color values (hex, named colors, hsl, etc.) 

  - **`primary`** - _[string]_ - Primary theme color 

  - **`secondary`** - _[string]_ - Secondary theme color 

  - **`surface`** - _[string]_ - Surface color 

  - **`background`** - _[string]_ - Background color 

  - **`color-sequential-1`** - _[string]_ - Sequential palette color 1 (lightest) 

  - **`color-sequential-2`** - _[string]_ - Sequential palette color 2 

  - **`color-sequential-3`** - _[string]_ - Sequential palette color 3 

  - **`color-sequential-4`** - _[string]_ - Sequential palette color 4 

  - **`color-sequential-5`** - _[string]_ - Sequential palette color 5 (medium) 

  - **`color-sequential-6`** - _[string]_ - Sequential palette color 6 

  - **`color-sequential-7`** - _[string]_ - Sequential palette color 7 

  - **`color-sequential-8`** - _[string]_ - Sequential palette color 8 

  - **`color-sequential-9`** - _[string]_ - Sequential palette color 9 (darkest) 

  - **`color-diverging-1`** - _[string]_ - Diverging palette color 1 

  - **`color-diverging-2`** - _[string]_ - Diverging palette color 2 

  - **`color-diverging-3`** - _[string]_ - Diverging palette color 3 

  - **`color-diverging-4`** - _[string]_ - Diverging palette color 4 

  - **`color-diverging-5`** - _[string]_ - Diverging palette color 5 

  - **`color-diverging-6`** - _[string]_ - Diverging palette color 6 (neutral) 

  - **`color-diverging-7`** - _[string]_ - Diverging palette color 7 

  - **`color-diverging-8`** - _[string]_ - Diverging palette color 8 

  - **`color-diverging-9`** - _[string]_ - Diverging palette color 9 

  - **`color-diverging-10`** - _[string]_ - Diverging palette color 10 

  - **`color-diverging-11`** - _[string]_ - Diverging palette color 11 

  - **`color-qualitative-1`** - _[string]_ - Qualitative palette color 1 

  - **`color-qualitative-2`** - _[string]_ - Qualitative palette color 2 

  - **`color-qualitative-3`** - _[string]_ - Qualitative palette color 3 

  - **`color-qualitative-4`** - _[string]_ - Qualitative palette color 4 

  - **`color-qualitative-5`** - _[string]_ - Qualitative palette color 5 

  - **`color-qualitative-6`** - _[string]_ - Qualitative palette color 6 

  - **`color-qualitative-7`** - _[string]_ - Qualitative palette color 7 

  - **`color-qualitative-8`** - _[string]_ - Qualitative palette color 8 

  - **`color-qualitative-9`** - _[string]_ - Qualitative palette color 9 

  - **`color-qualitative-10`** - _[string]_ - Qualitative palette color 10 

  - **`color-qualitative-11`** - _[string]_ - Qualitative palette color 11 

  - **`color-qualitative-12`** - _[string]_ - Qualitative palette color 12 

  - **`color-qualitative-13`** - _[string]_ - Qualitative palette color 13 

  - **`color-qualitative-14`** - _[string]_ - Qualitative palette color 14 

  - **`color-qualitative-15`** - _[string]_ - Qualitative palette color 15 

  - **`color-qualitative-16`** - _[string]_ - Qualitative palette color 16 

  - **`color-qualitative-17`** - _[string]_ - Qualitative palette color 17 

  - **`color-qualitative-18`** - _[string]_ - Qualitative palette color 18 

  - **`color-qualitative-19`** - _[string]_ - Qualitative palette color 19 

  - **`color-qualitative-20`** - _[string]_ - Qualitative palette color 20 

  - **`color-qualitative-21`** - _[string]_ - Qualitative palette color 21 

  - **`color-qualitative-22`** - _[string]_ - Qualitative palette color 22 

  - **`color-qualitative-23`** - _[string]_ - Qualitative palette color 23 

  - **`color-qualitative-24`** - _[string]_ - Qualitative palette color 24 

### `dark`

_[object]_ - Color customization for dark mode. Supports CSS color values (hex, named colors, hsl, etc.) 

  - **`primary`** - _[string]_ - Primary theme color 

  - **`secondary`** - _[string]_ - Secondary theme color 

  - **`surface`** - _[string]_ - Surface color 

  - **`background`** - _[string]_ - Background color 

  - **`color-sequential-1`** - _[string]_ - Sequential palette color 1 (lightest) 

  - **`color-sequential-2`** - _[string]_ - Sequential palette color 2 

  - **`color-sequential-3`** - _[string]_ - Sequential palette color 3 

  - **`color-sequential-4`** - _[string]_ - Sequential palette color 4 

  - **`color-sequential-5`** - _[string]_ - Sequential palette color 5 (medium) 

  - **`color-sequential-6`** - _[string]_ - Sequential palette color 6 

  - **`color-sequential-7`** - _[string]_ - Sequential palette color 7 

  - **`color-sequential-8`** - _[string]_ - Sequential palette color 8 

  - **`color-sequential-9`** - _[string]_ - Sequential palette color 9 (darkest) 

  - **`color-diverging-1`** - _[string]_ - Diverging palette color 1 

  - **`color-diverging-2`** - _[string]_ - Diverging palette color 2 

  - **`color-diverging-3`** - _[string]_ - Diverging palette color 3 

  - **`color-diverging-4`** - _[string]_ - Diverging palette color 4 

  - **`color-diverging-5`** - _[string]_ - Diverging palette color 5 

  - **`color-diverging-6`** - _[string]_ - Diverging palette color 6 (neutral) 

  - **`color-diverging-7`** - _[string]_ - Diverging palette color 7 

  - **`color-diverging-8`** - _[string]_ - Diverging palette color 8 

  - **`color-diverging-9`** - _[string]_ - Diverging palette color 9 

  - **`color-diverging-10`** - _[string]_ - Diverging palette color 10 

  - **`color-diverging-11`** - _[string]_ - Diverging palette color 11 

  - **`color-qualitative-1`** - _[string]_ - Qualitative palette color 1 

  - **`color-qualitative-2`** - _[string]_ - Qualitative palette color 2 

  - **`color-qualitative-3`** - _[string]_ - Qualitative palette color 3 

  - **`color-qualitative-4`** - _[string]_ - Qualitative palette color 4 

  - **`color-qualitative-5`** - _[string]_ - Qualitative palette color 5 

  - **`color-qualitative-6`** - _[string]_ - Qualitative palette color 6 

  - **`color-qualitative-7`** - _[string]_ - Qualitative palette color 7 

  - **`color-qualitative-8`** - _[string]_ - Qualitative palette color 8 

  - **`color-qualitative-9`** - _[string]_ - Qualitative palette color 9 

  - **`color-qualitative-10`** - _[string]_ - Qualitative palette color 10 

  - **`color-qualitative-11`** - _[string]_ - Qualitative palette color 11 

  - **`color-qualitative-12`** - _[string]_ - Qualitative palette color 12 

  - **`color-qualitative-13`** - _[string]_ - Qualitative palette color 13 

  - **`color-qualitative-14`** - _[string]_ - Qualitative palette color 14 

  - **`color-qualitative-15`** - _[string]_ - Qualitative palette color 15 

  - **`color-qualitative-16`** - _[string]_ - Qualitative palette color 16 

  - **`color-qualitative-17`** - _[string]_ - Qualitative palette color 17 

  - **`color-qualitative-18`** - _[string]_ - Qualitative palette color 18 

  - **`color-qualitative-19`** - _[string]_ - Qualitative palette color 19 

  - **`color-qualitative-20`** - _[string]_ - Qualitative palette color 20 

  - **`color-qualitative-21`** - _[string]_ - Qualitative palette color 21 

  - **`color-qualitative-22`** - _[string]_ - Qualitative palette color 22 

  - **`color-qualitative-23`** - _[string]_ - Qualitative palette color 23 

  - **`color-qualitative-24`** - _[string]_ - Qualitative palette color 24 

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
# Example: Modern theme with light and dark modes
type: theme
light:
    primary: "#3b82f6"
    secondary: "#8b5cf6"
    background: "#ffffff"
    foreground: "#0f172a"
    color-sequential-1: "#dbeafe"
    color-sequential-5: "#3b82f6"
    color-sequential-9: "#1e3a8a"
dark:
    primary: "#60a5fa"
    secondary: "#a78bfa"
    background: "#0f172a"
    foreground: "#f8fafc"
    color-sequential-1: "#1e3a8a"
    color-sequential-5: "#3b82f6"
    color-sequential-9: "#dbeafe"
```

```yaml
# Example: Legacy theme (deprecated, use light/dark instead)
type: theme
colors:
    primary: plum
    secondary: violet
```