---
title: User Group Management
sidebar_label: User Groups
sidebar_position: 20
---

import ComingSoon from '@site/src/components/ComingSoon';

## Managing User groups 

In Rill Cloud, access to projects can be granted via user groups via the CLI. 

```
rill usergroup

      Available Commands:
      create      Create a user group
      rename      Rename a user group
      edit        Edit a user group
      show        Show a user group
      list        List user groups
      delete      Delete a user group
      add         Add role to a user group in an organization or project
      set         Set role to a user group in an organization or project
      remove      Remove role of a user group in an organization or prodject
```


## Install and authenticate the Rill CLI

To manage cloud permissions with the Rill CLI, you must first authenticate it. If you have not already done so, run:
```
rill login
```


## Creating the User group

You can create a new user group by running the following and following the CLI instructions:

```
rill usergroup create
```
You will be prompted for the new user group name.

### Adding permissions to the group
Next, you will need to add the roles and access to the user group.

```
rill usergroup add --project <project_name>

      --group string     User group
      --org string       Organization (default "Rill_Learn")
      --project string   Project
      --role string      Role of the user group (options: admin, viewer)
```
You will be prompted for the role and the name of the group you are editing. If you want to specify a specific project, please use the --project flag. If no project flag is defined, you will be setting permission on the organization level.

- **Viewers** can view the project's dashboards
- **Admins** can additionally edit the project, and view and edit project members

### Add a member to the group

To add a member to the user group, run the following command:
```
rill user add --group <group_name>
```

You will be prompted for the email address for the user.

Once added, you can confirm the user group by running the following command:
```
rill user list --group <group_name>
```


### Other actions
Run `rill usergroup --help` to show commands for listing members or changing access.


:::note
Currently, users outside of the organization are unable to be added to usergroups.
:::