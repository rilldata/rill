---
id: embed-iframe-api
title: Embed Iframe API
description: Communicate with your embedded dashboards.
sidebar_label: Embed Iframe API
sidebar_position: 11

---

# Embed Iframe API

When embedding Rill inside of an `iframe` you can communicate with it using the [`postMessage`](https://developer.mozilla.org/en-US/docs/Web/API/Window/postMessage) API via a JSON-RPC 2.0-like protocol.


## Overview

The iframe exposes an API that enables external control and monitoring of its internal state. Communication is bidirectional and supports both **requests** and **notifications** using `window.postMessage`.

The state of a dashboard in Rill can be found in the URL as you are browsing it, the URL is fully human readable and will reflect whatever you are looking at on the screen.


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
Note: if including an `id` the server will respond, if you do not need a response you can omit the `id` property.

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

| Code    | Message           | Description          |
|---------|-------------------|----------------------|
| -32600  | Invalid Request    | Malformed request    |
| -32601  | Method Not Found   | Unknown method       |
| -32602  | Invalid Params     | Parameters incorrect |
| -32603  | Internal Error     | Unexpected failure   |
| -32700  | Parse Error        | Malformed JSON       |

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
  }

  if (event.data?.method === "stateChanged") {
    console.log("State changed:", event.data.params.state);
  }
});
```