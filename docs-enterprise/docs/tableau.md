---
title: "Tableau"
slug: "tableau"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Integrating Rill with Tableau"/>

## Setup instructions

### Install the Avatica JDBC driver

1. Download the Avatica Jar from [https://cdn.rilldata.com/avatica/avatica-1.17.0.jar](https://cdn.rilldata.com/avatica/avatica-1.17.0.jar)

2. Place the jar in Tableau's Drivers folder. 

| OS | Path |
|---|---|
| Windows | C:\\Program Files\\Tableau\\Drivers. |
| OSX | ~/Library/Tableau/Drivers |

### Create credentials that allow Tableau to connect to Rill
1. To access your Rill Druid database from Tableau, you will need to use either an [API Password](/api-password) or a [Service Account](/service-accounts). 

   If using an API password, when you connect to Rill from Tableau, you will provide your Rill username as the username and your API password as the password. If using a service account, you will provide the service account as your username and the service account password as your password.

### Launch Tableau (version 2018.3 or later)
1. Launch Tableau and choose Other databases (JDBC). If you have never used this before, you'll find it under More->Other databases (JDBC).
2. Fill in the Other Databases (JDBC) modal as follows: 
![](https://images.contentful.com/ve6smfzbifwz/HfMHwXwK8cSkWuViGyCKk/5a849c728cd54b5cbcf23ee2b2c8691d/9a83bff-tableau.png)
  * **URL:** Copy the Tableau JDBC URL from the Integrations page in RCC and paste it into the URL field of the Other Databases (JDBC) modal (*ws1.public used in the example below*) 
![](https://images.contentful.com/ve6smfzbifwz/7z6ezdn9IGyP2jFINgsSFY/aebfaf1560042acd9410129c6105b9de/69659a3-Screen_Shot_2021-07-01_at_11.14.51_AM.png)
  * **Dialect:** SQL92
  * **Username/Password:** Enter either your Rill username and your API password or your service account and service account password, as described above.  
  * **Properties:** Leave the properties field blank.
 3. Click Sign In and you'll be taken to a connection pane in Tableau.
 4. Choose druid from the Database dropdown. After this a Schema dropdown should appear.
 5. Choose druid from the Schema dropdown. After a short pause your tables will appear in the pane below and you can use them as you would use any table in Tableau.

:::caution Adjustment for Cardinality Aggregators
If you are using a column that has been created using one of Druid's cardinality aggregators (HyperUnique, HyperLogLog, etc), that column is a dimension. In order to get the "measure" value, you will need to convert it to a measure in Tableau. Using it without converting to a measure first will currently cause a Tableau exception.
:::
