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

### `params`

_[array of object]_ - List of typed parameters that canvases can bind values to when referencing this component. Bound values are available in the renderer properties' templating as `{{ .params.<name> }}`.

  - **`name`** - _[string]_ - Param name. Must be a valid identifier; referenced in the renderer properties' templating as `{{ .params.<name> }}`. _(required)_

  - **`type`** - _[string]_ - Param type. Params of type `metrics_view` must be named `metrics_view` or end with `_metrics_view`. _(required)_

  - **`description`** - _[string]_ - Human-facing description of the param

  - **`required`** - _[boolean]_ - If true, a canvas item referencing this component must bind a value for the param. Mutually exclusive with `default`.

  - **`default`** - _[string, number, boolean]_ - Default value used when the param is not bound

  - **`metrics_view`** - _[string]_ - For `measure`, `dimension` and `time_dimension` params, the name of a sibling param of type `metrics_view` whose bound metrics view the field must belong to. May be omitted when exactly one `metrics_view` param is declared.

  - **`options`** - _[array]_ - For scalar params, the allowed values. Renders as a select input in visual editors.

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

  - **`url`** - _[string]_ - Image URL _(required)_

  - **`css`** - _[object]_ - CSS styles

  - **`title`** - _[string]_ - Image title

  - **`description`** - _[string]_ - Image description
