---
title: MotherDuck
description: Connect to data in your MotherDuck account
sidebar_label: MotherDuck
sidebar_position: 60
---

<!-- WARNING: There are links to this page in source code. If you move it, find and replace the links and consider adding a redirect in docusaurus.config.js. -->

## How to configure credentials in Rill

How you configure access to MotherDuck depends on whether you are developing a project locally using `rill start` or are setting up a deployment using `rill deploy`.

### Configure credentials for local development

When developing a project locally, you need to set `motherduck_token` in your environment variables. 
An example of using this syntax in terminal:
```
export motherduck_token='<token>'
```
You can also add this line to your bash profile.

Refer to MotherDuck [docs](https://motherduck.com/docs/authenticating-to-motherduck/#using-the-service-token-to-connect) for more information on authenticating with token.

### Configure credentials for deployments on Rill Cloud

Once a project having a MotherDuck source has been deployed using `rill deploy`, Rill requires you to explicitly provide the motherduck token using following command:
```
rill env configure
```
Note that you must `cd` into the Git repository that your project was deployed from before running `rill env configure`.
