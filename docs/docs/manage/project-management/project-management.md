---
title: "Managing Projects in Rill Cloud"
description: Basic managment from projects 
sidebar_label: "Project Management"
sidebar_position: 20
---

Once an organization is created, you can populate it with projects. Each project is connected to a [single Github repository](https://docs.rilldata.com/deploy/deploy-dashboard/#syncing-your-github-repository) and has its own set or sources, models, metric views, and explore dashboards. A defined object in a project cannot be shared to another project. 

## Project

![project](/img/manage/project-management/project-view.png) 

A project is a single deployed instance from Rill Developer (or what we refer to as a Rill project). Each project can be connected to one GitHub repository. Once you have deployed a project to Rill Cloud, you can make changes to it via the CLI or via Rill Cloud.


### Checking deployment status
After deploying to Rill Cloud, if your projects are not quite ready to view yet, you will be navigating to the status page. Here you will be able to see your component's status and if there are any issues with loading or parsing.

:::tip Refresh all source and models
You can select the `Refresh all sources and models` in the Status page or run a full project refresh. 
:::

![img](/img/manage/project-management/status.png)

### Connect to GitHub Repository 

On first deployment your project , if you've deployed via the UI, will not be connected to a GitHub repository. You will need to manually select the `Connect to GitHub` in the Status page and following the steps to `write` your current project to the repository.

![img](/img/deploy/existing-project/select-repo.png)

:::note WRITING ONLY
If the repository that you select is not empty, Rill will prompt you to `overwrite` the contents of the repository with your project file contents. You will see a commit in your repository as "Auto committed by Rill".
:::
### Modifying Github Repository

In some cases, you will need to change the repsitory that your project is synced to. You can do this by selecting the dropdown and disconnecting your Rill project. This action has no effect on your current deployment and will not require a source reingest.

![img](/img/manage/project-management/disconnect-github.png)

From there, you can follow the same steps as [above](#connect-to-github-repository) to re-connect your project to a new repository.


### Deploying from a branch other than `main`
If you have already [setup your connection to GitHub](/deploy/deploy-dashboard/#syncing-your-github-repository), you can edit the branch from where the project is deployed from.

![img](/img/manage/project-management/main-branch.png)


### Managing Project settings
You can also manage project objects in the settings page including public URLs (created in an explore dashboard) and environmental variables. For more information on managing variables, see [variables and credentials](/manage/project-management/variables-and-credentials).

![img](/img/manage/project-management/project-settings.png)

## Managing Rill project from CLI
Managing a project includes the project itself and all components or resources that belong to the project. Via the CLI, you can make changes to the project's properties such as description, GitHub branch, etc. Run `rill project -h` for an overview of available commands.

### Refreshing the deployment

Your project on Rill Cloud will automatically redeploy every time you git push changes to Github. To manually refresh data sources without pushing code changes (or redeploying your project), run the following command:

```
rill project refresh [--source/model] (source_name or model_name) [--local]
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




## Make a project public

Projects on Rill Cloud are private by default. To make a project's dashboards publicly accessible without authentication, run:
```
rill project edit --public=true
```

:::caution Avoid Sharing Private Data

**Warning**: If you make a project public, make sure it does not expose any confidential data.

:::

