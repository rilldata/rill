---
title: "Embed Dashboards"
description: Embed Dashboards in your own applications
sidebar_label: "Embed Dashboards"
sidebar_position: 10
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';

## Embedding dashboards

### Introduction
You can embed Rill dashboards in your own applications using iframe. This allows you to use Rill dashboards as a component.

To embed a dashboard, the customer's backend needs to use a Rill service token to request an authenticated iframe URL from `admin.rilldata.com`. 
The iframe URL is then used to embed the dashboard in the customer's application. Here is a diagram of the flow:

```mermaid
sequenceDiagram
  participant A as ui.ezcommerce.com
  participant B as api.ezcommerce.com
  participant C as admin.rilldata.com
  participant D as ui.rilldata.com/-/embed
  participant E as node-4.us-east1.runtime.rilldata.com
  A ->> B: Get iframe URL
  B ->> B: Resolve the user's email <br />using ezcommerce's own auth
  B ->>+ C: Get iframe URL for:<br />project="ezcommerce"<br />user="john@example.com"<br/>(uses Rill service token)
  Note right of C: 1. Lookup deployment<br/>2. Generate JWT<br />3. Build iframe URL
  C ->>- B: iframe URL
  B ->> A: iframe URL
  A ->> D: Render iframe URL
  loop 
    D ->> E: Request data <br />(gets host and JWT from<br /> iframe URL query)
    E ->> D: Data with policies applied for<br />user="john@example.com"
  end
```

### Create a service token
Use rill cli to create a service token for your organization. You can create a new service account along with token for your organization using the following command:

```bash

rill service create <service_name> [--org "My Organization"]
```

:::caution

The service account provides admin-level access to the org and should be handled confidentially and NOT integrated in a frontend / other user-facing code.

:::

See the reference docs for more details on managing [service account and token](../reference/cli/service).

### Backend: Build an iframe URL
The backend needs to use the service token to request an iframe URL from `admin.rilldata.com`. The iframe URL is then used to embed the dashboard in the customer's application.
Here are examples of how to get an iframe URL using different languages:

<Tabs>
  <TabItem value="curl" label="Curl" default>

```bash
curl -X POST --location 'https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/iframe' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <rill-svc-token>' \
--data-raw '{
"resource": "<dashboard-name>",
"user_email":"<user-email>"
}'
```
  </TabItem>
  <TabItem value="js" label="Javascript">
    JavaScript (Node.js) with Express.js

```js
const express = require('express');
const fetch = require('node-fetch');
const app = express();
app.use(express.json());

app.post('/api/rill/iframe', async (req, res) => {
    const dashboardName = req.body.resource;
    try {
        const response = await fetch('https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/iframe', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer <rill-svc-token>`,
            },
            body: JSON.stringify({
                resource: dashboardName,
                user_email: '<user-email>',
            }),
        });
        const data = await response.json();
        res.json({ iframeResp: data.resp.body });
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
});

const PORT = 3000;
app.listen(PORT, () => {
    console.log(`Server running on port ${PORT}`);
});
```

  </TabItem>
  <TabItem value="python" label="Python">
    Python with Flask

```python
from flask import Flask, request, jsonify
import requests

app = Flask(__name__)

@app.route('/api/rill/iframe', methods=['POST'])
def get_rill_iframe():
    dashboard_name = request.json.get('resource')
    try:
        response = requests.post(
            'https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/iframe',
            headers={
                'Content-Type': 'application/json',
                'Authorization': 'Bearer <rill-svc-token>',
            },
            json={
                'resource': dashboard_name,
                'user_email': '<user-email>',
            }
        )
        response.raise_for_status()
        data = response.json()
        return jsonify(iframeResp=data['resp']['body'])
    except requests.RequestException as e:
        return jsonify(error=str(e)), 500

if __name__ == '__main__':
    app.run(port=3000)

```
  </TabItem>
  <TabItem value="go" label="Go">
    Go with net/http

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getRillIframe(w http.ResponseWriter, r *http.Request) {
	var reqBody map[string]string
	json.NewDecoder(r.Body).Decode(&reqBody)
	dashboardName := reqBody["resource"]

	requestBody, err := json.Marshal(map[string]string{
		"resource":   dashboardName,
		"user_email": "<user-email>",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := http.Post("https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/iframe", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var respBody map[string]map[string]map[string]string
	json.Unmarshal(body, &respBody)
	iFrameResp := respBody["resp"]["body"]

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(iFrameResp)
}

func main() {
	http.HandleFunc("api/rill/iframe", getRillIframe)
	fmt.Println("Server started at port 3000")
	http.ListenAndServe(":3000", nil)
}
```
  </TabItem>

  <TabItem value="java" label="Java">
    Java With Spring Boot 

```java
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestTemplate;
import org.springframework.http.ResponseEntity;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import java.util.Map;

@RestController
public class DashboardController {

    @PostMapping("/api/rill/iframe")
    public ResponseEntity<?> getRillIframe(@RequestBody Map<String, Object> payload) {
        String dashboardName = (String) payload.get("resource");
        HttpHeaders headers = new HttpHeaders();
        headers.set("Content-Type", "application/json");
        headers.set("Authorization", "Bearer <rill-svc-token>");

        Map<String, Object> request = Map.of(
                "resource", dashboardName,
                "user_email", "<user-email>"
        );

        HttpEntity<Map<String, Object>> entity = new HttpEntity<>(request, headers);
        RestTemplate restTemplate = new RestTemplate();
        ResponseEntity<Map> response = restTemplate.postForEntity("https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/iframe", entity, Map.class);

        Map<String, Object> resp = (Map<String, Object>) response.getBody().get("resp");
        Map<String, String> responseBody = Map.of("iframeResp", (String) ((Map<String, Object>) resp.get("body")));

        return ResponseEntity.ok(responseBody);
    }
}
```
  </TabItem>

</Tabs>



The API accepts the following parameters:

| Parameter | Description                                                                                                                                                                                    | Required                         |
| --- |------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------------|
| resource | The name of the dashboard to embed                                                                                                                                                             | Yes                              |
| user_email | The email of the user to embed the dashboard for                                                                                                                                               | No (either this or `attributes`) |
| attributes | Json payload to be put in the access token, used to pass attributes to the dashboard for enforcing policies. When using this make sure to pass all the attributes used in your security policy | No (either this or `user_email`) |
| ttl_seconds | The time to live for the iframe URL                                                                                                                                                            | No (Default: 86400)              |

The response will contain an `iframeSrc` that can be used to embed the dashboard in the customer's application along with `ttlSeconds` for which the iframe url will be valid after which it needs to be fetched again. Here's an example response:

```json
{
  "iframeSrc": "https://ui.rilldata.com/-/embed?access_token=<token>&instance_id=<id>&kind=MetricsView&resource=<dashboard-name>&runtime_host=<runtime_host>&state=&theme=",
  "runtimeHost": "<runtime_host>",
  "instanceId": "<id>",
  "accessToken": "<token>",
  "ttlSeconds": 86400
}
```

### Frontend: Embed the dashboard
The frontend just need to use the iframeSrc to embed the dashboard in the customer's application. Here's an example of how to embed the dashboard using iframe:

```html
<iframe title="rill-dashboard" src="<iframeSrc>" width="100%" height="1000" style="border:0;" />
```

### Example
Here's an example of how to create a dashboard component in a React application, this component will fetch the iframe URL from the backend and embed the dashboard in the frontend:

```jsx
import React, { useEffect, useState } from 'react';

const RillDashboard = () => {
  const [iframeUrl, setIframeUrl] = useState('');

  useEffect(() => {
      const getIframeUrl = async () => {
          try {
              // call the backend to get the iframe URL, see the "Backend: Build an iframe URL" section for details and examples
              const response = await fetch(`/api/rill/iframe`, {
                  method: 'POST',
                  headers: {
                      'Content-Type': 'application/json',
                  },
                  body: JSON.stringify({
                      resource: dashboardName,
                  }),
              });
              if (!response.ok) {
                  throw new Error('Network response was not ok');
              }
              const { iframeSrc } = await response.json().iframeSrc;
              setIframeUrl(iframeSrc);
          } catch (err) {
              console.error(err);
          }
      };
    getIframeUrl();
  }, []);

  return (
    <iframe title="rill-dashboard"
      src={iframeUrl}
      width="100%"
      height="1000" 
      style="border:0;"
    />
  );
};

export default Dashboard;
```
