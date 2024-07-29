---
title: "Sharing Dashboards with a public URL"
description: Sharing Dashboard with a few clicks using the public link
sidebar_label: "Sharing with public URL"
sidebar_position: 36
---

## Overview

Sharing your dashboard is a key way to promote collaboration within your team and allows for quick access to your dashboards. With shared public URLS, you can create a public link to a dashboard with specific filters applied.

<img src ='/img/explore/dashboard101/public-url.gif' class='rounded-gif'/>
<br />

:::tip
Recipients are not required to log in to view the dashboard! This can be used internally and for non-users to view a dashboard.
:::

The user who receives this link will be able to use the dashboard just as any other user but wont be able to modify the applied filters. You can also set an expiration date for the view.


### How to create a public URL from the UI



After applying the filters and modifying the dashboard to your liking, please select the `share` button. 
If not already selected, select `Create public link`. 
If you want to set an expiration, please select the toggle and set the expiration.


### What the recipient sees

As expected, when opening the public URL, the user can view the dashboard. Like a logged-in user, they are able to navigate with the dashboard and do any drilling they need. Unlike a logged in user, they are not able to make any changes to the filter that you set. 


### How to manage public URLs via the CLI
 > Starting from v.0.48, share-url will be rebranded to public-url
```
rill share-url
Manage shareable URLs

Usage:
  rill share-url [command]

Available Commands:
  list        List all shareable URLs
  create      Create a shareable URL
  delete      Delete a shareable URL

Flags:
      --org string   Organization Name (default "Rill_Learning")

Global Flags:
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)

Use "rill share-url [command] --help" for more information about a command.

```
Using the Rill CLI, you can list, create or delete public URLs.




To delete a public URL, you will need an `id` parameter. In order to get the `id`, you will need to list out the public URLs. You can do so using the below with any flags to help you. 

```
rill share-url list 
```

Once you have obtained the `id` you can run the following:

```
rill share-url delete <id>
```

If you are interested in creating a public URL directly from the CLI, you can do so by passing the required parameters. (You can use the --help flag to see what additional flags are required.)