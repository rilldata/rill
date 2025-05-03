---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 10
---
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

