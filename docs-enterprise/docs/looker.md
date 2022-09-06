---
title: "Looker"
slug: "looker"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Integrating Rill with Looker"/>

## Setup instructions

### Create credentials that allow Looker to connect to Rill
1. To access your Rill Druid database from Looker, you will need to use either an [API Password](/api-password)  or a [Service Account](/service-accounts). 

    If using an API password, when you connect to Rill from Looker, you will provide your Rill username as the username and your API password as the password. If using a service account, you will provide the service account as your username and the service account password as your password.

### Add a connection for your Rill workspace
1. Go to Looker and choose Admin -> Connections. 
2. Fill in the Connections modal as follows:
![](https://images.contentful.com/ve6smfzbifwz/QxfGwygyMxlmW8wIPXiui/12f28efef722cce06e33156cfeacf7ee/3d5bfd8-looker.png)
  * **Name:** Enter a descriptive name for your connection
  * **Dialect:** Select Apache Druid 0.18+
  * **Remote Host Port: ** Add your Remote Host connection string from the Integrations page in RCC. Then type the number 443.
![](https://images.contentful.com/ve6smfzbifwz/2V9LlWyR7QoY3I2Ziky1KW/db8b256512b9148800131641e897df9a/c519ce5-Screen_Shot_2021-07-01_at_11.17.01_AM.png)
  * **Database:** Enter druid
  * **Enter Username/Password:** Enter either your Rill username and your API password or your service account and service account password, as described above.
  * **SSL:** keep this box checked
  * **SQL Runner Precache:** keep this box checked
  * **Fetch information Schema For SQL Writing: **keep this box checked

 3. Click Test These Settings. If the connection works, you will see the message "Can connect" appear below the button.
 4. Click Save Connection.

### Create a Looker Project using your Rill Druid connection
1. Go to Looker and choose Develop -> Manage LookML Projets
2. Click New LookML Project
3. Name your project
4. In the Connection field, choose the connection you created above
5. For Build Views From, leave All Tables selected to see all of the tables in your workspace 
6. Click Create Project

### Explore your Druid data
1. Click on the Looker Explore menu and you should see the project you created below, with one menu item for each of your Druid tables.
2. Click on one of tables
3. You should now see the fields from your table in the Looker Explore menu. Click on dimensions and measures to add them to your Looker data pane and then click Run to execute the query and see the results. See Looker documentation for more info on how to use Looker.