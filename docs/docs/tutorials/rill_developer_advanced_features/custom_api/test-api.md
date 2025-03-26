---
title: "Testing the Custom APIs"
description:  "Testing the API"
sidebar_label: "Testing the API"
sidebar_position: 12
---


## Testing the deployed APIs

Now that we have deployed the API to Rill Cloud, we can test the APIs with the two tokens that we made. 

### Using the service token
Let's use the following to return the results from the CLI:

```bash
curl https://admin.rilldata.com/v1/orgs/Rill_Learn/projects/my-rill-tutorial/runtime/api/SQL_api \
-H "Authorization: Bearer <your_service_token>"
```

Note that the results from `SQL_api` and `metrics_view_api` are the same. 

```bash
[{"author_name":"avogar","net_line_changes":16331},{"author_name":"Sema Checherinda","net_line_changes":8118},{"author_name":"Blargian","net_line_changes":5629},{"author_name":"Max K","net_line_changes":1904},{"author_name":"robot-clickhouse","net_line_changes":1899},{"author_name":"Raúl Marín","net_line_changes":1434},{"author_name":"János Benjamin Antal","net_line_changes":1168},{"author_name":"yariks5s","net_line_changes":1078},{"author_name":"Nikita Taranov","net_line_changes":1035},{"author_name":"Antonio Andelic","net_line_changes":1032}]%  
```

### Using a user token 

Based on the access policy, we expect that the API does not work as the user does not have access to the metrics_view, but should work with the SQL api as this is just running against a SQL table.

**Viewer's Token**

Dashboard API: (metrics_view_api)
```bash
curl https://admin.rilldata.com/v1/orgs/Rill_Learn/projects/my-rill-tutorial/runtime/api/metrics_view_api \  ...

{"error":"action not allowed"}
```

Underlying SQL table: (SQL_api)
```bash
curl https://admin.rilldata.com/v1/orgs/Rill_Learn/projects/my-rill-tutorial/runtime/api/SQL_api \ ...

[{"author_name":"avogar","net_line_changes":16331},{"author_name":"Sema Checherinda","net_line_changes":8118},{"author_name":"Blargian","net_line_changes":5629},{"author_name":"Max K","net_line_changes":1904},{"author_name":"robot-clickhouse","net_line_changes":1899},{"author_name":"Raúl Marín","net_line_changes":1434},{"author_name":"János Benjamin Antal","net_line_changes":1168},{"author_name":"yariks5s","net_line_changes":1078},{"author_name":"Nikita Taranov","net_line_changes":1035},{"author_name":"Antonio Andelic","net_line_changes":1032}]%  
```


### Developing APIs 
For further information about our custom APIs, please refer to the following [documentation](https://docs.rilldata.com/integrate/custom-api) and [references](https://docs.rilldata.com/reference/project-files/apis).