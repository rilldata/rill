---
title: "Adding Users"
slug: "adding-users"
excerpt: "How to manage access to Explore"
hidden: false
createdAt: "2021-08-11T19:34:58.909Z"
updatedAt: "2022-08-22T19:16:27.752Z"
---
[block:api-header]
{
  "title": "Adding Users (for Admins)"
}
[/block]
Select "My Profile" on the top right to access User settings to [enter the Admin page](https://enterprise.rilldata.com/docs/explore-admin). 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/448f6f6-Admin_Access.png",
        "Admin Access.png",
        1318,
        241,
        "#d9e6eb"
      ]
    }
  ]
}
[/block]
On that page, Admins will find an option to add/edit users.
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/470125c-users.png",
        "users.png",
        1541,
        652,
        "#f0f5f8"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
To add a new user, select **Add User** on the top right of the screen. 

  * Add an email address (if the user email is within your domain, you will have the option to add as an Admin or Member) 
  * Select the security policies (which determine the dashboards visible to the user). See [Assigning dashboards to users](https://dash.readme.com/project/rill/v1.0/docs/admin-security) for adding/editing security policies

  * Send the invite (user receives link to edit password and access Rill)
[block:callout]
{
  "type": "info",
  "title": "Internal vs. Guest Users",
  "body": "By default, anyone with the email domain for your company is an Internal user. Any user with an external domain is created as a Guest. \n\nSecurity is managed via Security Policies and Guest users can never access Admin rights."
}
[/block]