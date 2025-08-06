---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 40
---

Themes allow you to customize the appearance of your dashboards and UI components.

## Properties

### `type`

_[string]_ - Refers to the resource type and must be `theme` _(required)_

### `display_name`

_[string]_ - Display name for the theme _(required)_

### `description`

_[string]_ - Description for the theme 

### `colors`

_[object]_ - Color palette for the theme 

  - **`primary`** - _[string]_ - Primary color 

  - **`secondary`** - _[string]_ - Secondary color 

  - **`accent`** - _[string]_ - Accent color 

  - **`background`** - _[string]_ - Background color 

  - **`text`** - _[string]_ - Text color 

### `fonts`

_[object]_ - Font configuration for the theme 

  - **`family`** - _[string]_ - Font family 

  - **`size`** - _[string]_ - Base font size 

### `spacing`

_[object]_ - Spacing configuration for the theme 

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually. 

### `refs`

_[array of string]_ - List of resource references 

### `dev`

_[object]_ - Overrides any properties in development environment. 

### `prod`

_[object]_ - Overrides any properties in production environment. 