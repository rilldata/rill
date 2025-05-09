---
note: GENERATED. DO NOT EDIT.
title: Canvas Dashboard YAML
sidebar_position: 33
---

In your Rill project directory, create a explore dashboard, `<dashboard_name>.yaml`, file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `canvas`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`refs`**  - _[array of oneOf]_ - List of resource references, each as a string or map. 

  *option 1* - _[object]_ - An object reference with at least a `<name>` and `<type>`.

  - **`type`**  - _[string]_ - type of resource 

  - **`name`**  - _[string]_ - name of resource  _(required)_

  *option 2* - _[string]_ - A string reference like `<resource-name>` or `<type/resource-name>`.

**`dev`**  - _[object]_ - Overrides properties in development 

**`prod`**  - _[object]_ - Overrides properties in production 

**`display_name`**  - _[string]_ - Refers to the display name for the canvas 

**`banner`**  - _[string]_ - Refers to the custom banner displayed at the header of an Canvas dashboard 

**`max_width`**  - _[integer]_ - Max width in pixels of the canvas 

**`gap_x`**  - _[integer]_ - Horizontal gap in pixels of the canvas 

**`gap_y`**  - _[integer]_ - Vertical gap in pixels of the canvas 

**`theme`**  - _[anyOf]_ - Name of the theme to use. Only one of theme and embedded_theme can be set. 

  *option 1* - _[string]_ 

  *option 2* - _[object]_ 

**`allow_custom_time_range`**  - _[boolean]_ - Defaults to true, when set to false it will hide the ability to set a custom time range for the user.  

**`time_ranges`**  - _[array of oneOf]_ - Overrides the list of default time range selections available in the dropdown. It can be string or an object with a 'range' and optional 'comparison_offsets' 

  *option 1* - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

  *option 2* - _[object]_ 

  - **`range`**  - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection  _(required)_

  - **`comparison_offsets`**  - _[array of oneOf]_ - list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

    *option 1* - _[string]_ - Offset string only (range is inferred)

    *option 2* - _[object]_ 

    - **`offset`**  - _[string]_  

    - **`range`**  - _[string]_  

**`time_zones`**  - _[array of string]_ - Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) 

**`filters`**  - _[object]_ - Indicates if filters should be enabled for the canvas. 

  - **`enable`**  - _[boolean]_  

**`defaults`**  - _[object]_ - Preset UI state to show by default 

  - **`time_range`**  - _[string]_  

  - **`comparison_mode`**  - _[string]_  

  - **`comparison_dimension`**  - _[string]_  

**`variables`**  - _[array of object]_ - Variables that can be used in the canvas 

  - **`name`**  - _[string]_   _(required)_

  - **`type`**  - _[string]_   _(required)_

  - **`value`**  - The value can be of any type. 

**`rows`**  - _[array of object]_ - Refers to all of the rows displayed on the Canvas  _(required)_

  - **`height`**  - _[string]_  

  - **`items`**  - _[array of object]_  

    - **`width`**   

    - **`component`**  - _[string]_  

**`security`**  - _[object]_ - Security rules to apply for access to the canvas 

  - **`access`**  - _[oneOf]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

    *option 1* - _[string]_ 

    *option 2* - _[boolean]_ 

  - **`row_filter`**  - _[string]_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause 

  - **`include`**  - _[array of object]_ - List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded 

    - **`if`**  - _[string]_ - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

    - **`names`**  - _[anyOf]_ - List of fields to include. Should match the name of one of the dashboard's dimensions or measures  _(required)_

      *option 1* - _[array of string]_ 

      *option 2* - _[string]_ 

  - **`exclude`**  - _[array of object]_ - List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included 

    - **`if`**  - _[string]_ - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

    - **`names`**  - _[anyOf]_ - List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures  _(required)_

      *option 1* - _[array of string]_ 

      *option 2* - _[string]_ 

  - **`rules`**  - _[array of object]_  

    - **`type`**  - _[string]_   _(required)_

    - **`action`**  - _[string]_  

    - **`if`**  - _[string]_  

    - **`names`**  - _[array of string]_  

    - **`all`**  - _[boolean]_  

    - **`sql`**  - _[string]_  