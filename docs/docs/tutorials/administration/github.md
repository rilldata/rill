---
title: "GitHub"
description: Connect to GitHub [note on status page]
sidebar_label: "GitHub"
---


## Connect to GitHub
After deploying to Rill Cloud, you will have the ability to sync your project to a GitHub repository. 

### via UI

<img src = '/img/tutorials/203/github-ui.gif' class='rounded-gif' />
<br />

Starting from `v0.48`, we have added a new feature to deploy your project directly from the UI. By selecting the `Connect to GitHub` button you will then need to follow the steps in the UI to create a new blank repository [or a select an existing one] and connect this to your current Rill project. Rill will then push the current contents [or overwrite]. You'll be direct back to the Status page and can see that your project is now <a href ='https://docs.rilldata.com/deploy/existing-project/' target="BLANK" >synced with your GitHub repository! </a>

Note that you will not get all the same alerts and warning that you do via the CLI, such as source credentials check.

:::tip
While not shown in the GIF above, you may want to take a moment to set up the[ GitHub repository connection to your local folder](https://docs.rilldata.com/deploy/existing-project/github-101).
:::
---
### via CLI

```bash
rill project connect-github --help
Deploy project to Rill Cloud by pulling project files from a git repository

Usage:
  rill project connect-github [flags]

Flags:
      --path string           Path to project repository (default: current directory) (default ".")
      --subpath string        Relative path to project in the repository (for monorepos)
      --remote string         Remote name (default: first Git remote)
      --org string            Org to deploy project in (default "Rill_Learn")
      --name string           Project name (default: Git repo name)
      --description string    Project description
      --public                Make dashboards publicly accessible
      --provisioner string    Project provisioner
      --prod-version string   Rill version (default: the latest release version) (default "latest")
      --prod-branch string    Git branch to deploy from (default: the default Git branch)
```

Navigating back to Terminal, we can run the following:
```
rill project connect-github
No git remote was found.
? Do you want to create a repo? Yes
? Select a Github account for the new repository royendo

Request submitted for creating repository. Checking completion status

Successfully created repository on "https://github.com/royendo/my-rill-tutorial-cli"

Pushing local project to Github

Successfully pushed your local project to Github

Using org "Rill_Learn".

Rill project names are derived from your Github repository name.
Created project "Rill_Learn/my-rill-tutorial-cli". Use `rill project rename` to change name if required.

Rill projects deploy continuously when you push changes to Github.
Your project can be accessed at: https://ui.rilldata.com/Rill_Learn/my-rill-tutorial-cli
Opening project in browser...
```

The CLI will ask if you want to create a repo, select Yes. 

If this is your first time, you will be prompted to log into Github. Once completed, you'll see this in the browser and can navigate back to the CLI.

![img](/img/tutorials/203/git_okay.png)

Rill will automatically use the project name as the repository and project name. If there are any issues with overlapping names, it will prompt your for a different name. Once this all completes, your browser will be automatically opened.


![img](/img/deploy/existing-project/cli-upload.png)

That's it! You have connected your GitHub repository to your Rill project. Now navigating back to the Status page, you can see the repository listed. Now you can push any changes that you've made locally to the Git repository and Rill will automatically update. For more information on how to use GitHub with Rill, please refer to the <a href= 'http://localhost:4004/deploy/existing-project/github-101#pushing-changes' target ="blank" > GitHub Basics docs</a>!


## Making Changes to the GitHub Repository
While this is an unusual step and most changes and development to your project should be done via branches on your Repository, there might be times you may need to completely change the repository.

### Via the UI

By selecting the edit button (blue pencil) next to the GitHub Repository, the following UI will open where you can change the repository. Note that this will **push** the contents of your Rill project to the repository. You can change the branch of the repository under Advanced options.

![img](/img/tutorials/admin/edit-github.png)

If you want to **pull** the contents of your new repository to Rill Cloud, you will need to do the following:

1. Pull contents of GitHub Repository locally.
2. (Optional) If you want to keep the same project name, you will need to change the current project name by running `rill project rename --project <project_name>`.
    - WARNING: Renaming a project will invalidate dashboard URLs, including public URLs
3. Deploy the new folder by running `rill start` and select `Deploy` from the dashboard
4. After confirming that the new project works, delete old project by running `rill project delete <project_name>`


### Via the CLI

You can redeploy a project by re-running the `rill project connect-github` with the new repository. Note that if you want to keep the same name of the project, you will need to rename the project before doing so.

**Modifying the Branch**
```bash
rill project edit  
? Select project my-rill-tutorial
? Enter the description 
? Enter the production branch main
? Make project public No
```

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />