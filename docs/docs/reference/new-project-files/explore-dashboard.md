---
note: GENERATED. DO NOT EDIT.
title: Explore Dashboard YAML
sidebar_position: 6
---

In your Rill project directory, create a explore dashboard, `<dashboard_name>.yaml`, file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `explore`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`refs`**  - _[array]_ - List of resource references, each as a string or map. 

     *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

     *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

    - **`name`**  - _[string]_ -   _(required)_

    - **`type`**  - _[string]_ -  

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

**`allow_custom_time_range`**  - _[boolean]_ - Defaults to true, when set to false it will hide the ability to set a custom time range for the user. 

**`defaults`**  - _[object]_ - defines the defaults YAML struct 

  - **`comparison_dimension`**  - _[string]_ - for dimension mode, specify the comparison dimension by name 

  - **`comparison_mode`**  - _[string]_ - Controls how to compare current data with historical or categorical baselines. Options: 'none' (no comparison), 'time' (compares with past based on default_time_range), 'dimension' (compares based on comparison_dimension values) 

  - **`dimensions`**  - _[one of]_ - Provides the default dimensions to load on viewing the dashboard 

     *option 1* - _[string]_ - Matches all fields

     *option 2* - _[array of string]_ - Explicit list of fields

     *option 3* - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

    - **`exclude`**  - _[object]_ - Select all dimensions except those listed here 

    - **`expr`**  - _[string]_ -  

    - **`regex`**  - _[string]_ - Select dimensions using a regular expression 

     *option 1* - 

     *option 2* - 

     *option 3* - 

  - **`measures`**  - _[one of]_ - Provides the default measures to load on viewing the dashboard 

     *option 1* - _[string]_ - Matches all fields

     *option 2* - _[array of string]_ - Explicit list of fields

     *option 3* - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

    - **`exclude`**  - _[object]_ - Select all dimensions except those listed here 

    - **`expr`**  - _[string]_ -  

    - **`regex`**  - _[string]_ - Select dimensions using a regular expression 

     *option 1* - 

     *option 2* - 

     *option 3* - 

  - **`time_range`**  - _[string]_ - Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, PT12H for 12 hours, P1M for 1 month, or P26W for 26 weeks) or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

**`description`**  - _[string]_ - Refers to the description of the explore dashboard 

**`display_name`**  - _[string]_ - Refers to the display name for the explore dashboard 

**`lock_time_zone`**  - _[boolean]_ - When true, the dashboard will be locked to the first time provided in the time_zones list. When no time_zones are provided, the dashboard will be locked to UTC 

**`measures`**  - _[one of]_ -  List of measure names. Use '*' to select all measures (default) 

   *option 1* - _[string]_ - Matches all fields

   *option 2* - _[array of string]_ - Explicit list of fields

   *option 3* - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

  - **`exclude`**  - _[object]_ - Select all dimensions except those listed here 

  - **`expr`**  - _[string]_ -  

  - **`regex`**  - _[string]_ - Select dimensions using a regular expression 

   *option 1* - 

   *option 2* - 

   *option 3* - 

**`security`**  - _[object]_ - Security rules to apply for access to the explore dashboard 

  - **`exclude`**  - _[array of object]_ - List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included 

      - **`if`**  - _[string]_ - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

      - **`names`**  - _[any of]_ - List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures  _(required)_

  - **`include`**  - _[array of object]_ - List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded 

      - **`if`**  - _[string]_ - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

      - **`names`**  - _[any of]_ - List of fields to include. Should match the name of one of the dashboard's dimensions or measures  _(required)_

  - **`row_filter`**  - _[string]_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause 

  - **`rules`**  - _[array of object]_ -  

      - **`type`**  - _[string]_ -   _(required)_

      - **`action`**  - _[string]_ -  

      - **`all`**  - _[boolean]_ -  

      - **`if`**  - _[string]_ -  

      - **`names`**  - _[array of string]_ -  

      - **`sql`**  - _[string]_ -  

  - **`access`**  - _[one of]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

     *option 1* - _[string]_ - 

     *option 2* - _[boolean]_ - 

**`time_zones`**  - _[array of string]_ - Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) 

**`banner`**  - _[string]_ - Refers to the custom banner displayed at the header of an explore dashboard 

**`dimensions`**  - _[one of]_ - List of dimension names. Use '*' to select all dimensions (default)  

   *option 1* - _[string]_ - Matches all fields

   *option 2* - _[array of string]_ - Explicit list of fields

   *option 3* - _[object]_ - Advanced matching using regex, DuckDB expression, or exclusion

  - **`regex`**  - _[string]_ - Select dimensions using a regular expression 

  - **`exclude`**  - _[object]_ - Select all dimensions except those listed here 

  - **`expr`**  - _[string]_ -  

   *option 1* - 

   *option 2* - 

   *option 3* - 

**`embeds`**  - _[object]_ -  

  - **`hide_pivot`**  - _[boolean]_ -  

**`metrics_view`**  - _[string]_ - Refers to the metrics view resource 

**`theme`**  - _[one of]_ - Name of the theme to use. Only one of theme and embedded_theme can be set. 

   *option 1* - _[string]_ - 

   *option 2* - _[object]_ - 

  - **`type`**  - _[string]_ - Refers to the resource type and must be `theme`  _(required)_

  - **`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

  - **`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

  - **`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

  - **`refs`**  - _[array]_ - List of resource references, each as a string or map. 

       *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

       *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

      - **`name`**  - _[string]_ -   _(required)_

      - **`type`**  - _[string]_ -  

  - **`colors`**  - _[object]_ -   _(required)_

**`time_ranges`**  - _[array]_ - Overrides the list of default time range selections available in the dropdown. It can be string or an object with a 'range' and optional 'comparison_offsets' 

     *option 1* - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

     *option 2* - _[object]_ - 

    - **`comparison_offsets`**  - _[array]_ - list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) 

         *option 1* - _[string]_ - Offset string only (range is inferred)

         *option 2* - _[object]_ - 

        - **`offset`**  - _[string]_ -  

        - **`range`**  - _[string]_ -  

    - **`range`**  - _[string]_ - a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection  _(required)_