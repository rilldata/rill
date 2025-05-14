---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 40
---



## Properties

### `type`

_[string]_ - Refers to the resource type and must be `theme`  _(required)_

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references, each as a string or map. 

### `dev`

_[object]_ - Overrides properties in development 

### `prod`

_[object]_ - Overrides properties in production 

### `colors`

_[anyOf]_   _(required)_

  **&nbsp;&nbsp;&nbsp;&nbsp;option 1** - _[object]_ 

  - **`primary`** - _[string]_   _(required)_

  **&nbsp;&nbsp;&nbsp;&nbsp;option 2** - _[object]_ 

  - **`secondary`** - _[string]_   _(required)_