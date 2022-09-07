---
title: "Jupyter"
slug: "jupyter"
---
import Excerpt from '@site/src/components/Excerpt'

<Excerpt text="Using Rill with Jupyter"/>

## Overview
Using Jupyter with Druid is easy. Simply post an SQL query with your authentication credentials  an then parse the json output that is returned.

## Credentials
To authenticate via Jupyter, you will need to use either an [API Password](/api-password)  or a [Service Account](/service-accounts). If using an API password, when you connect you will provide your Rill username as the username and your API password as the password. If using a service account, you will provide the service account as your username and the service account password as your password.

## Example
You can post a request to Druid using `requests.post()`.

Copy the Jupyter URL from the Integrations page in RCC. 
![](https://images.contentful.com/ve6smfzbifwz/79PBlzxgWErULHTBF5DXN7/3ca1b3e4917456644d875737ac74dbe4/2b3b3b3-Screen_Shot_2021-07-01_at_11.17.47_AM.png)
In the example below, **my_database** is an example source and the host url has been in the **database_url** variable. **Username** and **API password** would be passed as well.

```python title="Python"
import requests
import pandas
import json

sql = "SELECT * from my_database limit 10"
database_url = "https://druid.ws1.public.rilldata.com/druid/v2/sql"

r = requests.post(database_url, auth=(username, password), json={"query": sql})
result = r.content.decode('utf-8') # should be #200
data = json.loads(result)
pandas.read_json(json.dumps(data))
```
