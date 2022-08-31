---
title: "Service Accounts"
slug: "service-accounts"
hidden: false
createdAt: "2020-10-29T23:33:12.214Z"
updatedAt: "2021-06-17T16:49:04.527Z"
---
#Overview 
A Service Account is an account that your admin will create and use to authenticate API calls made to Druid by another application. If multiple users in your workspace will be using this application to access Druid, your admin can create a single service account and embed the credentials in the application that the users are running. As a user of that application, once your Admin creates the service account and provides the credentials to the application, you will be able to interact with Druid through that application without additional Druid credentials.

#Creating a Service Account
Only a user with Admin privilege can create a service account. 

To create a service account:
  * Select the Workspace to create the Service Account within (*Public in the example below*)
  * Select `Users`
  * Select `Add Service Account`
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/ade0381-Screen_Shot_2021-06-17_at_9.46.45_AM.png",
        "Screen Shot 2021-06-17 at 9.46.45 AM.png",
        2706,
        588,
        "#d3d4d6"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
  * Enter a name and password. Make sure the Service Account remains an Admin user. 

  "type": "warning",
  "body": "When providing the service account to an application, you must use the fully qualified string, i.e. `my-service-account`@`org`.rilldata.com.
  "title": "Fully Qualified Account Name Required"

