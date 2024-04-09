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

Rill Cloud connects to a repository on Github containing a Rill project, and continuously deploys that project on every push. Therefore, your project must be on Github before you deploy it to Rill.

- If your project is not yet on Github, follow the steps on [https://github.com/new](https://github.com/new) to create a new repository and push your project files to it.
- If your project is already on Github, make sure you have permission to grant access to it. If you're deploying a project controlled by someone else, first fork or copy it to a repository in your account.

## Deploy to Rill Cloud

With your project files on Github, you're ready to deploy the project. In the directory containing your project, run:

```
rill deploy
```

The CLI will guide you through authenticating with Rill Cloud and granting read-only access to your Rill project on Github.

:::tip configure credentials
Cloud datastores will require service keys to access data. Make sure to create the necessary key and then run ```rill env configure``` with the correct credentials. 

More details on credentials by source can be found in our [connectors section](/build/credentials/credentials.md). 
:::


## Checking deployment status

Once the deployment has completed, the browser will open on your project's status page. You can also check the project status from the command-line by running:
```
rill project status
```

:::info Resetting an Errored Project
Projects can sometimes be in an error state for a variety of a reasons. A hard reset can often clear these issues.

To execute a hard reset, use the following command from the CLI `rill project reconcile --reset` 
:::

## Updating the deployment

Your project on Rill Cloud will automatically re-deploy every time you `git push` changes to Github.

To refresh data sources without pushing code changes, run:
```
rill project reconcile --refresh
```

# Change your production branch

By default, Rill deploys from the [default branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches#about-the-default-branch) of your Git repository. You can change this to any branch you want.

To deploy your project from a different branch, run the following command:

```bash
rill deploy --prod-branch [PROD-BRANCH]
```

## Deploying from a branch other than `main`
A branch from which continuous deployment is setup can be changed while editing the project. To change the branch, run:
```
rill project edit
```

## Deploy from a monorepo

If your Rill project is in a sub-directory of a Git repository, use the `--subpath` option when creating your project:
```
rill deploy --subpath path/to/rill/project
```
Note that you must run `rill deploy` from the root of the Git repository, not the root of the Rill project.

