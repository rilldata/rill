---
title: "Jupyter"
slug: "jupyter"
excerpt: "Jupyter"
hidden: false
createdAt: "2020-10-30T00:03:41.056Z"
updatedAt: "2021-07-07T21:28:50.136Z"
---
#Overview
Using Jupyter with Druid is easy. Simply post an SQL query with your authentication credentials  an then parse the json output that is returned.

#Credentials
To authenticate via Jupyter, you will need to use either an [API Password](doc:api-password)  or a [Service Account](doc:service-accounts). If using an API password, when you connect you will provide your Rill username as the username and your API password as the password. If using a service account, you will provide the service account as your username and the service account password as your password.

#Example
You can post a request to Druid using requests.post(). 

Copy the Jupyter URL from the Integrations page in RCC. 
[block:image]
{
  "images": [
    {
      "image": [
        "https://files.readme.io/2b3b3b3-Screen_Shot_2021-07-01_at_11.17.47_AM.png",
        "Screen Shot 2021-07-01 at 11.17.47 AM.png",
        1356,
        204,
        "#f9f8f9"
      ],
      "sizing": "80"
    }
  ]
}
[/block]
In the example below, **my_database** is an example source and the host url has been in the **database_url** variable. **Username** and **API password** would be passed as well.

[block:code]
{
  "codes": [
    {
      "code": "import requests\nimport pandas\nimport json\n\nsql = \"SELECT * from my_database limit 10\"\ndatabase_url = \"https://druid.ws1.public.rilldata.com/druid/v2/sql\"\n\nr = requests.post(database_url, auth=(username, password), json={\"query\": sql})\nresult = r.content.decode('utf-8') # should be #200\ndata = json.loads(result)\npandas.read_json(json.dumps(data))\n",
      "language": "python"
    }
  ]
}
[/block]