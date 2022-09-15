---
title: "Superset"
slug: "superset"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt />

## Setup instructions

### Connect to your database/Rill workspace
1. Go to Superset and choose Data -> Datasets and click on the green + Database button in the top right corner. A Connect a database modal should appear.
2. Fill in the Connect a database modal as follows:
  * DISPLAY NAME: Enter a descriptive name for your connection
  * SQLALCHEMY URI: Copy the Superset SQLAlchemy URI from the Integrations page in RCC and paste it into the SQLALCHEMY URI field. 
  * Replace the username/password in the image with your username and API password. 
  ![](https://images.contentful.com/ve6smfzbifwz/3kE2qxTa3mmzQA6SjKqxIK/039d5ecab2ecfc806cbc895b030469ca/f17e194-Screen_Shot_2021-07-01_at_11.18.44_AM.png)
 3. Click Test Connection. If the connection works, you will see a "Connection Looks Good!" message appear.
 4. Click Connect to save the connection

### Create a Superset Dataset. 
1. Go to Superset and choose Data -> Datasets and click on the + Dataset button in the top right corner.
2. Fill in the Add dataset modal as follows:
 * Database: Choose the connection that you created above
 * Schema: Select druid
 * Table: Choose the table that you want to query from the dropdown menu. This dropdown menu should show the same datasets that are displayed in your RIll workspace. Each dataset represents a table.
  * Click Add
  * Now that you've created a dataset, you can click on it under Datasets and edit it (under Actions, over on the right) to configure column properties such as whether the column is a metric or dimension or whether it is your time column.  See Superset documentation on [Creating Your First Dashboard](https://superset.apache.org/docs/creating-charts-dashboards/first-dashboard) for more details.

### Create a Superset Chart
1. Go to Superset and choose Charts and then click on the + Chart button in the top right corner.
2. Choose the dataset you just created.  Alternately, when viewing the list of Datasets, you can click on the dataset name to create a chart.
3. You should now see the fields from your table in the Superset Explore menu and you can create charts as describe in the Superset documentation for [Creating Your First Dashboard](https://superset.apache.org/docs/creating-charts-dashboards/first-dashboard)