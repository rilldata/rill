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

Navigating back to Terminal, we can run the following:
```
rill deploy
No git remote was found.
You can connect to Github or use one-time uploads to deploy your project.
? Do you want to use one-time uploads? No
```

We will be prompted with the same UI but this time select no. This time, you will be navigating to GitHub to install Rill Cloud to your account. Please proceed with the procedure and authorize the repository.

Once completed, you will hit the following UI and can return to the CLI.

![img](/img/tutorials/203/git_okay.png)


Since we had deployed onto Rill Cloud earlier without GitHub, when you try to deploy this time, Rill will warn you that the project name is already used. Let's create a new project with the following name `my-rill-tutorial-git`.
```
? Select a Github account for the new repository <your_account>

Request submitted for creating repository. Checking completion status

Successfully created repository on "https://github.com/royendo/my-rill-tutorial"

Pushing local project to Github

Successfully pushed your local project to Github

Using org "Rill_Learn".

Rill project names are derived from your Github repository name.
The "my-rill-tutorial" project already exists under org "Rill_Learn". Please enter a different name.
? Enter a project name my-rill-tutorial-git
Created project "Rill_Learn/my-rill-tutorial-git". Use `rill project rename` to change name if required.

Rill projects deploy continuously when you push changes to Github.

Could not access all connectors. Rill requires credentials for the following connectors:

 - gcs (used by commits__ and others)

Run `rill env configure --project my-rill-tutorial` to provide credentials.

Your project can be accessed at: https://ui.rilldata.in/Rill_Learn/my-rill-tutorial-git
Opening project in browser...
```

That's it! You have connected your GitHub repository to your Rill project. Now navigating back to the Status page, you can see the repository listed. Now you can push any changes that you've made locally to the Git repository and Rill will automatically update. For more information on how to use GitHub with Rill, please refer to the <a href= 'http://localhost:4004/deploy/existing-project/github-101#pushing-changes' target ="blank" > GitHub Basics docs</a>!


![img](/img/tutorials/203/status-git.png)

:::tip
While not shown in the steps above, you may want to take a moment to set up the[ GitHub repository connection to your local folder](https://docs.rilldata.com/deploy/existing-project/github-101). In the next course, we will start making some changes to the files and setting this up now makes it easier to push to the repository.
:::


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />