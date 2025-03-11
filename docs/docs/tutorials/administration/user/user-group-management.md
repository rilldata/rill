---
title: "User Group Management"
description:  Let's get into further details of Rill Cloud
sidebar_label: "Creating Usergroups"
sidebar_position: 12
tags:
  - CLI
  - Administration
---

import ComingSoon from '@site/src/components/ComingSoon';


## Create Usergroups

### Managing User Groups on Rill Cloud:

<ComingSoon />

<div class='contents_to_overlay'>
Historically (pre 0.48), user management was only possible via the CLI. Now, it is also possible to do so via the UI! 

</div>


### Managing User Groups via the CLI:
_**Let's see the Rill commands**_

```
Usage:
  rill usergroup [command]

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
As the name suggests, user groups are designed to group your users together so that you do not need to set permissions on each user. Simply adding the user to the group, the users will inherit permissions from the group.

By default, some system-managed groups will be created in your project. Let's add a `tutorial-admin` group.
```
rill usergroup list

  NAME          ROLE   CREATED ON            UPDATED ON           
 ------------- ------ --------------------- --------------------- 
  all-users     -      2024-08-01 09:32:29   2024-08-01 09:32:29  
  all-members   -      2024-08-01 09:32:29   2024-08-01 09:32:29  
  all-guests    -      2024-08-01 09:32:29   2024-08-01 09:32:29  
```

```bash
rill usergroup create

? Enter user group name tutorial-admin
User group "tutorial-admin" created in organization "Rill_Learn"
```
Now, listing the usergroups we can see the new group created.

```bash
rill usergroup list                                                                               
  NAME             ROLE   CREATED ON            UPDATED ON           
 ---------------- ------ --------------------- --------------------- 
  all-users        -      2024-08-01 09:32:29   2024-08-01 09:32:29  
  all-members      -      2024-08-01 09:32:29   2024-08-01 09:32:29  
  all-guests       -      2024-08-01 09:32:29   2024-08-01 09:32:29  
  tutorial-admin   -      2024-08-22 01:21:37   2024-08-22 01:21:37  
  ```
Now let's give admin access to the group for the project `my-rill-tutorial`.


```
rill usergroup add --project my-rill-tutorial
? Select role admin
? Enter user group name tutorial-admin
Role "admin" added to user group "tutorial-admin" in project "my-rill-tutorial"
```

Next, let's try to add your user to this group and see what happens.

```bash
rill user add --group tutorial-admin
? Enter email <your_email>@domain.com 
? The user must be a member of "<your_org>" to join one of its groups. Do you want to invite the user to join "<your_org>"? Yes
User "<your_email>@domain.com " added to the organization "<your_org>" as "viewer"
User "<your_email>@domain.com " added to the user group "tutorial-admin"
```

Since the user was added to the project only, not the organization, when adding the user to a usergroup (which requires the user to be a part of the organization), we will prompt if you'd like to invite the user to the organization. Once added to the organization, the user will be added to the usergroup.

Let's confirm that the user is part of the user group.

```bash
rill user list --group tutorial-admin
```

Let's navigate to our project, my-rill-tutorial, to see some differences between viewer and admin.

<img src = '/img/tutorials/201/viewervsadmin.gif' class='rounded-gif' />
<br />

For a detailed description of the differences between `admin` and `viewer`, please refer to our <a href='https://docs.rilldata.com/manage/roles-permissions' target=' blank'> Roles and permissions documentation. </a>

Note that after you add the user to the organization, you can now see the user when running `rill user list`.

