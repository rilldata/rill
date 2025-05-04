---
note: GENERATED. DO NOT EDIT.
title: API YAML
sidebar_position: 2
---

In your Rill project directory, create a new file name `<api-name>.yaml` in the `apis` directory containing a custom API definition. See comprehensive documentation on how to define and use [custom APIs](/integrate/custom-apis/index.md)

## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `api`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`refs`**  - _[array]_ - List of resource references, each as a string or map. 

     *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

     *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

    - **`name`**  - _[string]_ -   _(required)_

    - **`type`**  - _[string]_ -  

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

**`openapi`**  - _[object]_ -  

  - **`request`**  - _[object]_ -  

    - **`parameters`**  - _[array of object]_ -  

  - **`response`**  - _[object]_ -  

    - **`schema`**  - _[object]_ -  

  - **`summary`**  - _[string]_ -  

**`security`**  - _[object]_ -  

  - **`row_filter`**  - _[string]_ - SQL expression to filter the underlying model by. Can leverage templated user attributes to customize the filter for the requesting user. Needs to be a valid SQL expression that can be injected into a WHERE clause 

  - **`rules`**  - _[array of object]_ -  

      - **`action`**  - _[string]_ -  

      - **`all`**  - _[boolean]_ -  

      - **`if`**  - _[string]_ -  

      - **`names`**  - _[array of string]_ -  

      - **`sql`**  - _[string]_ -  

      - **`type`**  - _[string]_ -   _(required)_

  - **`access`**  - _[one of]_ - Expression indicating if the user should be granted access to the dashboard. If not defined, it will resolve to false and the dashboard won't be accessible to anyone. Needs to be a valid SQL expression that evaluates to a boolean. 

     *option 1* - _[string]_ - 

     *option 2* - _[boolean]_ - 

  - **`exclude`**  - _[array of object]_ - List of dimension or measure names to exclude from the dashboard. If exclude is defined all other dimensions and measures are included 

      - **`if`**  - _[string]_ - Expression to decide if the column should be excluded or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

      - **`names`**  - _[any of]_ - List of fields to exclude. Should match the name of one of the dashboard's dimensions or measures  _(required)_

  - **`include`**  - _[array of object]_ - List of dimension or measure names to include in the dashboard. If include is defined all other dimensions and measures are excluded 

      - **`if`**  - _[string]_ - Expression to decide if the column should be included or not. It can leverage templated user attributes. Needs to be a valid SQL expression that evaluates to a boolean  _(required)_

      - **`names`**  - _[any of]_ - List of fields to include. Should match the name of one of the dashboard's dimensions or measures  _(required)_

**`skip_nested_security`**  - _[boolean]_ -  

 *option 1* - _[object]_ - 

 *option 1* - 

 *option 2* - 

 *option 3* - 

 *option 4* - 

 *option 5* - 