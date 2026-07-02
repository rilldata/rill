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
connector.webhook.signing_secret=whsec_MfKQ9r8GKYqrTwjUPD8ILPZIo2LaLaSw
```

Secrets prefixed with `whsec_` are treated as base64-encoded per the specification; other
values are used as raw key bytes. If no secret is configured, deliveries are sent unsigned —
recommended only for capability URLs (e.g. Zapier or n8n hooks) where the URL itself is the
secret.

## Static headers

If a receiver requires additional headers (e.g. an `Authorization` header for an API
gateway), configure them on the connector:

```shell
connector.webhook.headers.Authorization=Bearer <token>
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
- Failed deliveries are retried up to 3 times with exponential backoff (5xx responses,
  429s and network errors are retried; other 4xx responses are not).
- All URLs are always attempted, even if an earlier one fails. Delivery failures surface as
  the alert execution's error in its history.
