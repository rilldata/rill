---
title: Webhook Notifications
description: Send alert and report notifications to any HTTP endpoint
sidebar_label: Webhook
sidebar_position: 41
---

## Overview

The webhook connector sends alert and report notifications as JSON payloads to any HTTP(S)
endpoint. Use it to integrate Rill with incident tooling (PagerDuty, Opsgenie), automation
platforms (Zapier, n8n, Make) or your own services.

## Sending notifications to webhooks

Add a `webhook` block to the `notify` section of an alert or report:

```yaml
notify:
  webhook:
    urls:
      - https://example.com/rill-hook
```

Every URL receives the same payload. Duplicate URLs are only delivered to once.

## Payload

The payload is a versioned JSON envelope. There are two event types, `alert.status` and
`report.scheduled`:

```json
{
  "id": "d5f8a1e2-6a5b-4c3d-9e8f-0a1b2c3d4e5f",
  "type": "alert.status",
  "version": 1,
  "timestamp": "2026-07-02T15:04:05Z",
  "data": {
    "display_name": "Sales dropped",
    "execution_time": "2026-07-02T15:00:00Z",
    "status": "FAIL",
    "is_recover": false,
    "fail_row": { "region": "south", "sales": 0 },
    "open_link": "https://ui.rilldata.com/...",
    "edit_link": "https://ui.rilldata.com/..."
  }
}
```

For `report.scheduled` events, `data` contains `display_name`, `report_time`,
`download_format`, `summary`, `open_link` and `download_link`.

The `id` is unique per delivery and can be used by the receiver for deduplication.
The `version` field is incremented on backwards-incompatible payload changes.

## Signing

If a signing secret is configured, every delivery is signed following the
[Standard Webhooks](https://www.standardwebhooks.com/) specification: the
`webhook-id`, `webhook-timestamp` and `webhook-signature` headers are set, where the
signature is a base64 HMAC-SHA256 over `{id}.{timestamp}.{body}`. Receivers can verify
signatures with any of the [Standard Webhooks libraries](https://github.com/standard-webhooks/standard-webhooks).

Set the secret as a connector variable in your project's `.env` file (or via
`rill env set`):

```shell
connector.webhook.signing_secret=whsec_<base64-encoded-secret>
```

Secrets prefixed with `whsec_` are treated as base64-encoded per the specification; other
values are used as raw key bytes. If no secret is configured, deliveries are sent unsigned —
recommended only for capability URLs (e.g. Zapier or n8n hooks) where the URL itself is the
secret.

## Restricting destinations

Anyone who can create an alert or report chooses its webhook URLs. To limit deliveries to
known receivers, list them in `allowed_url_prefixes` on a connector YAML file:

```yaml
# connectors/webhook.yaml
type: connector
driver: webhook
allowed_url_prefixes:
  - https://hooks.example.com/rill/
```

A URL is allowed if it has the same scheme and host (including the port) as an entry and
its path starts with the entry's path. Deliveries to other URLs fail and are reported in the
execution's error, while the allowed URLs are still delivered to.

Without `allowed_url_prefixes`, deliveries to loopback, private and link-local addresses
(such as `localhost`, `10.0.0.0/8` or `169.254.169.254`) are blocked, and deliveries don't go
through an HTTP proxy configured in the environment (`HTTP_PROXY`, `HTTPS_PROXY`). To deliver
to a receiver on such an address, for example a local receiver during development, or through
a proxy, list the receivers explicitly:

```yaml
allowed_url_prefixes:
  - http://localhost:8080/
```

## Static headers

If a receiver requires additional headers (e.g. an `Authorization` header for an API
gateway), configure them on a connector YAML file (connector variables cannot express
nested maps, so headers cannot be set via `.env`). Because headers usually carry
credentials, they require `allowed_url_prefixes`, so they are only sent to the listed
receivers:

```yaml
# connectors/webhook.yaml
type: connector
driver: webhook
headers:
  Authorization: "Bearer {{ .env.connector.webhook.gateway_token }}"
allowed_url_prefixes:
  - https://gateway.example.com/
```

## Per-receiver configuration

To use different secrets or headers per receiver, declare named connector instances with
`driver: webhook` and reference them from the `notify` block:

```yaml
# connectors/my_hook.yaml
type: connector
driver: webhook
```

```yaml
# alerts/my_alert.yaml
notify:
  webhook:
    connector: my_hook
    urls:
      - https://example.com/rill-hook
```

```shell
# .env
connector.my_hook.signing_secret=whsec_...
```

## Delivery semantics

- Deliveries are `POST` requests with `Content-Type: application/json`.
- Any 2xx response counts as delivered; the response body is ignored.
- Each URL gets up to 3 attempts, waiting 1 and then 2 seconds between them. 5xx responses
  (except 501), 429s and network errors are retried; other 4xx responses are not. A
  `Retry-After` header is honored, capped at 4 seconds.
- Redirects are not followed: a 3xx response counts as a failed delivery.
- URLs are delivered to in parallel, up to 8 at a time, and always attempted, even if another
  one fails. A notification gives up after 60 seconds in total. Delivery failures surface as
  the alert execution's error in its history.
