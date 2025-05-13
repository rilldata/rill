---
note: GENERATED. DO NOT EDIT.
title: Component YAML
sidebar_position: 34
---



## Properties

### `type`

_[string]_ - Refers to the resource type and must be `component`  _(required)_

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references, each as a string or map. 

### `dev`

_[object]_ - Overrides properties in development 

### `prod`

_[object]_ - Overrides properties in production 

### `display_name`

_[string]_ - Refers to the display name for the component 

### `description`

_[string]_  

### `input`

_[array of object]_  

  - **`name`** - _[string]_   _(required)_

  - **`type`** - _[string]_   _(required)_

  - **`value`** - The value can be of any type. 

### `output`

_[object]_  

  - **`name`** - _[string]_   _(required)_

  - **`type`** - _[string]_   _(required)_

  - **`value`** - The value can be of any type. 