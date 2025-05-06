---
note: GENERATED. DO NOT EDIT.
title: Project YAML
sidebar_position: 11
---

The `rill.yaml` file contains metadata about your project.

## Properties


**`compiler`**  - _[string]_ - Specifies the parser version to use for compiling resources 

**`display_name`**  - _[string]_ - The display name of the project, shown in the upper-left corner of the UI 

**`description`**  - _[string]_ - A brief description of the project 

**`olap_connector`**  - _[string]_ - Specifies the default OLAP engine for the project. Defaults to duckdb if not set 

**`models`**  - Defines project-wide default settings for models. Unless overridden, individual models will inherit these defaults 

**`metrics_views`**  - Defines project-wide default settings for metrics_views. Unless overridden, individual metrics_views will inherit these defaults 

**`explores`**  - Defines project-wide default settings for explores. Unless overridden, individual explores will inherit these defaults 

**`features`**  - _[oneOf]_ - Optional feature flags. Can be specified as a map of feature names to booleans, or as a list of enabled feature names. 

  *option 1* - _[object]_ - Map of feature names to booleans.

  *option 2* - _[array of string]_ - List of enabled feature names.

**`public_paths`**  - _[array of string]_ - List of file or directory paths to expose over HTTP. Defaults to ['./public'] 

**`ignore_paths`**  - _[array of string]_ - A list of file or directory paths to exclude from parsing. Useful for ignoring extraneous or non-Rill files in the project 

**`mock_users`**  - _[array of object]_ - A list of mock users used to test dashboard security policies within the project 

  - **`email`**  - _[string]_ - The email address of the mock user. This field is required  _(required)_

  - **`name`**  - _[string]_ - The name of the mock user. 

  - **`admin`**  - _[boolean]_ - Indicates whether the mock user has administrative privileges 

  - **`groups`**  - _[array of string]_ - An array of group names that the mock user is a member of 

**`dev`**  - _[object]_ - Overrides properties in development 

**`prod`**  - _[object]_ - Overrides properties in production 