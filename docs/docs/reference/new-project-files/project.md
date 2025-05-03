---
note: GENERATED. DO NOT EDIT.
title: Project YAML
sidebar_position: 11
---
## Project YAML

The `rill.yaml` file contains metadata about your project.

Type: `object`

## Properties:
#### All of the following:
- Part 1:
  ## rillyaml_properties

  Type: `object`

  ## Properties:

  - **explores**:
    Defines project-wide default settings for explores. Unless overridden, individual explores will inherit these defaults

    Type: `%!s(<nil>)`


  - **ignore_paths**:
    A list of file or directory paths to exclude from parsing. Useful for ignoring extraneous or non-Rill files in the project

    Type: `array`

    #### Array Items:
      Type: `string`


  - **mock_users**:
    A list of mock users used to test dashboard security policies within the project

    Type: `array`

    #### Array Items:
      Type: `object`

      ## Properties:

      - **admin**:
        Indicates whether the mock user has administrative privileges

        Type: `boolean`


      - **email** _(required)_:
        The email address of the mock user. This field is required

        Type: `string`


      - **groups**:
        An array of group names that the mock user is a member of

        Type: `array`

        #### Array Items:
          Type: `string`


      - **name**:
        The name of the mock user.

        Type: `string`


  - **olap_connector**:
    Specifies the default OLAP engine for the project. Defaults to duckdb if not set

    Type: `string`


  - **public_paths**:
    List of file or directory paths to expose over HTTP. Defaults to ['./public']

    Type: `array`

    #### Array Items:
      Type: `string`


  - **compiler**:
    Specifies the parser version to use for compiling resources

    Type: `string`


  - **features**:
    Enables or disables experimental or optional features using key-value pairs, where the key is the feature name and the value is a boolean

    Type: `%!s(<nil>)`

    #### One of the following:
    - Option 1:
      Type: `object`

      ## Properties:
    - Option 2:
      Type: `array`

      #### Array Items:
        Type: `string`


  - **metrics_views**:
    Defines project-wide default settings for metrics_views. Unless overridden, individual metrics_views will inherit these defaults

    Type: `%!s(<nil>)`


  - **models**:
    Defines project-wide default settings for models. Unless overridden, individual models will inherit these defaults

    Type: `%!s(<nil>)`


  - **description**:
    A brief description of the project

    Type: `string`


  - **display_name**:
    The display name of the project, shown in the upper-left corner of the UI

    Type: `string`

- Part 2:
  ## environment_overrides

  Type: `%!s(<nil>)`

