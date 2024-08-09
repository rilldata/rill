---
title: "Organization and Project Management"
description: Basic managment from projects 
sidebar_label: "Organization & Project Management"
sidebar_position: 19
---

Once a project is ready to be deployed onto Rill Cloud, there are a few concepts that you need to consider. As an administrator, you will need to manage your organization, project, and users access. Depending on where you set access, the permissions vary. Please see our [Roles and Permission](roles-permissions.md) page for more details.

## Organization

An Organization in Rill is the largest management object. Organizations are designed to hold the differnet components of your Rill project. Projects exists within an organization. Within each project exists sources, models, dashboards, etc. 

You can create / delete / modify / edit an organization.
### Via the CLI
```
rill org 
```


### Via the UI
import ComingSoon from '@site/src/components/ComingSoon';

<ComingSoon />

<div class='contents_to_overlay'>
a
</div>


[Access to Rill can be granted on the organization level](user-management.md#adding-a-member-to-the-organization).

## Project

A project is a single deployed instance from Rill Developer. Each project can be connected to one GitHub repository. Once you have deployed a project to Rill Cloud, you can make changes to it via the CLI or the UI.

### Via the CLI
Managing a project includes the project itself and the components within. Via the CLI, you can make changes to the project's properties such as description, GitHub branch, etc using the following:
```
rill  project 
```

### Updating the deployment

Your project on Rill Cloud will automatically redeploy every time you git push changes to Github. To manually refresh data sources without pushing code changes (or redeploying your project), run the following command:

```
rill project refresh
```


### Checking deployment status

In case you need to check the project status via the CLI, you can run the following:
```
rill project status
```

### Deploying from a branch other than `main`
A branch from which continuous deployment is setup can be changed while editing the project. To change the branch, run the following command:
```
rill project edit
```

### Via the UI

### Checking deployment status
After deploying to Rill Cloud, you will be navigating to the status page. Here you will be able to see your component's status and if there are any issues with loading or parsing.

![img](/img/manage/project-management/status.png)


### Deploying from a branch other than `main`
If you have already setup your connection to GitHub, you can edit the branch from where the project is deployed from.

![img](/img/manage/project-management/main-branch.png)


### and more!
<ComingSoon />

<div class='contents_to_overlay'>
aaa
</div>


[Access to Rill can be granted from the project level](user-management.md#adding-a-member-to-a-specific-project).



## Make a project public

Projects on Rill Cloud are private by default. To make a project's dashboards publicly accessible without authentication, run:
```
rill project edit --public=true
```

:::caution Avoid Sharing Private Data

Warning: If you make a project public, make sure it does not expose any confidential data.

:::


## Example of Different Access to Rill


In the following example, you can see the different levels of access to Rill via the organization, project-specific access, user group and user privileges.


<img src = '/img/manage/project-management/project-access.png' class='rounded-gif' />
<br />


A few key things to note:
1. There are three levels of access: organization, project, and group.
2. User groups can only exist within an organization.
3. User groups permissions can either be added for the organization as a whole, or specific projects.
    - `rill usergroup create [--project project_name]`
4. User groups can currently only be applied to users within the organization. 
    - pink_user can not be added to a user group
5. All users added to an organization must have at least `viewer` privilege. 
    - yellow_user is redundant as there's already `viewer` access