---
title: Publish Dashboards to Rill Cloud
sidebar_label: Deploy Dashboards 
sidebar_position: 15
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->


Deploying dashboards from Rill Developer allows you to share dashboards with other users, leverage [Rill Cloud capabilities](../../explore/dashboard-101), [embed Rill](/integrate/embedding.md) into other applications, and more! Simply click the Deploy button in Rill Developer and follow the steps!

<img src = '/img/deploy/existing-project/deploy.png' class='rounded-gif' />
<br />
:::tip Configure credentials
Cloud datastores will typically require service keys to access data. Make sure you create the necessary key for your service account and either add these credentials to your `.env` file directly or deploy your project and then run ```rill env configure``` with the correct credentials. For more details, please refer to our [connector documentation](/connect/credentials).
:::

## First Deploy to Rill Cloud 

When deploying to Rill Cloud for the first time via Rill Developer, you will be taken through the following steps, if not already completed.

1. Create an account / Sign into your account
<img src = '/img/deploy/existing-project/rill-cloud-sign-in.png' class='rounded-gif' />
<br />
2. Create an Organization
<img src = '/img/deploy/create-org-deploy.png' class='rounded-gif' />
<br />
3. Invite Users to Project (which will add them as org members)
<img src = '/img/deploy/invite-users.png' class='rounded-gif' />
<br />


This will result in a single project deployed onto an organization whose files are managed by Rill, and updates to the project can be done using the UI or CLI. At this point, another developer can [clone the project](/reference/cli/project/clone), make changes, and push to this deployed project. For a detailed guide, see [clone a project](/guides/clone-a-project).

If you decide to skip inviting users on the first pass, don't worry, you'll have a chance to do so again. Please refer to the [user management](/manage/user-management.md) section for more details.


:::tip On an older version of Rill?

You can easily check the version of Rill that you are using in Rill Developer by running the following command:

```bash
rill --version
```

If you are on an older version of Rill, it is **strongly recommended** to [upgrade](/get-started/install.md#upgrade-to-the-newest-version-of-rill-developer) to the latest version.

:::

## Sync Rill Project to GitHub

We recommend syncing your Rill project to your own [GitHub repository](https://docs.github.com/en/repositories/creating-and-managing-repositories/creating-a-new-repository), which allows you to maintain your files with all of GitHub's features such as version control and code review requirements. For some basic tips, see our [GitHub Basics](/deploy/github-101) page!

### Syncing your GitHub Repository

In order to sync your Rill project to your GitHub repository, you will need to provide access to your repository following the steps below:

1. Navigate to the Status page and select `Connect to GitHub`.

This will prompt you to log into GitHub and create a repository for your project. If you've already created a repository, check the box 'I've created a GitHub Repo' and follow the wizard to add the required permissions for Rill to access the repository.

<img src = '/img/deploy/existing-project/install-rill-cloud.png' class='rounded-gif' />
<br />


:::info Check with your GitHub organization admin

If you're not the admin of your GitHub organization, they will likely need to first install the Rill Cloud app in your organization before you can proceed with deploying a project. After the Rill Cloud app is installed, it should have the following privileges:
:::


2. Select your repository from the dropdown.
   
For most use cases, you will not need to adjust the advanced options. However, if you have a [monorepo and need to select a subfolder], or are using a [different branch than main], you can set those here.
<img src = '/img/deploy/existing-project/select-repo.png' class='rounded-gif' />
<br />


3. When redirected back, confirm that the repository is set.

<img src = '/img/deploy/existing-project/finished.png' class='rounded-gif' />
<br />


:::warning Still unable to connect?
If you encounter issues, check that the app installation is not pending. Go to your organization's settings and click on Installed GitHub Apps. You will see a section of Pending GitHub Apps installation requests. If you're an Owner or App Manager, grant access to the Rill app if it is pending.
:::


## Deploying a project via the CLI

Similar to deploying from Rill Developer, it is possible to deploy via the CLI. This allows for automation of project deploys/updates or simply for those who prefer this method.

### Deploy project without GitHub Repository

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

If you have not already [configured your connections' credentials](https://docs.rilldata.com/connect/credentials), you will be reminded here which connections are required.


**Project Uploaded Successfully**

Once the project has been uploaded to Rill Cloud, you should be able to see the following page: 

<img src = '/img/deploy/existing-project/status.png' class='rounded-gif' />
<br />

### Deploy Project with Repository
Follow the instructions in the Terminal to log in to GitHub (if not already done so), and select your repository.
If you do not set any parameters, Rill will infer the project name based on the folder path and use this as both the repository and project name. If there are any overlaps, we will request a new name.
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
**Project Uploaded Successfully**

Once completed, you will see the following in the status page. Note that the GitHub repository is already set up!

<img src = '/img/deploy/existing-project/cli-upload.png' class='rounded-gif' />
<br />



## Detecting Changes

If you decide to manage your Rill projects using GitHub, Rill will automatically detect changes that you have pushed locally and update your deployed project accordingly. Depending on the changes, this may result in a project reconciliation. If you are experiencing issues with the project after pushing changes via the CLI, please refer to the project's status page for more information, or you can run via the CLI:

```
rill project status
```

Likewise, if using the UI by selecting the `Deploy` button, Rill will detect the changes in files and update your deployed project accordingly. Along with the above CLI command, you can view the status of the objects on the Status page.

:::tip Interested in using Gitlab?

Check out our documentation on deploying a [Rill project using GitLab](deploy-from-cli)!

:::


## Change your production branch

By default, Rill deploys from the [default branch](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/about-branches#about-the-default-branch) of your Git repository. You can change this to any branch you want.

To deploy your project from a different branch, run the following command:

```bash
rill connect-github --prod-branch [PROD-BRANCH]
```



## Deploy from a monorepo

If your Rill project is in a subdirectory of a Git repository, use the `--subpath` option when creating your project:
```
rill connect-github --subpath path/to/rill/project
```
:::warning
Note that you must run `rill connect-github` from the <u>root</u> of your Git repository, **not** the root of your Rill project.
:::


<!-- 
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

In this case, the project will be deployed to Rill Cloud without a GitHub repository connected. You can always [add a repository via the UI](#syncing-your-github-repository) at a later time. -->