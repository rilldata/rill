---
note: GENERATED. DO NOT EDIT.
title: API YAML
sidebar_position: 2
---
## API YAML

In your Rill project directory, create a new file name `<api-name>.yaml` in the `apis` directory containing a custom API definition. See comprehensive documentation on how to define and use [custom APIs](/integrate/custom-apis/index.md)

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


  - **namespace**:
    Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`.

    Type: `string`

- Part 3:
  ## api_properties

  Type: `object`

  ## Properties:

  - **openapi**:
    Type: `object`

    ## Properties:

    - **request**:
      Type: `object`

      ## Properties:

      - **parameters**:
        Type: `array`

        #### Array Items:
          Type: `object`

          ## Properties:

    - **response**:
      Type: `object`

      ## Properties:

      - **schema**:
        Type: `object`

        ## Properties:

    - **summary**:
      Type: `string`


  - **security**:
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


        - **type** _(required)_:
          Type: `string`

          Enum: `[access field_access row_filter]`


  - **skip_nested_security**:
    Type: `boolean`

  #### One of the following:
  - Option 1:
    Type: `object`

    ## Properties:
    #### One of the following:
    - Option 1:
      ## sql

      Type: `%!s(<nil>)`

    - Option 2:
      ## metrics_sql

      Type: `%!s(<nil>)`

    - Option 3:
      ## api

      Type: `%!s(<nil>)`

    - Option 4:
      ## glob

      Type: `%!s(<nil>)`

    - Option 5:
      ## resource_status

      Type: `%!s(<nil>)`

- Part 4:
  ## environment_overrides

  Type: `%!s(<nil>)`

