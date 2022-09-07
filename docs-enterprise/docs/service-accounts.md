---
title: "Service Accounts"
slug: "service-accounts"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt />

## Overview 
A Service Account is an account that your admin will create and use to authenticate API calls made to Druid by another application. If multiple users in your workspace will be using this application to access Druid, your admin can create a single service account and embed the credentials in the application that the users are running. As a user of that application, once your Admin creates the service account and provides the credentials to the application, you will be able to interact with Druid through that application without additional Druid credentials.

#Creating a Service Account
Only a user with Admin privilege can create a service account. 

To create a service account:
  - Select the Workspace to create the Service Account within (*Public in the example below*)
  - Select `Users`
  - Select `Add Service Account`
    ![](https://images.contentful.com/ve6smfzbifwz/4EGgTfCsJgKDwQlxNzsdfr/a31c91da011337878483b2a015c8a20b/ade0381-Screen_Shot_2021-06-17_at_9.46.45_AM.png)
  - Enter a name and password. Make sure the Service Account remains an Admin user. 


:::caution Fully Qualified Account Name Required
When providing the service account to an application, you must use the fully qualified string, i.e. `my-service-account`@`org`.rilldata.com.
:::

