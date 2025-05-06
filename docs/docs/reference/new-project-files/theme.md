---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 10
---



## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `theme`  _(required)_

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`refs`**  - _[array of oneOf]_ - List of resource references, each as a string or map. 

  *option 1* - _[object]_ - An object reference with at least a `<name>` and `<type>`.

  - **`type`**  - _[string]_ - type of resource 

  - **`name`**  - _[string]_ - name of resource  _(required)_

  *option 2* - _[string]_ - A string reference like `<resource-name>` or `<type/resource-name>`.

**`dev`**  - _[object]_ - Overrides properties in development 

**`prod`**  - _[object]_ - Overrides properties in production 

**`colors`**  - _[anyOf]_   _(required)_

  *option 1* - _[object]_ 

  - **`primary`**  - _[string]_   _(required)_

  *option 2* - _[object]_ 

  - **`secondary`**  - _[string]_   _(required)_