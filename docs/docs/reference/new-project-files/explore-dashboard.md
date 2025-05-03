---
note: GENERATED. DO NOT EDIT.
title: Explore Dashboard YAML
sidebar_position: 6
---
## Explore Dashboard YAML

In your Rill project directory, create a explore dashboard, `<dashboard_name>.yaml`, file in the `dashboards` directory. Rill will ingest the dashboard definition next time you run `rill start`.

Type: `object`

## Properties:
#### All of the following:
- Part 1:
  ## type

  Type: `object`

  ## Properties:

  - **type** _(required)_:
    Type: `%!s(<nil>)`

- Part 2:
  ## common_properties

  Type: `object`

  ## Properties:

  - **namespace**:
    Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`.

    Type: `string`


  - **refs**:
    List of resource references, each as a string or map.

    Type: `array`

    #### Array Items:
      Type: `%!s(<nil>)`

      #### One of the following:
      - Option 1:
        A string reference like 'resource-name' or 'Kind/resource-name'.

        Type: `string`

      - Option 2:
        An object reference with at least a 'name' and 'type'.

        Type: `object`

        ## Properties:

        - **name** _(required)_:
          Type: `string`


        - **type**:
          Type: `string`


  - **version**:
    Version of the parser to use for this file. Enables backwards compatibility for breaking changes.

    Type: `integer`


  - **name**:
    Name is usually inferred from the filename, but can be specified manually.

    Type: `string`

- Part 3:
  ## explore_properties

  Type: `object`

  ## Properties:

  - **description**:
    Refers to the description of the explore dashboard

    Type: `string`


  - **allow_custom_time_range**:
    Defaults to true, when set to false it will hide the ability to set a custom time range for the user.

    Type: `boolean`


  - **dimensions**:
    can be:
1. The string '*'
2. An array of strings
3. An object with one of: 'regex', 'expr', or 'exclude'

    Type: `%!s(<nil>)`

    #### One of the following:
    - Option 1:
      Matches all fields

      Type: `string`

    - Option 2:
      Explicit list of fields

      Type: `array`

      #### Array Items:
        Type: `string`

    - Option 3:
      Advanced matching using regex, DuckDB expression, or exclusion

      Type: `object`

      ## Properties:

      - **exclude**:
        Select all dimensions except those listed here

        Type: `object`

        ## Properties:

      - **expr**:
        Type: `string`


      - **regex**:
        Select dimensions using a regular expression

        Type: `string`

      #### One of the following:
      - Option 1:
        Type: `%!s(<nil>)`

      - Option 2:
        Type: `%!s(<nil>)`

      - Option 3:
        Type: `%!s(<nil>)`


  - **theme**:
    Name of the theme to use. Only one of theme and embedded_theme can be set.

    Type: `%!s(<nil>)`

    #### One of the following:
    - Option 1:
      Type: `string`

    - Option 2:
      ## Theme YAML

      Type: `object`

      ## Properties:
      #### All of the following:
      - Part 1:
        ## type

        Type: `object`

        ## Properties:

        - **type** _(required)_:
          Type: `%!s(<nil>)`

      - Part 2:
        ## common_properties

        Type: `object`

        ## Properties:

        - **namespace**:
          Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`.

          Type: `string`


        - **refs**:
          List of resource references, each as a string or map.

          Type: `array`

          #### Array Items:
            Type: `%!s(<nil>)`

            #### One of the following:
            - Option 1:
              A string reference like 'resource-name' or 'Kind/resource-name'.

              Type: `string`

            - Option 2:
              An object reference with at least a 'name' and 'type'.

              Type: `object`

              ## Properties:

              - **name** _(required)_:
                Type: `string`


              - **type**:
                Type: `string`


        - **version**:
          Version of the parser to use for this file. Enables backwards compatibility for breaking changes.

          Type: `integer`


        - **name**:
          Name is usually inferred from the filename, but can be specified manually.

          Type: `string`

      - Part 3:
        ## theme_properties

        Type: `object`

        ## Properties:

        - **colors** _(required)_:
          Type: `object`

          ## Properties:
          #### Any of the following:
          - Option 1:
            Type: `%!s(<nil>)`

          - Option 2:
            Type: `%!s(<nil>)`

      - Part 4:
        ## environment_overrides

        Type: `%!s(<nil>)`


  - **time_ranges**:
    Overrides the list of default time range selections available in the dropdown. It can be string or an object with a 'range' and optional 'comparison_offsets'

    Type: `array`

    #### Array Items:
      Type: `%!s(<nil>)`

      #### One of the following:
      - Option 1:
        a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

        Type: `string`

      - Option 2:
        Type: `object`

        ## Properties:

        - **comparison_offsets**:
          list of time comparison options for this time range selection (optional). Must be one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions)

          Type: `array`

          #### Array Items:
            Type: `%!s(<nil>)`

            #### One of the following:
            - Option 1:
              Offset string only (range is inferred)

              Type: `string`

            - Option 2:
              Type: `object`

              ## Properties:

              - **range**:
                Type: `string`


              - **offset**:
                Type: `string`


        - **range** _(required)_:
          a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

          Type: `string`


  - **time_zones**:
    Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)

    Type: `array`

    #### Array Items:
      Type: `string`


  - **type**:
    Type: `%!s(<nil>)`


  - **embeds**:
    Type: `object`

    ## Properties:

    - **hide_pivot**:
      Type: `boolean`


  - **measures**:
    can be:
1. The string '*'
2. An array of strings
3. An object with one of: 'regex', 'expr', or 'exclude'

    Type: `%!s(<nil>)`

    #### One of the following:
    - Option 1:
      Matches all fields

      Type: `string`

    - Option 2:
      Explicit list of fields

      Type: `array`

      #### Array Items:
        Type: `string`

    - Option 3:
      Advanced matching using regex, DuckDB expression, or exclusion

      Type: `object`

      ## Properties:

      - **exclude**:
        Select all dimensions except those listed here

        Type: `object`

        ## Properties:

      - **expr**:
        Type: `string`


      - **regex**:
        Select dimensions using a regular expression

        Type: `string`

      #### One of the following:
      - Option 1:
        Type: `%!s(<nil>)`

      - Option 2:
        Type: `%!s(<nil>)`

      - Option 3:
        Type: `%!s(<nil>)`


  - **banner**:
    Refers to the custom banner displayed at the header of an explore dashboard

    Type: `string`


  - **display_name**:
    Refers to the display name for the explore dashboard

    Type: `string`


  - **lock_time_zone**:
    When true, the dashboard will be locked to the first time provided in the time_zones list. When no time_zones are provided, the dashboard will be locked to UTC

    Type: `boolean`


  - **metrics_view**:
    Refers to the metrics view resource

    Type: `string`


  - **security**:
    Security rules to apply for access to the explore dashboard

    Type: `object`

    ## Properties:

    - **access**:
      Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean.

      Type: `%!s(<nil>)`

      #### One of the following:
      - Option 1:
        Type: `string`

      - Option 2:
        Type: `boolean`


    - **exclude**:
      List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included

      Type: `array`

      #### Array Items:
        Type: `object`

        ## Properties:

        - **if** _(required)_:
          Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean

          Type: `string`


        - **names** _(required)_:
          List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures

          Type: `%!s(<nil>)`

          #### Any of the following:
          - Option 1:
            Type: `array`

            #### Array Items:
              Type: `string`

          - Option 2:
            Type: `string`

            Enum: `[*]`


    - **include**:
      List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded

      Type: `array`

      #### Array Items:
        Type: `object`

        ## Properties:

        - **if** _(required)_:
          Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean

          Type: `string`


        - **names** _(required)_:
          List of fields to include. Should match the name of one of the dashboard's dimensions or measures

          Type: `%!s(<nil>)`

          #### Any of the following:
          - Option 1:
            Type: `array`

            #### Array Items:
              Type: `string`

          - Option 2:
            Type: `string`

            Enum: `[*]`


    - **row_filter**:
      SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause

      Type: `string`


    - **rules**:
      Type: `array`

      #### Array Items:
        Type: `object`

        ## Properties:

        - **if**:
          Type: `string`


        - **names**:
          Type: `array`

          #### Array Items:
            Type: `string`


        - **sql**:
          Type: `string`


        - **type** _(required)_:
          Type: `string`

          Enum: `[access field_access row_filter]`


        - **action**:
          Type: `string`

          Enum: `[allow deny]`


        - **all**:
          Type: `boolean`


  - **defaults**:
    defines the defaults YAML struct

    Type: `object`

    ## Properties:

    - **time_range**:
      Refers to the default time range shown when a user initially loads the dashboard. The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, PT12H for 12 hours, P1M for 1 month, or P26W for 26 weeks) or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions)

      Type: `string`


    - **comparison_dimension**:
      for dimension mode, specify the comparison dimension by name

      Type: `string`


    - **comparison_mode**:
      Controls how to compare current data with historical or categorical baselines. Options: 'none' (no comparison), 'time' (compares with past based on default_time_range), 'dimension' (compares based on comparison_dimension values)

      Type: `string`


    - **dimensions**:
      Provides the default dimensions to load on viewing the dashboard

      Type: `%!s(<nil>)`

      #### One of the following:
      - Option 1:
        Matches all fields

        Type: `string`

      - Option 2:
        Explicit list of fields

        Type: `array`

        #### Array Items:
          Type: `string`

      - Option 3:
        Advanced matching using regex, DuckDB expression, or exclusion

        Type: `object`

        ## Properties:

        - **exclude**:
          Select all dimensions except those listed here

          Type: `object`

          ## Properties:

        - **expr**:
          Type: `string`


        - **regex**:
          Select dimensions using a regular expression

          Type: `string`

        #### One of the following:
        - Option 1:
          Type: `%!s(<nil>)`

        - Option 2:
          Type: `%!s(<nil>)`

        - Option 3:
          Type: `%!s(<nil>)`


    - **measures**:
      Provides the default measures to load on viewing the dashboard

      Type: `%!s(<nil>)`

      #### One of the following:
      - Option 1:
        Matches all fields

        Type: `string`

      - Option 2:
        Explicit list of fields

        Type: `array`

        #### Array Items:
          Type: `string`

      - Option 3:
        Advanced matching using regex, DuckDB expression, or exclusion

        Type: `object`

        ## Properties:

        - **regex**:
          Select dimensions using a regular expression

          Type: `string`


        - **exclude**:
          Select all dimensions except those listed here

          Type: `object`

          ## Properties:

        - **expr**:
          Type: `string`

        #### One of the following:
        - Option 1:
          Type: `%!s(<nil>)`

        - Option 2:
          Type: `%!s(<nil>)`

        - Option 3:
          Type: `%!s(<nil>)`

- Part 4:
  ## environment_overrides

  Type: `%!s(<nil>)`

