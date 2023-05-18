---
title: Deploy an existing Rill project
sidebar_label: Deploy an existing project
sidebar_position: 0
---

Follow this tutorial to deploy an existing Rill project to Rill Cloud. When you deploy a project, its dashboards become available online and you can invite other people to access it.

## Push the project to Github

Rill Cloud connects to a repository on Github containing a Rill project, and continuously deploys that project on every push. Therefore, your project must be on Github before you deploy it to Rill.

Follow these steps to push your project to Github:

1. Initialize `git`:
```
git init
```
2. Add and commit the project files:
```
git add .
git commit -m "Initial commit"
```
3. Create a new Github repository on Github: [https://github.com/new](https://github.com/new)
4. Link `git` to the remote repository
```
git remote add origin https://github.com/your-account/your-repo.git
```
5. Push your repository to Github
```
git push -u origin main
```

## Deploy to Rill Cloud

With your project files on Github, you're ready to deploy the project. In the directory containing your project, run:

```
rill deploy
```

The CLI will guide you through authenticating with Rill Cloud and granting read-only access to your Rill project on Github.

## Check status

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
