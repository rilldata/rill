---
title: Roles and Permissions
description: Learn more about roles and permissions for organizations and projects in Rill Cloud
sidebar_label: Roles and Permissions
sidebar_position: 30
---

Access permissions in Rill Cloud are organized into roles at the organization and project level.

## Role inheritance

Organization and project level roles are managed separately, but are connected in several ways:
1. By default, all organization members (but not guests) are added to new projects through a user group membership with the **viewer** role. You can manually remove this relationship in the project's member settings.
2. If you grant a project level role to someone who is not a member of the parent organization, they will automatically be added to the organization with the **guest** role.
3. Removing an organization member or guest automatically also removes them from all projects in the organization.
4. Organization admins implicitly have admin privileges on all projects in the organization.

## Organization-level permissions

There are four roles available at the organization-level: **Admin**, **Editor**, **Viewer** and **Guest**.

| Permission           | Description                                         | Admin | Editor | Viewer | Guest |
| :------------------- | :-------------------------------------------------- | ----: | -----: | -----: | ----: |
| `read_org`           | View basic info about the organization              |     ✔ |      ✔ |      ✔ |     ✔ |
| `manage_org`         | Change organization settings                        |     ✔ |        |        |       |
| `read_projects`      | Act as a viewer on all projects in the organization |     ✔ |      ✔ |      ✔ |     ✔ |
| `create_projects`    | Create new projects in the organization             |     ✔ |      ✔ |        |       |
| `manage_projects`    | Act as an admin on all projects in the organization |     ✔ |        |        |       |
| `read_org_members`   | View members of the organization                    |     ✔ |      ✔ |        |       |
| `manage_org_members` | Add, remove or change roles of organization members |     ✔ |      ✔ |        |       |

## Project-level permissions

There are three roles available at the project-level: **Admin**, **Editor**, and **Viewer**.

| Permission                     | Description                                                | Admin | Editor | Viewer |
| :----------------------------- | :--------------------------------------------------------- | ----: | -----: | -----: |
| `read_project`                 | View basic info about the project                          |     ✔ |      ✔ |      ✔ |
| `manage_project`               | Change project settings                                    |     ✔ |        |        |
| `read_prod`                    | View dashboards deployed from the production (main) branch |     ✔ |      ✔ |      ✔ |
| `read_prod_status`             | View logs for the production deployment                    |     ✔ |        |        |
| `manage_prod`                  | Trigger actions on the production deployment               |     ✔ |        |        |
| `read_provisioner_resources`   | View managed resources for the project                     |     ✔ |        |        |
| `manage_provisioner_resources` | Add or remove managed resources for the project            |     ✔ |        |        |
| `read_project_members`         | View members of the project                                |     ✔ |      ✔ |        |
| `manage_project_members`       | Add, remove or change roles of project members             |     ✔ |      ✔ |        |
| `create_magic_auth_tokens`     | Create shareable URLs                                      |     ✔ |      ✔ |        |
| `manage_magic_auth_tokens`     | Remove shareable URLs created by others                    |     ✔ |      ✔ |        |
| `create_reports`               | Create and edit new scheduled reports                      |     ✔ |      ✔ |      ✔ |
| `manage_reports`               | Edit and change scheduled reports created by others        |     ✔ |        |        |
| `create_alerts`                | Create and edit new alerts                                 |     ✔ |      ✔ |      ✔ |
| `manage_alerts`                | Edit and change alerts created by others                   |     ✔ |        |        |
| `create_bookmarks`             | Create and edit new bookmarks                              |     ✔ |      ✔ |      ✔ |
| `manage_bookmarks`             | Edit and change bookmarks created by others                |     ✔ |        |        |
