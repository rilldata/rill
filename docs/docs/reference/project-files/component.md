---
note: GENERATED. DO NOT EDIT.
title: Component YAML
sidebar_position: 41
---

Defines a reusable dashboard component that can be embedded in canvas dashboards

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `component` _(required)_

### `display_name`

_[string]_ - Refers to the display name for the component

### `description`

_[string]_ - Detailed description of the component's purpose and functionality

### `input`

_[array of object]_ - List of input variables that can be passed to the component

  - **`name`** - _[string]_ - Unique identifier for the variable _(required)_

  - **`type`** - _[string]_ - Data type of the variable (e.g., string, number, boolean) _(required)_

  - **`value`** - _[string, number, boolean, object, array]_ - Default value for the variable. Can be any valid JSON value type

### `output`

_[object]_ - Output variable that the component produces

  - **`name`** - _[string]_ - Unique identifier for the variable _(required)_

  - **`type`** - _[string]_ - Data type of the variable (e.g., string, number, boolean) _(required)_

  - **`value`** - _[string, number, boolean, object, array]_ - Default value for the variable. Can be any valid JSON value type

### `line_chart`

_[object]_ - (no description)

  - **`config`** - _[object]_ - (no description) _(required)_

    - **`metrics_view`** - _[string]_ - Reference to the metrics view to use _(required)_

    - **`x`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

    - **`y`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

    - **`color`** - _[oneOf]_ - (no description)

      - **option 1** - _[object]_ - (no description)

        - **`field`** - _[string]_ - Field name from the metrics view _(required)_

        - **`title`** - _[string]_ - Display title for the field

        - **`format`** - _[string]_ - Format string for the field

        - **`type`** - _[string]_ - Data type of the field _(required)_

        - **`timeUnit`** - _[string]_ - Time unit for temporal fields

      - **option 2** - _[string]_ - (no description)

    - **`tooltip`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

  - **`title`** - _[string]_ - Chart title

  - **`description`** - _[string]_ - Chart description

### `bar_chart`

_[object]_ - (no description)

  - **`config`** - _[object]_ - (no description) _(required)_

    - **`metrics_view`** - _[string]_ - Reference to the metrics view to use _(required)_

    - **`x`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

    - **`y`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

    - **`color`** - _[oneOf]_ - (no description)

      - **option 1** - _[object]_ - (no description)

        - **`field`** - _[string]_ - Field name from the metrics view _(required)_

        - **`title`** - _[string]_ - Display title for the field

        - **`format`** - _[string]_ - Format string for the field

        - **`type`** - _[string]_ - Data type of the field _(required)_

        - **`timeUnit`** - _[string]_ - Time unit for temporal fields

      - **option 2** - _[string]_ - (no description)

    - **`tooltip`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

  - **`title`** - _[string]_ - Chart title

  - **`description`** - _[string]_ - Chart description

### `stacked_bar_chart`

_[object]_ - (no description)

  - **`config`** - _[object]_ - (no description) _(required)_

    - **`metrics_view`** - _[string]_ - Reference to the metrics view to use _(required)_

    - **`x`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

    - **`y`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

    - **`color`** - _[oneOf]_ - (no description)

      - **option 1** - _[object]_ - (no description)

        - **`field`** - _[string]_ - Field name from the metrics view _(required)_

        - **`title`** - _[string]_ - Display title for the field

        - **`format`** - _[string]_ - Format string for the field

        - **`type`** - _[string]_ - Data type of the field _(required)_

        - **`timeUnit`** - _[string]_ - Time unit for temporal fields

      - **option 2** - _[string]_ - (no description)

    - **`tooltip`** - _[object]_ - (no description)

      - **`field`** - _[string]_ - Field name from the metrics view _(required)_

      - **`title`** - _[string]_ - Display title for the field

      - **`format`** - _[string]_ - Format string for the field

      - **`type`** - _[string]_ - Data type of the field _(required)_

      - **`timeUnit`** - _[string]_ - Time unit for temporal fields

  - **`title`** - _[string]_ - Chart title

  - **`description`** - _[string]_ - Chart description

### `kpi`

_[object]_ - (no description)

  - **`metrics_view`** - _[string]_ - Reference to the metrics view to use _(required)_

  - **`measure`** - _[string]_ - Measure to display _(required)_

  - **`time_range`** - _[string]_ - Time range for the KPI _(required)_

  - **`comparison_range`** - _[string]_ - Comparison time range

  - **`filter`** - _[string]_ - Filter expression

  - **`title`** - _[string]_ - KPI title

  - **`description`** - _[string]_ - KPI description

### `table`

_[object]_ - (no description)

  - **`metrics_view`** - _[string]_ - Reference to the metrics view to use _(required)_

  - **`measures`** - _[array of string]_ - List of measures to display _(required)_

  - **`time_range`** - _[string]_ - Time range for the table _(required)_

  - **`row_dimensions`** - _[array of string]_ - Dimensions for table rows

  - **`col_dimensions`** - _[array of string]_ - Dimensions for table columns

  - **`hide_totals_row`** - _[boolean]_ - Whether to hide the totals row. Defaults to false.

  - **`hide_totals_col`** - _[boolean]_ - Whether to hide the totals column. Defaults to false.

  - **`totals_row_position`** - _[string]_ - Where to pin the totals row, either "top" (default) or "bottom".

  - **`row_limit`** - _[string]_ - Maximum number of rows to display in a pivot table (one of "5", "10", "25", "50", "100"). Omit or set to "all" for all rows.

  - **`comparison_range`** - _[string]_ - Comparison time range

  - **`filter`** - _[string]_ - Filter expression

  - **`title`** - _[string]_ - Table title

  - **`description`** - _[string]_ - Table description

### `markdown`

_[object]_ - (no description)

  - **`content`** - _[string]_ - Markdown content _(required)_

  - **`css`** - _[object]_ - CSS styles

  - **`title`** - _[string]_ - Markdown title

  - **`description`** - _[string]_ - Markdown description

### `image`

_[object]_ - (no description)

  - **`url`** - _[string]_ - Image URL. Used in light mode, and in dark mode when `dark_url` is not set. _(required)_

  - **`dark_url`** - _[string]_ - Image URL to use when the dashboard is displayed in dark mode. Defaults to `url`.

  - **`css`** - _[object]_ - CSS styles

  - **`title`** - _[string]_ - Image title

  - **`description`** - _[string]_ - Image description

### `map`

_[object]_ - (no description)

  - **`metrics_view`** - _[string]_ - Reference to the metrics view to use _(required)_

  - **`geo_dimension`** - _[object]_ - Dimension that holds the location of each row, as a GeoJSON geometry string or a DuckDB `POINT_2D` or `POLYGON_2D` value, in longitude and latitude. Only dimensions with `type: geo` are offered in the visual editor. _(required)_

    - **`field`** - _[string]_ - Name of the dimension _(required)_

    - **`type`** - _[string]_ - Field type, always `nominal`

  - **`color`** - _[object]_ - Measure that colors each point or region _(required)_

    - **`measure`** - _[string]_ - Name of the measure _(required)_

    - **`colorRange`** - _[object]_ - Color scale for the measure. Use `mode: scheme` with a `scheme` (`sequential`, `diverging`, `tealblues`, `viridis`, `magma`, `inferno`, `plasma`, `cividis`, `blues`, `teals`, `greens`, `greys`, `oranges`, `purples`, `reds`, `turbo`, `spectral`), or `mode: gradient` with `start` and `end` colors. Defaults to the `tealblues` scheme.

      - **`mode`** - _[string]_ - Whether to use a named color scheme or a two-color gradient

      - **`scheme`** - _[string]_ - Name of the color scheme, used when `mode` is `scheme`

      - **`start`** - _[string]_ - Color for the lowest value, used when `mode` is `gradient`. A hex color, `primary` or `secondary`.

      - **`end`** - _[string]_ - Color for the highest value, used when `mode` is `gradient`. A hex color, `primary` or `secondary`.

  - **`size_measure`** - _[object]_ - Measure that scales the radius of each point. Points only; ignored for polygons.

    - **`field`** - _[string]_ - Name of the measure _(required)_

    - **`type`** - _[string]_ - Field type, always `quantitative`

  - **`tooltip_dimension`** - _[object]_ - Dimension whose value is shown as the heading of the hover tooltip

    - **`field`** - _[string]_ - Name of the dimension _(required)_

    - **`type`** - _[string]_ - Field type, always `nominal`

  - **`initial_view`** - _[object]_ - Camera position the map opens with. When set, the map does not zoom to fit the data.

    - **`longitude`** - _[number]_ - Longitude of the map center _(required)_

    - **`latitude`** - _[number]_ - Latitude of the map center _(required)_

    - **`zoom`** - _[number]_ - Zoom level, from `0` (whole world) to about `22` (building level)

  - **`title`** - _[string]_ - Map title

  - **`description`** - _[string]_ - Map description

  - **`show_description_as_tooltip`** - _[boolean]_ - Show the description as a tooltip on the title instead of below it

  - **`time_filters`** - _[string]_ - Time range and comparison for this map, as `tr` and `compare_tr` URL parameters. Defaults to the canvas time range.

  - **`dimension_filters`** - _[string]_ - SQL filter expression applied to this map only, such as `country IN ('US')`
