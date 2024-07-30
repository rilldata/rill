---
title: User Group Management
sidebar_label: User Group Management
sidebar_position: 22
---

## Managing User groups 

In Rill Cloud, access to projects can be granted via user groups via the CLI. First, create the user group, add the required roles, then add your users to the group.

## Install and authenticate the Rill CLI

To manage cloud permissions with the Rill CLI, you must first authenticate it. If you have not already done so, run:
```
rill login
```


## Creating the User group

In order to invite a user to a user group, you need to create it first.

```
rill usergroup create
```
You will be prompted for the new user group name.

Next, you will need to add the roles and access to the user group.

```
rill usergroup add
Flags:
      --group string     User group
      --org string       Organization (default "Rill_Learn")
      --project string   Project
      --role string      Role of the user group (options: admin, viewer)
```
You will be prompted for the role and the name of the group you are editing. If you want to specify a specific project, please use the --project flag.

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