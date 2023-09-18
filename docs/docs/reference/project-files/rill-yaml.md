---
title: Project YAML
sidebar_label: Project YAML
sidebar_position: 40
---

The `rill.yaml` file contains metadata about your project.

## Properties

- _**`title`**_ — the name of your project which will be displayed in the upper left hand corner
- _**`compiler`**_ — the Rill project compiler version compatible with your project files (currently only supports: `rill-beta`)
- _**`mock_users`**_ — a list of mock users to test against dashboard [security policies](../../develop/security). For each mock user, possible attributes include:
  - _**`email`**_ — the mock user's email _(required)_
  - _**`name`**_ — the mock user's name
  - _**`admin`**_ — whether or not the mock user is an admin