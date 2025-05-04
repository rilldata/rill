---
note: GENERATED. DO NOT EDIT.
title: Canvas Dashboard YAML
sidebar_position: 3
---

In your Rill project directory, create a explore dashboard, `<dashboard_name>.yaml`, file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `canvas`  _(required)_

**`refs`**  - _[array]_ - List of resource references, each as a string or map. 

     *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

     *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

    - **`type`**  - _[string]_ -  

    - **`name`**  - _[string]_ -   _(required)_

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`theme`**  - _[any of]_ - Name of the theme to use. Only one of theme and embedded_theme can be set. 

**`time_ranges`**  - _[array]_ - Overrides the list of default time range selections available in the dropdown. It can be string or an object with a 'range' and optional 'comparison_offsets' 

     *option 1* - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

     *option 2* - _[object]_ - 

    - **`comparison_offsets`**  - _[array]_ - list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

         *option 1* - _[string]_ - Offset string only (range is inferred)

         *option 2* - _[object]_ - 

        - **`offset`**  - _[string]_ -  

        - **`range`**  - _[string]_ -  

    - **`range`**  - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection  _(required)_

**`time_zones`**  - _[array of string]_ - Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) 

**`filters`**  - _[object]_ - Indicates if filters should be enabled for the canvas. 

  - **`enable`**  - _[boolean]_ -  

**`gap_x`**  - _[integer]_ - Horizontal gap in pixels of the canvas 

**`max_width`**  - _[integer]_ - Max width in pixels of the canvas 

**`rows`**  - _[array of object]_ - Refers to all of the rows displayed on the Canvas  _(required)_

    - **`items`**  - _[array of object]_ -  

        - **`component`**  - _[string]_ -  

        - **`width`**  -  

    - **`height`**  - _[string]_ -  

**`security`**  - _[object]_ - Security rules to apply for access to the canvas 

  - **`access`**  - _[one of]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

     *option 1* - _[string]_ - 

     *option 2* - _[boolean]_ - 

  - **`exclude`**  - _[array of object]_ - List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included 

      - **`names`**  - _[any of]_ - List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures  _(required)_

      - **`if`**  - _[string]_ - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

  - **`include`**  - _[array of object]_ - List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded 

      - **`if`**  - _[string]_ - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

      - **`names`**  - _[any of]_ - List of fields to include. Should match the name of one of the dashboard's dimensions or measures  _(required)_

  - **`row_filter`**  - _[string]_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause 

  - **`rules`**  - _[array of object]_ -  

      - **`sql`**  - _[string]_ -  

      - **`type`**  - _[string]_ -   _(required)_

      - **`action`**  - _[string]_ -  

      - **`all`**  - _[boolean]_ -  

      - **`if`**  - _[string]_ -  

      - **`names`**  - _[array of string]_ -  

**`variables`**  - _[array of object]_ - Variables that can be used in the canvas 

    - **`name`**  - _[string]_ -   _(required)_

    - **`type`**  - _[string]_ -   _(required)_

    - **`value`**  - The value can be of any type. 

**`allow_custom_time_range`**  - _[boolean]_ - Defaults to true, when set to false it will hide the ability to set a custom time range for the user.  

**`banner`**  - _[string]_ - Refers to the custom banner displayed at the header of an Canvas dashboard 

**`defaults`**  - _[object]_ - Preset UI state to show by default 

  - **`comparison_dimension`**  - _[string]_ -  

  - **`comparison_mode`**  - _[string]_ -  

  - **`time_range`**  - _[string]_ -  

**`display_name`**  - _[string]_ - Refers to the display name for the canvas 

**`gap_y`**  - _[integer]_ - Vertical gap in pixels of the canvas 