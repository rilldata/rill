---
title: "How to Embed a Dashboard"
sidebar_label: "How to Embed your Rill Dashboards"
sidebar_position: 10
hide_table_of_contents: false
---

The following guide is based on the example repository: [Rill Embedding Example](https://github.com/rilldata/rill-embedding-example). We will create the same three embed dashboards in this guide. (We will not go over how to create the web application) To view our publicly available web application: [Click here!](https://rill-embedding-example.netlify.app/)

## Embedding a dashboard 

To begin, please review [the documentation](https://docs.rilldata.com/integrate/embedding) on embedding Rill Dashsboards via iframes and the various requirements after you have deployed your project to Rill Cloud.


We will continue assuming that you've:

1. Reviewed and understand how the authenticated iframe URL are generated and passed to the frontend application,
2. Created your service token via:
```bash
rill service create <service_name>
```
3. Reviewed the documentation for available parameters

### Using the demo dashboard

We will be using the [demo dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ad) that contains three dashboards. However, you will need to use your own dashboard as the token you created is specifically for your organization. Note the dashboard ID will be used, not the title.


- Spend vs Request Canvas Dashboard (spend-request-canvas-dashboard)
- Progammatic Ads Auction (auction)
- Programmtic Ads Bids (bids)

To test whether we are able to generate a iframe URL, please run the following from the CLI. Please replace the `org-name`, `project-name`, `rill-svc-token`, `dashboard-name` and your `user-email`.
```bash
curl -X POST --location 'https://admin.rilldata.com/v1/organizations/<org-name>/projects/<project-name>/iframe' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer <rill-svc-token>' \
--data-raw '{
"resource": "<dashboard-name>",
"user_email":"<user-email>"
}'
```
Your results should provide you a iframeSrc URL. Open this in your browser and see if you are able to view and interact with your dashboard. This should work exactly like the dashboard page within Rill Cloud.

```
"iframeSrc":"https://ui.rilldata.com/-/embed?access_token=...
```

Note, if you are trying to embed a dashboard that has navigation, or is a canvas dashboard, you will need to add a few components into `--data-raw`. 

**Navigation Enabled:**
```bash
--data-raw '{
"resource": "bids",
 "navigation": true,  
"user_email":"email@domain.com"
```


**Canvas Dashboard:**
```bash
--data-raw '{
"resource": "canvas",
"kind": "rill.runtime.v1.Dashboard",
"user_email":"email@domain.com"
}'
```

Once you can confirm that all the dashboards are working as designed, you can embed these into your site. However, note that there is a ttl_seconds paramter that is default to 86400 seconds, 24 hours. This will keep the iframe URL alive for only 1 day. Therefore, creating a static site would require you to manually create this URL and update the site accordingly. 

Please refer to the examples in the documents for different ways to generate the iframe URL.
https://docs.rilldata.com/integrate/embedding#backend-build-an-iframe-url


### Sample JavaScript Code: [Click me for source!](https://github.com/rilldata/rill-embedding-example/blob/main/pages/api/iframe.js)
Taken from our example repository, you can create a js file that retrieves the iframe URL. Note that the service token will be retrieved from an environmental variable, but all the other components are defined before fetching the URL.

```js
// Get the secret Rill service token from an environment variable.
const rillServiceToken = process.env.RILL_SERVICE_TOKEN;

// Configure the dashboard to request an iframe URL for.
// Note that the organization must be the same as the one the service token is associated with.
const rillOrg = "demo";
const rillProject = "rill-openrtb-prog-ads";
const rillDashboard = "bids";

// This is a serverless function that makes an authenticated request to the Rill API to get an iframe URL for a dashboard.
// The iframe URL is then returned to the client.
// Iframe URLs must be requested from the backend to prevent exposing the Rill service token to the browser.
export default async function handler(req, res) {
    try {
        const url = `https://admin.rilldata.com/v1/organizations/${rillOrg}/projects/${rillProject}/iframe`;
        const response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${rillServiceToken}`,
            },
            body: JSON.stringify({
                resource: rillDashboard,
                // You can pass additional parameters for row-level security policies here.
                // For details, see: https://docs.rilldata.com/integrate/embedding
            }),
        });
        const data = await response.json();
        if (response.ok) {
            res.json(data);
        } else {
            throw new Error(data.message);
        }
    } catch (error) {
        res.status(500).json({ error: error.message });
    }
}
```

