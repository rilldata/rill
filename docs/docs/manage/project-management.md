---
title: "Organizations and Projects (Rill Cloud)"
description: Basic managment from projects 
sidebar_label: "Organizations and Projects"
sidebar_position: 19
---

Once a project is ready to be deployed onto Rill Cloud, as an admin, you will need to manage your organization, project, and user access. Depending on where you set this access, the permissions can vary. Please see our [Roles and Permission](roles-permissions.md) page for more details.

## Organization

An Organization in Rill is the parent management object and encompasses how your team or organization interfaces with Rill Cloud. Organizations are designed to hold the differnet components of your Rill project. Projects exist within an organization, which itself contains sources, models, dashboards, and other resources that belong to your standalone Rill projects.

If you'd like to create, edit, modify, or delete an organization, run the following command.

```
rill org
Manage organisations

Usage:
  rill org [command]

Available Commands:
  create      Create organization
  edit        Edit organization details
  switch      Switch to other organization
  list        List all organizations
  delete      Delete organization
  rename      Rename organization

Global Flags:
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)
```

:::tip

[Access to Rill can be granted on the organization level](user-management.md#adding-a-member-to-the-organization).

:::

## Project

A project is a single deployed instance from Rill Developer (or what we refer to as a Rill project). Each project can be connected to one GitHub repository. Once you have deployed a project to Rill Cloud, you can make changes to it via the CLI or the UI.

### CLI
Managing a project includes the project itself and all components or resources that belong to the project. Via the CLI, you can make changes to the project's properties such as description, GitHub branch, etc using the following:
```
rill  project 
```

#### Updating the deployment

Your project on Rill Cloud will automatically redeploy every time you git push changes to Github. To manually refresh data sources without pushing code changes (or redeploying your project), run the following command:

```
rill project refresh
```


#### Checking deployment status

In case you need to check the project status via the CLI, you can run the following:
```
rill project status
```

#### Deploying from a branch other than `main`
A branch from which continuous deployment is setup can be changed while editing the project. To change the branch, run the following command:
```
rill project edit
```

### UI

#### Checking deployment status
After deploying to Rill Cloud, you will be navigating to the status page. Here you will be able to see your component's status and if there are any issues with loading or parsing.

![img](/img/manage/project-management/status.png)


#### Deploying from a branch other than `main`
If you have already setup your connection to GitHub, you can edit the branch from where the project is deployed from.

![img](/img/manage/project-management/main-branch.png)

:::tip

[Access to Rill can be granted from the project level](user-management.md#adding-a-member-to-a-specific-project).

:::


## Make a project public

Projects on Rill Cloud are private by default. To make a project's dashboards publicly accessible without authentication, run:
```
rill project edit --public=true
```

:::caution Avoid Sharing Private Data

**Warning**: If you make a project public, make sure it does not expose any confidential data.

:::


## Reference: Walking through access levels


In the following example, you can see the different levels of access to Rill via the organization, project-specific access, user group and user privileges.


<img src = '/img/manage/project-management/project-access.png' class='rounded-gif' />


### Key things to note
1. There are **three** levels of access: organizations, projects, and groups.
2. User groups can _only exist_ within an organization.
    - In the case of adding a user who is not part of the organization to a user group, you will prompted to add them first.
3. User groups permissions can either be added for the organization as a whole, or specific projects.
    - `rill usergroup create [--project project_name]`    
4. All users added to an organization must have at least `viewer` privilege. 
    - In the above diagram, `User 5` is redundant as there's already `viewer` access.