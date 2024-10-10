---
title: "Sharing Dashboards with a public URL"
description: Sharing Dashboard with a few clicks using the public link
sidebar_label: "Public Shareable URLs"
sidebar_position: 36
---

## Overview

Sharing your dashboard is a key way to promote collaboration within other users and allows for quick access to your dashboards. As an admin, you also have the ability to create public shareable URLs, meaning that you can generate an expirable public link to a dashboard with specific filters pre-applied. The receiver of the link will then be able to engage and interact with this dashboard as if they were logged in (with the parent filters locked).

<img src ='/img/explore/dashboard101/public-url.gif' class='rounded-gif'/>
<br />

:::tip
Recipients are not required to log in to view the dashboard and will not be able to change your pre-defined filters!
:::

### How to create a public URL from the UI

After applying the filters and modifying the dashboard to your liking, please select the `Share` button. 
If not already selected, select `Create public URL`. 
If you want to set an expiration, please select the toggle and set the expiration.


### What the recipient sees

As expected, when opening the public URL, the user can view the dashboard. Like a logged-in user, they are able to navigate within the dashboard and drill or slice-and-dice as needed. Unlike a logged in user, they are not able to make any changes to the filter that you've set.

:::info Did you know?

In fact, users who click on a public shareable URL cannot see the parent filters that have been applied!

:::

## How to manage public URLs

### via the UI
You can now manage public URLs via the UI. You will find a new "settings" tab in the Rill Cloud UI as an administrator.

![img](/img/explore/publicurl/public-url-settings.png)

### via the CLI
:::tip
Starting from v.0.48, `public-url` has been rebranded to `public-url`.
:::
```
rill public-url
Manage public URLs

Usage:
  rill public-url [command]

Available Commands:
  list        List all public URLs
  create      Create a public URL
  delete      Delete a public URL

Flags:
      --org string   Organization Name (default "Rill_Learning")

Global Flags:
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters (default true)

Use "rill public-url [command] --help" for more information about a command.

```
Using the Rill CLI, you can list, create or delete public URLs.

#### Deleting a public URL

To delete a public URL, you will need an `id` parameter. In order to retrieve the appropriate `id`, you will need to first list out the public URLs. You can do so using the below with any flags to help you. 

```
rill public-url list 
```

Once you have obtained the `id` you can run the following:

```
rill public-url delete <id>
```

If you are interested in creating a public URL directly from the CLI, you can do so by passing the required parameters. (You can use the --help flag to see what additional flags are required.)



