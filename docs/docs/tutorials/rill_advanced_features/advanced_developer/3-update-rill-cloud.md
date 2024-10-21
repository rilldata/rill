---
title: "How to push your changes to  Rill Cloud"
description:  Redeploy onto Rill Cloud
sidebar_label: "Pushing changes to Rill Cloud"
---

## How to push changes made on Rill Developer to Rill Cloud

If you need to make any changes in Rill Cloud, you will need to do so via Rill Developer and push the changes to Rill Cloud. There are two ways to do so:

1. Re-deploy from UI
2. Push changes via GitHub / GitHub Desktop


### Re-deploy for UI

You'll notice that the `Deploy to share` button has now changes to `Update`. When you have made changes that you'd like to push to Rill Cloud, you can select this button. We'll take care of the background tasks for you and navigate you to the project dashboard page.

<img src = '/img/tutorials/204/redeploy.gif' class='rounded-gif' />
<br />
>



### Pushing changes via GitHub via the CLI
While not going into all the details of GitHub and its commands [see our docs for more details](https://docs.rilldata.com/deploy/existing-project/github-101), we will assume that you have synced your folder to your GitHub repository.
Let's first make sure that this current folder is synced with the correct repository running the following:
```
git remote -v
```

If we try to commit a change without adding our new files we will see the following error.

```
git commit -m "Adding new files"
On branch main
Your branch is up to date with 'origin/main'.

Untracked files:
  (use "git add <file>..." to include in what will be committed)
	.DS_Store
	explore-dashboards/advanced_metrics_view_explore.yaml
  metrics/advanced_metrics_view.yaml
	models/advanced_commits___model.sql
	tmp/

nothing added to commit but untracked files present (use "git add" to track)
```

Let's try again after adding the mentioned files.

```
git add explore-dashboards/advanced_metrics_view_explore.yaml
git add metrics/advanced_metrics_view.yaml
git add models/advanced_commits___model.sql 
git commit -m "Adding new files"           
    [main 68b293e] Adding new files
    ...
    To https://github.com/royendo/my-rill-tutorial.git
    59beefe..68b293e  main -> main
    branch 'main' set up to track 'origin/main'.
```

:::tip
It's probably not best practice to push your changes directly onto your `main` branch and instead push your changes to another branch then after confirming everything, merge the branches in Github.

:::

Checking on our GitHub repository, we can see that the dashboard and model folder have been updated with the commit message, "Adding new files."

![my-rill-project](/img/tutorials/204/github-pushed-changes.png)


That's it! You can now make you changes locally and push to GitHub to sync to your Rill Cloud project whenever you need. Now that we have setup the project's contents, let's build out some feature on Rill Cloud.

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />