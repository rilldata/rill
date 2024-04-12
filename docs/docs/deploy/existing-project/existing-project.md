---
title: Deploy Dashboards 
sidebar_label: Deploy Dashboards 
sidebar_position: 00
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview

Follow this tutorial to deploy an existing Rill project to Rill Cloud. Deploying a project makes its dashboards available online and enables you to invite others to access it. Benefits of deploying your project:

- Share dashboards with other users
- Leverage Rill Cloud capabililies like [scheduled reports](/explore/exports.md) and [alerts](explore//alerts/alerts.md) 
- [Embed Rill](/integrate/embedding.md) in other applications

The flow diagram below shows the steps needed for deploying an existing project.  
```mermaid
graph LR;
    A(Local code files);
    B(GitHub);
    C(Rill Cloud);
    A--Pushed -->B;
    B--Continuous Deployment-->C;
```
    
## Push the project to Github

Rill Cloud connects to a repository on Github containing a Rill project, and continuously deploys that project on every push. Rill Cloud has the ability to auto-create a Git repository on your behalf when first deploying your project or you have the option to manually create the Git repository yourself before deploying the project to Rill Cloud.

### Automated repository creation

If you'd like Rill Cloud to automaticaly create the Git repository for a Rill project that you deploy, you can skip to the [next step](#deploy-to-rill-cloud).

:::note GitHub app permissions

This assumes that the installed Github app in your organization has write access. If unsure, please check with your Github admin.

:::

### Manual repository creation

If you'd like to create the Git repository manually, the project must be on Github <u>before</u> you deploy it to Rill.
- If your project is not yet on Github, you can follow the steps on Github [here](https://github.com/new) to create a new repository and push your project files to it.
- If your project is already on Github, make sure you have appropriate permissions to grant access to it. If you're deploying a project controlled by someone else, you may need to fork or copy it to a repository in your account.

:::info Custom Git repository name

When Rill attempts to create a Git repository on your behalf, _the new repository will mirror the name of your Rill project_. If you'd like more flexibility and/or to give the Git repository a different name, you should create the repository manually.

:::

## Deploy to Rill Cloud

To deploy a project to Rill Cloud, from the directory containing your project, it's as simple as running:

```
rill deploy
```

The CLI will guide you through authenticating with Rill Cloud and granting appropriate access to your Rill project on Github.

:::tip Configure credentials
Cloud datastores will typically require service keys to access data. Make sure to create the necessary key for your service account and then run ```rill env configure``` with the correct credentials. For more details, please refer to our [connector documentation](/build/credentials/credentials.md).

:::

### First deployment

If this is your first deployment to Rill Cloud, you will get prompted to either sign up or log in (if you have an existing account on [Rill Cloud](https://ui.rilldata.com/)). Proceed with the sign up and email verification process for new users or authorization process for existing users. As a new user, you can expect to see the following page:

![Sign in to Rill Cloud](/img/deploy/existing-project/rill-cloud-sign-in.png)

After successfully signing in and/or authorizing the Rill CLI, you will get prompted to connect to Github when deploying your project.

![Connect to Github](/img/deploy/existing-project/connect-github.png)

:::warning Select the correct Github organization when installing the Rill Cloud app

Make sure that you are selecting the correct Github organization when installing / connecting the Rill Cloud app. The organization should correspond to where you want your repositories to belong to when deploying Rill projects.

:::

After connecting Rill Cloud to Github and selecting a [default organization](/reference/cli/org) in the CLI, you should now be able to continuously deploy new projects and/or update existing projects. These projects, [unless specified otherwise](/reference/cli/deploy), will be deployed to your selected [organization](/manage/user-management#managing-members-of-an-organization).

:::info Check with your Github organization admin

If you're not the admin of your Github organization, they will likely need to first install the Rill Cloud app in your organization before you can proceed with deploying a project. After the Rill Cloud app is installed, it should have the following privileges:

![Github app permissions](/img/deploy/existing-project/github-app-permissions.png)

:::

When deploying a project, Rill Cloud will first check whether there is a git remote present. If there is no git remote associated, you should get prompted whether you'd like Rill to create a Git repository on your behalf. If you enter **`Y`**, a new Git repository will be created and the project will be deployed.

:::warning Beware of existing repositories with the same name!

Rill Cloud will automatically attempt to create a Git repository using the <u>same name</u> as your Rill project for auto-deployments. If a Git repository with the same name already exists, you should get prompted and receive a warning in the CLI.

:::

## Checking deployment status

Once the deployment has completed, the browser will open on your project's status page. Alternatively, you can check the project status from the command-line (or CLI) by running the following command:
```
rill project status
```

:::info Resetting an Errored Project
Projects can sometimes be in an error state for a variety of a reasons. A hard reset can often clear these issues.

To execute a hard reset of your project deployment, you can use the `rill project reset` command from the CLI.
:::

## Updating the deployment

Your project on Rill Cloud will automatically redeploy every time you `git push` changes to Github.

To manually refresh data sources without pushing code changes (or redeploying your project), run the following command:
```
rill project refresh
```

# Change your production branch

By default, Rill deploys from the [default branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches#about-the-default-branch) of your Git repository. You can change this to any branch you want.

To deploy your project from a different branch, run the following command:

```bash
rill deploy --prod-branch [PROD-BRANCH]
```

## Deploying from a branch other than `main`
A branch from which continuous deployment is setup can be changed while editing the project. To change the branch, run the following command:
```
rill project edit
```

## Deploy from a monorepo

If your Rill project is in a sub-directory of a Git repository, use the `--subpath` option when creating your project:
```
rill deploy --subpath path/to/rill/project
```
:::warning
Note that you must run `rill deploy` from the <u>root</u> of your Git repository, **not** the root of your Rill project.
:::

