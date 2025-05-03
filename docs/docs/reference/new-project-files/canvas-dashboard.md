---
note: GENERATED. DO NOT EDIT.
title: Canvas Dashboard YAML
sidebar_position: 3
---
## Canvas Dashboard YAML

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
  ## canvas_properties

  Type: `object`

  ## Properties:

  - **allow_custom_time_range**:
    Defaults to true, when set to false it will hide the ability to set a custom time range for the user. 

    Type: `boolean`


  - **banner**:
    Refers to the custom banner displayed at the header of an Canvas dashboard

    Type: `string`


  - **defaults**:
    Preset UI state to show by default

    Type: `object`

    ## Properties:

    - **comparison_dimension**:
      Type: `string`


    - **comparison_mode**:
      Type: `string`


    - **time_range**:
      Type: `string`


  - **display_name**:
    Refers to the display name for the canvas

    Type: `string`


  - **filters**:
    Indicates if filters should be enabled for the canvas.

    Type: `object`

    ## Properties:

    - **enable**:
      Type: `boolean`


  - **gap_x**:
    Horizontal gap in pixels of the canvas

    Type: `integer`


  - **gap_y**:
    Vertical gap in pixels of the canvas

    Type: `integer`


  - **max_width**:
    Max width in pixels of the canvas

    Type: `integer`


  - **rows** _(required)_:
    Refers to all of the rows displayed on the Canvas

    Type: `array`

    #### Array Items:
      Refers to each row of components, multiple items can be listed in a single items

      Type: `object`

      ## Properties:

      - **height**:
        Type: `string`


      - **items**:
        Type: `array`

        #### Array Items:
          Type: `object`

          ## Properties:

          - **component**:
            Type: `string`


          - **width**:
            Type: `[string integer]`


  - **security**:
    Security rules to apply for access to the canvas

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


        - **if** _(required)_:
          Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean

          Type: `string`


    - **row_filter**:
      SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause

      Type: `string`


    - **rules**:
      Type: `array`

      #### Array Items:
        Type: `object`

        ## Properties:

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


        - **type** _(required)_:
          Type: `string`

          Enum: `[access field_access row_filter]`


        - **action**:
          Type: `string`

          Enum: `[allow deny]`


  - **theme**:
    Name of the theme to use. Only one of theme and embedded_theme can be set.

    Type: `%!s(<nil>)`

    #### Any of the following:
    - Option 1:
      Type: `string`

    - Option 2:
      Type: `object`

      ## Properties:

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

              - **offset**:
                Type: `string`


              - **range**:
                Type: `string`


        - **range** _(required)_:
          a valid [ISO 8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) duration or one of the [Rill ISO 8601 extensions](https://docs.rilldata.com/reference/rill-iso-extensions#extensions) extensions for the selection

          Type: `string`


  - **time_zones**:
    Refers to the time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)

    Type: `array`

    #### Array Items:
      Type: `string`


  - **variables**:
    Variables that can be used in the canvas

    Type: `array`

    #### Array Items:
      Type: `object`

      ## Properties:

      - **name** _(required)_:
        Type: `string`


      - **type** _(required)_:
        Type: `string`


      - **value**:
        The value can be of any type.

        Type: `[string number boolean object array null]`

- Part 4:
  ## environment_overrides

  Type: `%!s(<nil>)`

