---
id: embed-iframe-api
title: Embed Iframe API
description: Communicate with your embedded dashboards.
sidebar_label: Embed Iframe API
sidebar_position: 11

---

# Embed Iframe API

When embedding Rill inside of an `iframe`, you can communicate with it using the [`postMessage`](https://developer.mozilla.org/en-US/docs/Web/API/Window/postMessage) API via a JSON-RPC 2.0-like protocol.


## Overview

The iframe exposes an API that enables external control and monitoring of its internal state. Communication is bidirectional and supports both **requests** and **notifications** using `window.postMessage`.

The state of a dashboard in Rill can be found in the URL as you are browsing it. The URL is fully human-readable and will reflect whatever you are looking at on the screen.


## Embedding and Initialization

Embed the iframe in your page:

```html
<iframe id="my-iframe" src="<your rill embed url>" width="600" height="400"></iframe>
```

Set up message handling and send requests from the parent window:

```js
const iframe = document.getElementById("my-iframe");

window.addEventListener("message", (event) => {
  const { id, result, error, method, params } = event.data;
  
  // notifications
  if (method === "ready") {
    console.log("Iframe is ready");
  }

  if (method === "stateChanged") {
    console.log("State changed to:", params.state);
  }

  if (method === "navigation") {
    console.log("Navigated from:", params.from, "to:", params.to);
    // params.from and params.to will be dashboard names or "dashboardListing"
  }

  if (method === "resized") {
    console.log("Iframe resized to:", params.width, "x", params.height);
  }

  // responses
  if (id && result) {
    console.log("Response to request:", result);
  }

  if (id && error) {
    console.error("RPC error:", error);
  }
});
```

## Supported Methods

These methods are called **from the parent** and handled **by the iframe**.  
Note: if including an `id`, the server will respond. If you do not need a response, you can omit the `id` property.

### `setState(state)`

Sets the current state inside the iframe.

```js
iframe.contentWindow.postMessage({
  id: 1,
  method: "setState",
  params: "view=pivot&tr=PT24H&grain=hour",
}, "*");
```

**Response:**

```json
{ "id": 1, "result": true }
```


### `getState()`

Fetches the current internal state of the iframe.

```js
iframe.contentWindow.postMessage({
  id: 2,
  method: "getState"
}, "*");
```

**Response:**

```json
{ "id": 2, "result": {"state": "<rill state string>"} }
```


### `getThemeMode()`

Fetches the current theme mode of the iframe.

```js
iframe.contentWindow.postMessage({
  id: 3,
  method: "getThemeMode"
}, "*");
```

**Response:**

```json
{ "id": 3, "result": {"themeMode": "light"} }
```

The `themeMode` value will be one of: `"light"`, `"dark"`, or `"system"`.


### `setThemeMode(themeMode)`

Sets the theme mode inside the iframe.

```js
iframe.contentWindow.postMessage({
  id: 4,
  method: "setThemeMode",
  params: "dark"
}, "*");
```

**Parameters:**
- `themeMode` (string): The theme mode to set. Must be one of: `"light"`, `"dark"`, or `"system"`.

**Response:**

```json
{ "id": 4, "result": true }
```

**Error Response (if invalid themeMode):**

```json
{
  "id": 4,
  "error": {
    "code": -32603,
    "message": "Expected themeMode to be one of \"dark\", \"light\", or \"system\""
  }
}
```

### `getTheme()`

Fetches the current theme name applied to the dashboard.

```js
iframe.contentWindow.postMessage({
  id: 5,
  method: "getTheme"
}, "*");
```

**Response:**

```json
{ "id": 5, "result": {"theme": "my-custom-theme"} }
```

If no theme is set, the response will be:

```json
{ "id": 5, "result": {"theme": "default"} }
```

The `theme` value will be the name of the theme resource, or `"default"` if no theme is set.


### `setTheme(theme)`

Sets the theme name to apply to the dashboard.

```js
iframe.contentWindow.postMessage({
  id: 6,
  method: "setTheme",
  params: "my-custom-theme"
}, "*");
```

**Parameters:**
- `theme` (string | null): The theme name to set. Must be the name of an existing theme resource. To clear the theme and use the default, pass `null` or `"default"`.

**Response:**

```json
{ "id": 6, "result": true }
```

**Error Response (if invalid theme):**

```json
{
  "id": 6,
  "error": {
    "code": -32603,
    "message": "Expected theme to be a string or null"
  }
}
```

**Note:** The theme name must correspond to an existing theme resource in your Rill project. Setting an invalid theme name will not cause an error, but the theme will not be applied.

## Notifications

Notifications are sent **from the iframe** to the parent window. These do not include an `id`.

### `ready()`

Fired once when the iframe is initialized and ready to receive messages.

```json
{ "method": "ready" }
```

### `stateChanged({ state: string })`

Fired whenever the internal state of the iframe changes.

```json
{ "method": "stateChanged", "params": { "state": "<rill state string>" } }
```

### `navigation({ from: string, to: string })`

Fired whenever a user navigates between dashboards. This event is only emitted when navigation is enabled in the embed configuration.

- `from`: The name of the dashboard the user navigated from, or `"dashboardListing"` if navigating from the dashboard listing page
- `to`: The name of the dashboard the user navigated to, or `"dashboardListing"` if navigating to the dashboard listing page

This event fires for all dashboard navigation scenarios:
- Navigating from one explore dashboard to another
- Navigating from one canvas dashboard to another
- Navigating from an explore dashboard to a canvas dashboard (or vice versa)
- Navigating from the dashboard listing page to any dashboard
- Navigating from any dashboard to the dashboard listing page

```json
{ "method": "navigation", "params": { "from": "dashboard-name", "to": "another-dashboard-name" } }
```

Example when navigating from the listing page:
```json
{ "method": "navigation", "params": { "from": "dashboardListing", "to": "dashboard-name" } }
```

Example when navigating to the listing page:
```json
{ "method": "navigation", "params": { "from": "dashboard-name", "to": "dashboardListing" } }
```

### `resized({ width: number, height: number })`

Fired whenever the iframe content is resized.

```json
{ "method": "resized", "params": { "width": 1200, "height": 800 } }
```


## Error Handling

All errors follow the JSON-RPC 2.0 structure:

```json
{
  "id": 3,
  "error": {
    "code": -32601,
    "message": "Method not found"
  }
}
```

**Common Error Codes:**

| Code   | Message          | Description          |
| ------ | ---------------- | -------------------- |
| -32600 | Invalid Request  | Malformed request    |
| -32601 | Method Not Found | Unknown method       |
| -32602 | Invalid Params   | Parameters incorrect |
| -32603 | Internal Error   | Unexpected failure   |
| -32700 | Parse Error      | Malformed JSON       |

---

## Full Example

```js
const iframe = document.getElementById("my-iframe");

function sendRequest(method, params) {
  const id = Math.random().toString(36).substr(2, 9);
  return new Promise((resolve, reject) => {
    function handler(event) {
      if (event.data?.id === id) {
        window.removeEventListener("message", handler);
        if (event.data.result !== undefined) resolve(event.data.result);
        else reject(event.data.error);
      }
    }
    window.addEventListener("message", handler);
    iframe.contentWindow.postMessage({ id, method, params }, "*");
  });
}

window.addEventListener("message", async (event) => {
  if (event.data?.method === "ready") {
    console.log("Iframe ready");

    await sendRequest("setState", "view=pivot&tr=PT24H&grain=hour");
    const currentState = await sendRequest("getState");
    console.log("Current state:", currentState);

    const currentThemeMode = await sendRequest("getThemeMode");
    console.log("Current theme mode:", currentThemeMode.themeMode);
    await sendRequest("setThemeMode", "dark");

    const currentTheme = await sendRequest("getTheme");
    console.log("Current theme:", currentTheme.theme);
    await sendRequest("setTheme", "my-custom-theme");
  }

  if (event.data?.method === "stateChanged") {
    console.log("State changed:", event.data.params.state);
  }

  if (event.data?.method === "navigation") {
    console.log("Navigated from:", event.data.params.from, "to:", event.data.params.to);
  }

  if (event.data?.method === "resized") {
    console.log("Iframe resized:", event.data.params.width, "x", event.data.params.height);
  }
});
```