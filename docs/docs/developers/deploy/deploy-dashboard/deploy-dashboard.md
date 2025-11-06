---
title: Deploy Dashboards 
sidebar_label: Deploy Dashboards 
sidebar_position: 00
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## Overview


Deploying dashboards from Rill Developer allows you to share dashboards with other users, leverage [Rill Cloud capabilities](../../explore/dashboard-101), [embed Rill](/integrate/embedding) into other applications, and more!

:::tip Configure credentials
Cloud datastores will typically require service keys to access data. Make sure you create the necessary key for your service account and either add these credentials to your `.env` file directly or deploy your project and then run ```rill env configure``` with the correct credentials. For more details, please refer to our [connector documentation](/build/connectors/credentials).
:::

The flow diagram below shows two options for deploying an existing project. 

**Deploy via the UI or CLI using `rill project deploy`**: 
```mermaid
graph LR;
    A(Local code files);
    B(Rill Cloud);
    A--deploy (once)/update via UI -->B;

```
---
**Deploy via the CLI via `rill project connect-github`**:
```mermaid
graph LR;
    A(Local code files);
    B(Rill Cloud);
    C(GitHub);


    A--1. deploy via CLI (once) -->B;
    C-- Continuous Deployment-->B;
    A--2. Push changes to GitHub-->C;
```
    
## Deploying a project from Rill Developer
Starting from **v0.48**, we have introduced the possibility to push dashboards _directly from Rill Developer to Rill Cloud_. On the dashboard page, you can select the `Deploy` button and follow the steps to deploy to Rill Cloud.

<img src = '/img/deploy/existing-project/deploy-ui.gif' class='rounded-gif' />
<br />

Now that you project has been deployed to Rill Cloud, you will need to ensure that your users have access! Please refer to the [user management](../../manage/user-management) section. 

If you make changes locally on Rill Developer, you will need to push the contents to Rill Cloud by selecting the `Update` button.

<img src = '/img/deploy/existing-project/redeploy.gif' class='rounded-gif' />
<br />

:::tip On an older version of Rill?

You can easily check the version of Rill that you are using in Rill Developer by running the following command:

```bash
rill --version
```

If you are on an older version of Rill, it is **strongly recommended** to [upgrade](/get-started/install#upgrade-to-the-newest-version-of-rill-developer) to the latest version.

:::

### Syncing your GitHub Repository
:::note GitHub app permissions
This assumes that the installed GitHub app in your organization has write access. If unsure, please check with your GitHub admin.

The required permissions are:
 - Read access to metadata and pull requests
 - Read and write access to administration and code 
:::


At this point, you have the option to connect your Rill project to a GitHub Repository.

Navigating to the Status page and selecting `Connect to GitHub` will prompt you to login and create a repository for your project. If you've already created a repository, check the box 'I've created a GitHub Repo' and add the permissions for Rill to access the repository.

:::info Check with your GitHub organization admin

If you're not the admin of your GitHub organization, they will likely need to first install the Rill Cloud app in your organization before you can proceed with deploying a project. After the Rill Cloud app is installed, it should have the following privileges:
:::



<img src = '/img/deploy/existing-project/install-rill-cloud.png' class='rounded-gif' />
<br />


Once the permissions to the repository have been confirmed and set, you can continue to select the repository in the dropdown.
<img src = '/img/deploy/existing-project/select-repo.png' class='rounded-gif' />
<br />


Once completed, you'll see the newly updated repository on the UI of your status page!

<img src = '/img/deploy/existing-project/finished.png' class='rounded-gif' />
<br />


:::warning Still unable to connect?
If you encounter issues, check that the app installation is not pending. Go to your organization's settings and click on Installed GitHub Apps. You will see a section of Pending GitHub Apps installation requests. If you're an Owner or App Manager, grant access to the Rill app if it is pending."
:::


## Deploying a project via the CLI

:::note
Starting from v0.49, we have deprecated `rill deploy` in favor of `rill project deploy` and `rill project connect-github`. For more information on the `rill deploy` command click [here](#deprecated-rill-deploy).
:::

### Deploy project without GitHub Repository
You can add a GitHub Repository later.
```
rill project deploy
Using org "Rill_Learn".

Starting upload.
All files uploaded successfully.

Created project "Rill_Learn/my-rill-tutorial". Use `rill project rename` to change name if required.

...

Your project can be accessed at: https://ui.rilldata.com/Rill_Learn/my-rill-tutorial
Opening project in browser...
```

If you have not already [configured your connections' credentials](https://docs.rilldata.com/build/connectors/credentials), you will be reminded here which connections are required.

**First deployment**

If this is your first deployment to Rill Cloud, you will get prompted to either sign up or log in (if you have an existing account on [Rill Cloud](https://ui.rilldata.com/)). Proceed with the sign up and email verification process for new users or authorization process for existing users. As a new user, you can expect to see the following page:

<img src = '/img/deploy/existing-project/rill-cloud-sign-in.png' class='rounded-gif' />
<br />



**Project Uploaded Successfully**

Once the project has been uploaded to Rill Cloud, you should be able to see the following page: 

<img src = '/img/deploy/existing-project/status.png' class='rounded-gif' />
<br />






### Deploy Project with Repository
Follow the instructions in the Terminal to login to GitHub (if not already done so), and select your repository.
If you do not set any parameters, Rill will infer the project name based on the folder path and use this as both the repository and project name. If there are any overlaps, we will request for a new name.
```bash
rill project connect-github
No git remote was found.
? Do you want to create a repo? Yes
? Select a GitHub account for the new repository royendo
Repository name "my-rill-tutorial" is already taken
? Please provide alternate name my-rill-tutorial-cli

Request submitted for creating repository. Checking completion status

Successfully created repository on "https://github.com/royendo/my-rill-tutorial-cli"

Pushing local project to GitHub

Successfully pushed your local project to GitHub

Using org "Rill_Learn".

Created project "Rill_Learn/my-rill-tutorial-cli". Use `rill project rename` to change name if required.

Rill projects deploy continuously when you push changes to GitHub.

...

Your project can be accessed at: https://ui.rilldata.com/Rill_Learn/my-rill-tutorial-cli
Opening project in browser...
```

Once completed, you will see the following the in the status page. Note that the GitHub repository is already setup!

<img src = '/img/deploy/existing-project/cli-upload.png' class='rounded-gif' />
<br />



## Continuous Deployment 

Whether you decide to manage your Rill projects using GitHub or re-running `rill project deploy`, Rill should automatically detect changes that you have pushed locally and update your deployed project accordingly. Depending on the changes, this may result in a project reconciliation. If you are experiencing some issues with the project after pushing changes to the CLI, please refer to the project's status page for more information or you can run via the CLI:

```
rill project status
```

Likewise, if using the UI by selecting the `Update` button, Rill will detect the changes in files and update your deployed project accordingly. Along with the above CLI command, you can view the status of the objects in the Status page.

:::tip Interested in using Gitlab?

Check out our documentation on deploying a [Rill project using Gitlab](/deploy/deploy-dashboard/deploy-from-cli)!

:::


## Change your production branch

By default, Rill deploys from the [default branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches#about-the-default-branch) of your Git repository. You can change this to any branch you want.

To deploy your project from a different branch, run the following command:

```bash
rill connect-github --prod-branch [PROD-BRANCH]
```



## Deploy from a monorepo

If your Rill project is in a sub-directory of a Git repository, use the `--subpath` option when creating your project:
```
rill connect-github --subpath path/to/rill/project
```
:::warning
Note that you must run `rill connect-github` from the <u>root</u> of your Git repository, **not** the root of your Rill project.
:::



## Deprecated Rill Deploy

When running `rill deploy` you have two options: 
1. Enable automatic deploys to Rill Cloud via GitHub
2. Disable automatic deploys to Rill Cloud via GitHub

```
rill deploy
? Enable automatic deploys to Rill Cloud from GitHub? 
```

### Enable Automatic deploys

Like running `rill project connect-github`, you will be [prompted to create a github repository](#deploy-project-with-repository). Once created, Rill will deploy the project. You can confirm that the project has the correct repository linked from the UI on the status page.


### Disable Automatic deploys

In this case, the project will be deployed to Rill Cloud without a GitHub repository connected. You can always [add a repository via the UI](#syncing-your-github-repository) at a later time.