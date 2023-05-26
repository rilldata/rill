---
title: Deploy a project
sidebar_label: Deploy project
sidebar_position: 0
---

Follow this tutorial to deploy an existing Rill project to Rill Cloud. Deploying a project makes its dashboards available online and enables you to invite others to access it.
Flow diagram below shows the steps needed for deploying an existing project.  
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

## Checking deployment status

Once the deployment has completed, the browser will open on your project's status page. You can also check the project status from the command-line by running:
```
rill project status
```

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

