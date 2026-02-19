---
id: iframe-api
title: Embed Iframe API
description: Control embedded Rill dashboards programmatically.
sidebar_label: Embed Iframe API
sidebar_position: 11
---

# Embed Iframe API

Control embedded Rill dashboards from the parent window using [`postMessage`](https://developer.mozilla.org/en-US/docs/Web/API/Window/postMessage) with a JSON-RPC 2.0-like protocol. Communication is bidirectional: the parent sends **requests**, and the iframe sends **responses** and **notifications**.

## Setup

```html
<iframe id="rill" src="<your rill embed url>" width="100%" height="100%"></iframe>
```

```js
const iframe = document.getElementById("rill");

// Helper to send requests and await responses
function sendRequest(method, params) {
  const id = crypto.randomUUID();
  return new Promise((resolve, reject) => {
    function handler(event) {
      if (event.data?.id === id) {
        window.removeEventListener("message", handler);
        event.data.result !== undefined
          ? resolve(event.data.result)
          : reject(event.data.error);
      }
    }
    window.addEventListener("message", handler);
    iframe.contentWindow.postMessage({ id, method, params }, "*");
  });
}
```

## Methods

Methods are called **from the parent** and handled **by the iframe**. Include an `id` to receive a response; omit it for fire-and-forget.

### State

| Method | Params | Response | Description |
|---|---|---|---|
| `setState` | `string` — query params without leading `?` | `true` | Set dashboard state. See [URL Parameters](/reference/url-syntax/url-parameters) for available params. |
| `getState` | — | `{ state: string }` | Get current dashboard state as a query string |

```js
await sendRequest("setState", "view=pivot&tr=PT24H&grain=hour");
const { state } = await sendRequest("getState");
```

### Theme

| Method | Params | Response | Description |
|---|---|---|---|
| `setThemeMode` | `"light"`, `"dark"`, or `"system"` | `true` | Set light/dark mode |
| `getThemeMode` | — | `{ themeMode: string }` | Get current mode |
| `setTheme` | Theme name (`string`) or `null` for default | `true` | Apply a named theme resource |
| `getTheme` | — | `{ theme: string }` | Get current theme name (`"default"` if none) |

```js
await sendRequest("setThemeMode", "dark");
await sendRequest("setTheme", "my-custom-theme");
```

:::note
The theme name must match an existing theme resource in your Rill project. An invalid name won't error but won't apply.
:::

### AI Pane

| Method | Params | Response | Description |
|---|---|---|---|
| `setAiPane` | `boolean` | `true` | Show (`true`) or hide (`false`) the AI chat pane |
| `getAiPane` | — | `{ open: boolean }` | Get current AI pane visibility |

Only available on Explore dashboards with AI chat enabled.

## Notifications

Notifications are sent **from the iframe** to the parent. They have no `id`.

| Notification | Params | Description |
|---|---|---|
| `ready` | — | Iframe initialized and ready for requests |
| `stateChanged` | `{ state: string }` | Dashboard state changed |
| `navigation` | `{ from: string, to: string }` | User navigated between dashboards. Values are dashboard names or `"dashboardListing"`. Only fires when navigation is enabled. |
| `resized` | `{ width: number, height: number }` | Iframe content resized |
| `aiPaneChanged` | `{ open: boolean }` | AI pane opened or closed |

```js
window.addEventListener("message", (event) => {
  const { method, params } = event.data;

  if (method === "ready") console.log("Ready");
  if (method === "stateChanged") console.log("State:", params.state);
  if (method === "navigation") console.log(params.from, "→", params.to);
  if (method === "resized") console.log(params.width, "x", params.height);
  if (method === "aiPaneChanged") console.log("AI pane:", params.open);
});
```

## Error Handling

Errors follow JSON-RPC 2.0 structure:

```json
{ "id": 1, "error": { "code": -32601, "message": "Method not found" } }
```

| Code | Message | Description |
|---|---|---|
| -32600 | Invalid Request | Malformed request |
| -32601 | Method Not Found | Unknown method |
| -32602 | Invalid Params | Parameters incorrect |
| -32603 | Internal Error | Unexpected failure |
| -32700 | Parse Error | Malformed JSON |

## Full Example

```js
const iframe = document.getElementById("rill");

function sendRequest(method, params) {
  const id = crypto.randomUUID();
  return new Promise((resolve, reject) => {
    function handler(event) {
      if (event.data?.id === id) {
        window.removeEventListener("message", handler);
        event.data.result !== undefined
          ? resolve(event.data.result)
          : reject(event.data.error);
      }
    }
    window.addEventListener("message", handler);
    iframe.contentWindow.postMessage({ id, method, params }, "*");
  });
}

window.addEventListener("message", async (event) => {
  if (event.data?.method === "ready") {
    // Set dashboard state
    await sendRequest("setState", "view=pivot&tr=PT24H&grain=hour");

    // Configure theme
    await sendRequest("setThemeMode", "dark");
    await sendRequest("setTheme", "my-custom-theme");

    // Open AI chat
    await sendRequest("setAiPane", true);
  }

  if (event.data?.method === "stateChanged") {
    console.log("State:", event.data.params.state);
  }
});
```

## See Also

- [Embed Dashboards](/developers/embed/dashboards) — Setup, authentication, and iframe URL generation
- [URL Parameters](/reference/url-syntax/url-parameters) — Full reference for dashboard state parameters
