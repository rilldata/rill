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

### `light`

_[object]_ - Light theme color configuration

  - **`primary`** - _[string]_ - Primary color for light theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`secondary`** - _[string]_ - Secondary color for light theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`kpi-positive`** - _[string]_ - Color for positive KPI delta values in light theme. Defaults to fg-secondary (gray).

  - **`kpi-negative`** - _[string]_ - Color for negative KPI delta values in light theme. Defaults to red.

  - **`variables`** - _[object]_ - Custom CSS variables for light theme

### `dark`

_[object]_ - Dark theme color configuration

  - **`primary`** - _[string]_ - Primary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`secondary`** - _[string]_ - Secondary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`kpi-positive`** - _[string]_ - Color for positive KPI delta values in dark theme. Defaults to fg-secondary (gray).

  - **`kpi-negative`** - _[string]_ - Color for negative KPI delta values in dark theme. Defaults to red.

  - **`variables`** - _[object]_ - Custom CSS variables for dark theme

### `font`

_[object]_ - Font for the text of the dashboard, including its charts, filter bar and AI panel. Rill Cloud's own navigation outside the dashboard keeps Rill's font.

  - **`family`** - _[string]_ - A single font family name such as `Brand Sans`, or a generic family such as `system-ui`. Rill falls back to its default font while the font loads or if it is unavailable. _(required)_

  - **`files`** - _[array of object]_ - Font files that provide the family. Without files, the font must be installed on the viewer's device.

    - **`path`** - _[string]_ - Path of a `.woff2`, `.woff`, `.ttf` or `.otf` file, relative to the project root. It must be inside the project's `public` folder (or another of the `public_paths` in `rill.yaml`). _(required)_

    - **`weight`** - _[string, integer]_ - Font weight the file provides, from 1 to 1000. Use a range such as `100 900` for a variable font. Defaults to `400`.

    - **`style`** - _[string]_ - Font style the file provides. Defaults to `normal`.

### `chrome`

_[object]_ - Border and font tokens for the dashboard's chrome, which is the navigation bar of an embedded dashboard, the filter bar, the filter pills and the AI panel. Their colors are set in `light` and `dark`.

  - **`border_width`** - _[string, integer]_ - Width of the chrome's borders, from `0px` to `8px`. Defaults to `1px`.

  - **`font_size`** - _[string, integer]_ - Font size of the chrome's labels, from `10px` to `20px`. Defaults to each label's own size.

  - **`font_weight`** - _[string, integer]_ - Font weight of the chrome's labels, from 100 to 900. Defaults to each label's own weight.

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually.

### `refs`

_[array of string]_ - List of resource references

### `tags`

_[array of string]_ - Tags for organizing and filtering the resource (e.g. on the project dashboards list).

### `metadata`

_[object]_ - User-defined key-value metadata attached to the resource. Rill does not read or write it; it is exposed as-is on the resource's meta over the API for external tooling. Values are strings, with numbers and booleans coerced. Not supported in rill.yaml defaults.

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
    color-sequential-1: "hsl(180deg 80% 95%)"
    color-sequential-5: "hsl(180deg 80% 50%)"
    color-sequential-9: "hsl(180deg 80% 25%)"
    color-qualitative-1: "hsl(156deg 56% 52%)"
    color-qualitative-2: "hsl(27deg 100% 65%)"
dark:
    primary: "2DD4BF"
    secondary: "34D399"
    color-sequential-1: "hsl(180deg 40% 30%)"
    color-sequential-5: "hsl(180deg 50% 50%)"
    color-sequential-9: "hsl(180deg 60% 70%)"
```

```yaml
# Example: Theme with KPI delta colors (green for positive, red for negative)
type: theme
light:
    primary: "#6366f1"
    secondary: "#8b5cf6"
    kpi-positive: "#16a34a" # Green for positive deltas (default: gray)
    kpi-negative: "#dc2626" # Red for negative deltas (default: red)
dark:
    primary: "#818cf8"
    secondary: "#a78bfa"
    kpi-positive: "#4ade80" # Brighter green for dark mode
    kpi-negative: "#f87171" # Brighter red for dark mode
```

```yaml
# Example: Theme with a brand font loaded from the project's public folder
type: theme
font:
    family: Brand Sans
    files:
        - path: public/BrandSans-Regular.woff2
          weight: 400
        - path: public/BrandSans-Bold.woff2
          weight: 700
chrome:
    border_width: 2px
    font_weight: 600
light:
    primary: "#4F46E5"
    secondary: "#8B5CF6"
```
