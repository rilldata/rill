---
title: Roles and permissions
description: Learn more about roles and permissions for organizations and projects in Rill Cloud
sidebar_label: Roles and permissions
sidebar_position: 30
---

Access permissions in Rill Cloud are organized into roles at the organization and project level. It is usually sufficient to grant access at the organization-level because those permissions are inherited for projects by default. 

For information about granting, revoking and changing roles, see [Share with others](./share-rill.md).

## Role inheritance

Some project-level roles can be inherited from the organization-level:

- Users with `read_projects` permission on an organization get *viewer* role on all projects in the organization.
- Users with `manage_projects` permission on an organization get *admin* role on all projects in the organization.

## Organization-level permissions

There are two roles available at the organization-level: **Viewer** and **Admin**.

| Permission           | Description                                         | Viewer | Admin |
| :------------------- | :-------------------------------------------------- | -----: | ----: |
| `read_org`           | View basic info about the organization              |      ✔ |     ✔ |
| `manage_org`         | Change organization settings                        |        |     ✔ |
| `read_projects`      | Act as a viewer on all projects in the organization |      ✔ |     ✔ |
| `create_projects`    | Create new projects in the organization             |        |     ✔ |
| `manage_projects`    | Act as an admin on all projects in the organization |        |     ✔ |
| `read_org_members`   | View members of the organization                    |        |     ✔ |
| `manage_org_members` | Add, remove or change roles of organization members |        |     ✔ |

## Project-level permissions

There are two roles available at the project-level: **Viewer** and **Admin**.

| Permission               | Description                                                | Viewer | Admin |
| :----------------------- | :--------------------------------------------------------- | -----: | ----: |
| `read_project`           | View basic info about the project                          |      ✔ |     ✔ |
| `manage_project`         | Change project settings                                    |        |     ✔ |
| `read_prod`              | View dashboards deployed from the production (main) branch |      ✔ |     ✔ |
| `read_prod_status`       | View logs for the production deployment                    |        |     ✔ |
| `manage_prod`            | Trigger actions on the production deployment               |        |     ✔ |
| `read_project_members`   | View members of the project                                |        |     ✔ |
| `manage_project_members` | Add, remove or change roles of project members             |        |     ✔ |
<!--
| `read_dev`               | View dashboards deployed from non-production branches      |        |     ✔ |
| `read_dev_status`        | View logs for non-production deployments                   |        |     ✔ |
| `manage_dev`             | Trigger actions on non-production deployments              |        |     ✔ |
 -->
