---
title: "Project Management"
description: Basic managment from projects 
sidebar_label: "Project Management"
sidebar_position: 31
---


## Make a project public

Projects on Rill Cloud are private by default. To make a project's dashboards publicly accessible without authentication, run:
```
rill project edit --public=true
```

:::caution Avoid Sharing Private Data

Warning: If you make a project public, make sure it does not expose any confidential data.

:::


## Example of Project Access


In the following example, you can see the different levels of access to Rill via the organization, project-specific access and group and user privileges.


![img](/img/manage/project-management/project-access.png)

A few key things to note:
1. There are three levels of access: organization, project, and group.
2. User groups can currently only be applied to users within the organization. 
    - pink_user can not be added to a user group
3. All users added to an organization must have at least `viewer` privilege. 
    - yellow_user is redundant as there's already `viewer` access