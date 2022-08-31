---
title: "Tableau"
slug: "tableau"
excerpt: "Integrating Rill with Tableau"
hidden: false
createdAt: "2020-09-29T23:44:04.538Z"
updatedAt: "2021-09-21T18:51:45.545Z"
---
# Setup instructions

## Install the Avatica JDBC driver

1. Download the Avatica Jar from [https://cdn.rilldata.com/avatica/avatica-1.17.0.jar](https://cdn.rilldata.com/avatica/avatica-1.17.0.jar)

2. Place the jar in Tableau's Drivers folder. 
[block:parameters]
{
  "data": {
    "0-0": "Windows",
    "0-1": "C:\\Program Files\\Tableau\\Drivers.",
    "1-0": "OSX",
    "1-1": "~/Library/Tableau/Drivers",
    "h-0": "OS",
    "h-1": "Path"
  },
  "cols": 2,
  "rows": 2
}
[/block]
## Create credentials that allow Tableau to connect to Rill
1. To access your Rill Druid database from Tableau, you will need to use either an [API Password](doc:api-password) or a [Service Account](doc:service-accounts). 

   If using an API password, when you connect to Rill from Tableau, you will provide your Rill username as the username and your API password as the password. If using a service account, you will provide the service account as your username and the service account password as your password.

## Launch Tableau (version 2018.3 or later)
1. Launch Tableau and choose Other databases (JDBC). If you have never used this before, you'll find it under More->Other databases (JDBC).
2. Fill in the Other Databases (JDBC) modal as follows: 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/9a83bff-tableau.png",
        "tableau.png",
        1387,
        645,
        "#fbfbfb"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
  * **URL:** Copy the Tableau JDBC URL from the Integrations page in RCC and paste it into the URL field of the Other Databases (JDBC) modal (*ws1.public used in the example below*) 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/69659a3-Screen_Shot_2021-07-01_at_11.14.51_AM.png",
        "Screen Shot 2021-07-01 at 11.14.51 AM.png",
        1364,
        210,
        "#f5f5f6"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
  * **Dialect:** SQL92
  * **Username/Password:** Enter either your Rill username and your API password or your service account and service account password, as described above.  
  * **Properties:** Leave the properties field blank.

 3. Click Sign In and you'll be taken to a connection pane in Tableau.
 4. Choose druid from the Database dropdown. After this a Schema dropdown should appear.
 5. Choose druid from the Schema dropdown. After a short pause your tables will appear in the pane below and you can use them as you would use any table in Tableau.
[block:callout]
{
  "type": "warning",
  "body": "If you are using a column that has been created using one of Druid's cardinality aggregators (HyperUnique, HyperLogLog, etc), that column is a dimension. In order to get the \"measure\" value, you will need to convert it to a measure in Tableau. Using it without converting to a measure first will currently cause a Tableau exception.",
  "title": "Adjustment for Cardinality Aggregators"
}
[/block]