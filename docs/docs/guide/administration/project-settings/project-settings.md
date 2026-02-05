---
title: "Managing Projects in Rill Cloud"
description: Basic managment from projects 
sidebar_position: 20
---

Once an organization is created, you can populate it with projects. Each project has its own set of sources, models, metrics views, and explore dashboards. A defined object in a project cannot be shared to another project.

:::info GitHub Integration
Projects can be connected to a GitHub repository for continuous deployment. See [GitHub Integration](/guide/administration/project-settings/github-integration) for details on connecting and managing repository connections.
::: 

## Project

<img src = '/img/manage/project-management/project-view.png' class='rounded-gif' />
<br />


A project is a single deployed instance from Rill Developer (or what we refer to as a Rill project). Once you have deployed a project to Rill Cloud, you can make changes to it via the CLI or via Rill Cloud.


### Checking deployment status
After deploying to Rill Cloud, if your projects are not quite ready to view yet, you will be navigating to the status page. Here you will be able to see your component's status and if there are any issues with loading or parsing.

:::tip Refresh all source and models
You can select the `Refresh all sources and models` in the Status page or run a full project refresh. 
:::

<img src = '/img/manage/project-management/status.png' class='rounded-gif' />
<br />


### Managing Project settings
You can also manage project objects in the settings page including public URLs (created in an explore dashboard) and environmental variables. For more information on managing variables, see [variables and credentials]( /guide/administration/project-settings/variables-and-credentials).

<img src = '/img/manage/project-management/project-settings.png' class='rounded-gif' />
<br />


## Managing Rill project from CLI
Managing a project includes the project itself and all components or resources that belong to the project. Via the CLI, you can make changes to the project's properties such as description, public access, etc. Run `rill project -h` for an overview of available commands.

### Refreshing the deployment

If your project is connected to a GitHub repository, it will automatically redeploy every time you push changes. To manually refresh data sources without pushing code changes (or redeploying your project), run the following command:

```
rill project refresh [--source/model] (source_name or model_name) [--local]
```


### Checking deployment status

In case you need to check the project status via the CLI, you can run the following:
```
rill project status
```




## Make a project public

Projects on Rill Cloud are private by default. To make a project's dashboards publicly accessible without authentication, run:
```
rill project edit --public=true
```

:::caution Avoid Sharing Private Data

**Warning**: If you make a project public, make sure it does not expose any confidential data.

:::

