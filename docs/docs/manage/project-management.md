---
title: "Organization and Project Management"
description: Basic managment from projects 
sidebar_label: "Organization & Project Management"
sidebar_position: 19
---
## Organization

An Organization in Rill is the largest management object. Within a single organization, you can create many projects.  Within those project can contain sources, models, dashboards, etc. Organizations are designed for ...

You can create / delete / modify an organization via the CLI.
```
rill org 
```

[Access into Rill can be granted from the organization level](user-management.md#adding-a-member-to-the-organization).

## Project

A project is a single folder deployed instance from Rill Developer. Once you have deployed a project to Rill Cloud, you can make changes to it via the CLI.

```
rill  project 
```


[Access into Rill can be granted from the project level](user-management.mdt#adding-a-member-to-a-specific-project).



## Make a project public

Projects on Rill Cloud are private by default. To make a project's dashboards publicly accessible without authentication, run:
```
rill project edit --public=true
```

:::caution Avoid Sharing Private Data

Warning: If you make a project public, make sure it does not expose any confidential data.

:::


## Example of Access to Rill


In the following example, you can see the different levels of access to Rill via the organization, project-specific access, user group and user privileges.


![img](/img/manage/project-management/project-access.png)

A few key things to note:
1. There are three levels of access: organization, project, and group.
2. User groups can currently only be applied to users within the organization. 
    - pink_user can not be added to a user group
3. All users added to an organization must have at least `viewer` privilege. 
    - yellow_user is redundant as there's already `viewer` access