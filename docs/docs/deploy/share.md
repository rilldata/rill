---
title: Share with others
sidebar_label: Share with others
sidebar_position: 20
---

In Rill Cloud, access can be granted at the organization or project level using the Rill CLI.

## Install and authenticate the Rill CLI

If you have not already installed the Rill CLI, see [Install Rill](../develop/install.md).

To manage cloud permissions with the Rill CLI, you must first authenticate it. If you have not already done so, run:
```
rill login
```

## Managing members of an organization

When you invite a user to an organization on Rill Cloud, they automatically get access to *all projects* in the organization. Users can have one of two roles:

- **Viewers** can browse projects and view dashboards
- **Admins** can manage projects by deploying new projects, making changes to existing projects, or deleting deployed projects. They can also manage members of an organization by granting or revoking access to other users.  
  
### Add a member

To add a member to an organization, run the following command:
```
rill user add
```
You will then be prompted for details about the user.

If you add a user who has not yet signed up for Rill, they will receive an email inviting them to sign up and join.

### Other actions

Run `rill user --help` to show commands for listing members or changing access.

## Managing members of a project

By default, adding a user to an organization grants them access to all its projects. You can alternatively add a user only to a specific project. Users can have one of two roles on a project:

- **Viewers** can view the project's dashboards
- **Admins** can additionally edit the project, and view and edit project members

### Add a member

To add a member to a project, run the following command:
```
rill user add --project [PROJECT NAME]
```
You will then be prompted for details about the user. HINT: Run `rill project list` to show available projects.

If you add a user who has not yet signed up for Rill, they will receive an email inviting them to join.

### Other actions

Run `rill user --help` to show commands for listing members or changing access.

## Make a project public

Projects on Rill Cloud are private by default. To make a project's dashboards publicly accessible without authentication, run:
```
rill project edit --public=true
```

Warning: If you make a project public, make sure it does not expose any confidential data.
