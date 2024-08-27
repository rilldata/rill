---
title: "Testing the Custom APIs"
description:  "Testing the API"
sidebar_label: "Testing the API"
sidebar_position: 12
---


## Testing the deployed APIs

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

**Viewer Token**

```bash
{"error":"does not have access to custom APIs"}
```

### Confirming the error here, getting same for admin filtered from access policy and viewer

This is expected behavior as we tried to use a user that does not have access to the custom APIs. 
