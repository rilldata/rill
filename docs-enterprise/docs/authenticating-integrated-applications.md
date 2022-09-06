---
title: "🔓 Authenticate & Connect"
slug: "authenticating-integrated-applications"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Set up Service Accounts and Generic JDBC/API access"/>

## Overview
When you query Druid via an application, that application must provide authentication credentials to Druid. There are two ways to do this: a service account or an API password.

## Service Account
A Service Account is an account that your admin will create and use to authenticate API calls made to Druid by another application. If multiple users in your workspace will be using this application to access Druid, your admin can create a single service account and embed the credentials in the application that the users are running. 

As a user of that application, once your Admin creates the service account and provides the credentials to the application, you should be able to interact with Druid through that application without additional Druid credentials.

Instructions for creating a service account can be found [here](/service-accounts).

## API Password
If you want to use an application that makes API calls to Druid and no service account has been created, you may create an API password and authenticate the API call using your Rill login and the API password.

Instructions for creating an API Password can be found [here](/api-password).