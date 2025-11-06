---
title: User group Permissions
sidebar_label: User Group Management 
sidebar_position: 24
---

Creating user groups in Rill allows administrators to easily grant permission to multiple projects at different access levels. It is possible to mix and match viewer and administrator permission in a single group and users can be part of multiple groups. However, please keep in mind that the higher permission will be applied.

## Managing User groups Permissions
There are two ways to set up user groups in Rill.

1. Administrator via Rill Cloud
2. Administrator via CLI 

### How to Manage User Groups in Rill Cloud
From the organization page, you can manage user groups under the Users tab. Adding user groups from this page will add the user group to the organization. You can then add users to a user group to inherit the group [permissions](/manage/roles-permissions).

<img src = '/img/manage/user-management/usergroup-management.png' class='rounded-gif' />
<br />

### How to Manage User Groups via the CLI
```
rill usergroup
Manage user groups

Usage:
  rill usergroup [command]

Available Commands:
  list        List groups
  show        Show group
  create      Create a group
  rename      Rename a group
  edit        Edit a group
  delete      Delete a group
  add         Add a group to a project or organization
  set-role    Change a group's role on a project or organization
  remove      Remove a group's role on a project or organization
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
      --role string      Role of the user group (options: admin, editor, viewer)
```
You will be prompted for the role and the name of the group you are editing. If you want to specify a specific project, please use the --project flag. If no project flag is defined, you will be setting permission on the organization level.

If you have any questions on permission levels, please review the [Roles and Permissions page](roles-permissions).

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


## Reference: Walking through access levels

In the following example, you can see the different levels of access to Rill via the organization, project-specific access, user group and user privileges.

<img src = '/img/manage/project-management/project-access.png' class='rounded-gif' />

### Key things to note
1. There are **three** kinds of access: organization, project, and group.
2. User groups can _only exist_ within an organization.
    - In the case of adding a user who is not part of the organization to a user group, you will prompted to add them first.
