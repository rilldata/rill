---
note: GENERATED. DO NOT EDIT.
title: Component YAML
sidebar_position: 4
---



## Properties


**`type`**  - _[string]_ - Refers to the resource type and must be `component`  _(required)_

**`name`**  - _[string]_ - Name is usually inferred from the filename, but can be specified manually. 

**`namespace`**  - _[string]_ - Optional value to group resources by. Prepended to the resource name as `<namespace>/<name>`. 

**`refs`**  - _[array]_ - List of resource references, each as a string or map. 

     *option 1* - _[string]_ - A string reference like 'resource-name' or 'Kind/resource-name'.

     *option 2* - _[object]_ - An object reference with at least a 'name' and 'type'.

    - **`name`**  - _[string]_ -   _(required)_

    - **`type`**  - _[string]_ -  

**`version`**  - _[integer]_ - Version of the parser to use for this file. Enables backwards compatibility for breaking changes. 

**`description`**  - _[string]_ -  

**`display_name`**  - _[string]_ - Refers to the display name for the component 

**`input`**  - _[array of object]_ -  

    - **`name`**  - _[string]_ -   _(required)_

    - **`type`**  - _[string]_ -   _(required)_

    - **`value`**  - The value can be of any type. 

**`output`**  - _[object]_ -  

  - **`name`**  - _[string]_ -   _(required)_

  - **`type`**  - _[string]_ -   _(required)_

  - **`value`**  - The value can be of any type. 