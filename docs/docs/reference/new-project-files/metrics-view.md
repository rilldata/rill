---
note: GENERATED. DO NOT EDIT.
title: Metrics View YAML
sidebar_position: 7
---
## Metrics View YAML

In your Rill project directory, create a metrics view, `<metrics_view>.yaml`, file in the `metrics` directory. Rill will ingest the metric view definition next time you run `rill start`.

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

  - **name**:
    Name is usually inferred from the filename, but can be specified manually.

    Type: `string`


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

- Part 3:
  ## metrics_view_properties

  Type: `object`

  ## Properties:

  - **first_day_of_week**:
    Refers to the first day of the week for time grain aggregation (for example, Sunday instead of Monday). The valid values are 1 through 7 where Monday=1 and Sunday=7

    Type: `integer`


  - **first_month_of_year**:
    Refers to the first month of the year for time grain aggregation. The valid values are 1 through 12 where January=1 and December=12

    Type: `integer`


  - **measures**:
    Used to define the numeric aggregates of columns from your data model

    Type: `array`

    #### Array Items:
      Type: `object`

      ## Properties:

      - **description**:
        a freeform text description of the dimension

        Type: `string`


      - **display_name**:
        the display name of your measure.

        Type: `string`


      - **format_d3**:
        Controls the formatting of this measure using a [d3-format](https://d3js.org/d3-format) string. If an invalid format string is supplied, the measure will fall back to `format_preset: humanize`. A measure cannot have both `format_preset` and `format_d3`. If neither is provided, the humanize preset is used by default. Example: `format_d3: ".2f"` formats using fixed-point notation with two decimal places. Example: `format_d3: ",.2r"` formats using grouped thousands with two significant digits. (optional)

        Type: `string`


      - **format_d3_locale**:
        locale configuration passed through to D3, enabling changing the currency symbol among other things. For details, see the docs for D3's [formatLocale](https://d3js.org/d3-format#formatLocale)

        Type: `object`

        ## Properties:

      - **format_preset**:
        Controls the formatting of this measure using a predefined preset. Measures cannot have both `format_preset` and `format_d3`. If neither is supplied, the measure will be formatted using the `humanize` preset by default. Available options:
- `humanize`: Round numbers into thousands (K), millions (M), billions (B), etc.
- `none`: Raw output.
- `currency_usd`: Round to 2 decimal points with a dollar sign ($).
- `currency_eur`: Round to 2 decimal points with a euro sign (€).
- `percentage`: Convert a rate into a percentage with a % sign.
- `interval_ms`: Convert milliseconds into human-readable durations like hours (h), days (d), years (y), etc. (optional)

        Type: `string`


      - **requires**:
        using an available measure or dimension in your metrics view to set a required parameter, cannot be used with simple measures

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `string`

        - Option 2:
          Type: `array`

          #### Array Items:
            Type: `%!s(<nil>)`

            #### Any of the following:
            - Option 1:
              Shorthand field selector, interpreted as the name.

              Type: `string`

            - Option 2:
              Type: `object`

              ## Properties:

              - **name** _(required)_:
                Type: `string`


              - **time_grain**:
                Time grain for time-based dimensions.

                Type: `string`

                Enum: `[ ms millisecond s second min minute h hour d day w week month q quarter y year]`


      - **treat_nulls_as**:
        used to configure what value to fill in for missing time buckets. This also works generally as COALESCING over non empty time buckets.

        Type: `string`


      - **type**:
        Type: `string`


      - **expression**:
        a combination of operators and functions for aggregations

        Type: `string`


      - **name**:
        a stable identifier for the measure

        Type: `string`


      - **per**:
        for per dimensions

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Type: `string`

        - Option 2:
          Type: `array`

          #### Array Items:
            Type: `%!s(<nil>)`

            #### Any of the following:
            - Option 1:
              Shorthand field selector, interpreted as the name.

              Type: `string`

            - Option 2:
              Type: `object`

              ## Properties:

              - **name** _(required)_:
                Type: `string`


              - **time_grain**:
                Time grain for time-based dimensions.

                Type: `string`

                Enum: `[ ms millisecond s second min minute h hour d day w week month q quarter y year]`


      - **valid_percent_of_total**:
        a boolean indicating whether percent-of-total values should be rendered for this measure 

        Type: `boolean`


      - **window**:
        A measure window can be defined as a keyword string (e.g., 'time' or 'all') or an object with detailed window configuration.

        Type: `%!s(<nil>)`

        #### Any of the following:
        - Option 1:
          Shorthand: 'time' or 'true' means time-partitioned, 'all' means non-partitioned.

          Type: `string`

          Enum: `[time true all]`

        - Option 2:
          Type: `object`

          ## Properties:

          - **order**:
            to order the window

            Type: `%!s(<nil>)`

            #### Any of the following:
            - Option 1:
              Type: `string`

            - Option 2:
              Type: `array`

              #### Array Items:
                Type: `%!s(<nil>)`

                #### Any of the following:
                - Option 1:
                  Shorthand field selector, interpreted as the name.

                  Type: `string`

                - Option 2:
                  Type: `object`

                  ## Properties:

                  - **name** _(required)_:
                    Type: `string`


                  - **time_grain**:
                    Time grain for time-based dimensions.

                    Type: `string`

                    Enum: `[ ms millisecond s second min minute h hour d day w week month q quarter y year]`


          - **partition**:
            Type: `boolean`


          - **frame**:
            sets the frame of your window

            Type: `string`


  - **model**:
    Refers to the model powering the dashboard (either model or table is required)

    Type: `string`


  - **security**:
    Defines a security policy for the dashboard

    Type: `object`

    ## Properties:

    - **rules**:
      Type: `array`

      #### Array Items:
        Type: `object`

        ## Properties:

        - **type** _(required)_:
          Type: `string`

          Enum: `[access field_access row_filter]`


        - **action**:
          Type: `string`

          Enum: `[allow deny]`


        - **all**:
          Type: `boolean`


        - **if**:
          Type: `string`


        - **names**:
          Type: `array`

          #### Array Items:
            Type: `string`


        - **sql**:
          Type: `string`


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


  - **database_schema**:
    Refers to the schema to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified

    Type: `string`


  - **dimensions**:
    Relates to exploring segments or dimensions of your data and filtering the dashboard

    Type: `array`

    #### Array Items:
      Type: `object`

      ## Properties:

      - **description**:
        a freeform text description of the dimension

        Type: `string`


      - **display_name**:
        a display name for your dimension

        Type: `string`


      - **expression**:
        a non-aggregate expression such as string_split(domain, '.'). One of column and expression is required but cannot have both at the same time

        Type: `string`


      - **name**:
        a stable identifier for the dimension

        Type: `string`


      - **unnest**:
        if true, allows multi-valued dimension to be unnested (such as lists) and filters will automatically switch to "contains" instead of exact match 

        Type: `boolean`


      - **uri**:
        enable if your dimension is a clickable URL to enable single click navigation (boolean or valid SQL expression) 

        Type: `[string boolean]`


      - **column**:
        a categorical column

        Type: `string`

      #### Any of the following:
      - Option 1:
        Type: `%!s(<nil>)`

      - Option 2:
        Type: `%!s(<nil>)`


  - **smallest_time_grain**:
    Refers to the smallest time granularity the user is allowed to view. The valid values are: millisecond, second, minute, hour, day, week, month, quarter, year

    Type: `string`


  - **table**:
    Refers to the table powering the dashboard, should be used instead of model for dashboards create from external OLAP tables (either table or model is required)

    Type: `string`


  - **timeseries**:
    Refers to the timestamp column from your model that will underlie x-axis data in the line charts. If not specified, the line charts will not appear

    Type: `string`


  - **watermark**:
    A SQL expression that tells us the max timestamp that the metrics are considered valid for. Usually does not need to be overwritten

    Type: `string`


  - **database**:
    Refers to the database to use in the OLAP engine (to be used in conjunction with table). Otherwise, will use the default database or schema if not specified

    Type: `string`


  - **description**:
    Refers to the description for the metrics view

    Type: `string`


  - **display_name**:
    Refers to the display name for the metrics view

    Type: `string`

- Part 4:
  ## environment_overrides

  Type: `%!s(<nil>)`

