---
note: GENERATED. DO NOT EDIT.
title: Component YAML
sidebar_position: 34
---

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `component` _(required)_

### `display_name`

_[string]_ - Refers to the display name for the component 

### `description`

_[string]_ - (no description) 

### `input`

_[array of object]_ - (no description) 

  - **`name`** - _[string]_ - (no description) _(required)_

  - **`type`** - _[string]_ - (no description) _(required)_

  - **`value`** - _[string, number, boolean, object, array]_ - The value can be of any type. 

### `output`

_[object]_ - (no description) 

  - **`name`** - _[string]_ - (no description) _(required)_

  - **`type`** - _[string]_ - (no description) _(required)_

  - **`value`** - _[string, number, boolean, object, array]_ - The value can be of any type. 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 