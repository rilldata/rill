# Open links for embedded reports and alerts

## Problem

Reports and alerts created from an embed send emails with no usable open link:

- Embedded reports are attachment-only; the reconciler drops every link (`embedCreated` in `runtime/reconcilers/report.go`).
- Embedded alerts get Rill Cloud links (`/{org}/{project}/-/alerts/{alert}/open?token=...`) plus an unsubscribe link,
  because `GetAlertMeta` only recognizes owners by user ID.
  These take the embed user out of the embedder's app and into Rill Cloud.

The person who opens one of these links should land on the embedder's page,
with the dashboard embedded and in the state the report or alert captured.

## Scope

- Applies to both reports and alerts created from an embed (identified by the `admin_owner_user_email` annotation).
- Embedded reports and alerts are only for the user who created them:
  the creator is the only email recipient, and there are no edit, download or unsubscribe links.
- Links don't carry a magic token.
  The person opening the link goes through the embedder's normal embed integration,
  which authenticates them and calls `GetIFrame` to get an embed access token.
- Not in scope: reports and alerts created in Rill Cloud, Slack and other non-email notifiers, unsubscribe.

## Flow

```
Email link
  https://app.acme.com/analytics/open?rill_open=<blob>
        │  embedder authenticates its user (own auth)
        ▼
Embedder backend
  GetIFrame{ ..., user_email, open: <blob> }
        │  admin decodes blob, scopes token, returns iframe_src
        ▼
Iframe
  /-/embed?...&open=<blob>
        │  embed layout redirects
        ▼
  /-/embed/{reports|alerts}/{name}/open?execution_time=...
        │  maps report/alert query to dashboard state
        ▼
  /-/embed/explore/{name}?<state>   or   /-/embed/canvas/{name}?<state>
```

## The `rill_open` param

One opaque query param that Rill appends to the embedder's open URL.
Embedders forward it to `GetIFrame` without reading it, so the payload can change over time.

Encoding: unpadded base64url of a JSON object:

```json
{"kind": "report", "name": "weekly_revenue", "execution_time": "2026-10-20T00:00:00Z"}
```

- `kind`: `report` or `alert`.
- `name`: name of the report or alert resource.
- `execution_time`: RFC3339 timestamp of the run, used to resolve relative time ranges; optional.

The blob isn't signed: it only identifies a resource, and the embed access token plus the runtime security rules decide access.

The format is shared by the runtime (encode) and the admin service (decode),
so it lives in one small package, e.g. `runtime/pkg/embedopen`, with `Encode`/`Decode` and a `Kind` type.

## Configuring the open URL

The open URL is a page on the embedder's site that runs their embed integration, e.g. `https://app.acme.com/analytics/open`.
It can include its own query params; Rill keeps them and adds `rill_open`.
It's resolved in this order:

1. **Per embed:** `GetIFrameRequest.open_url`.
   Saved on the embed auth token and copied into the report or alert as the `admin_embed_open_url` annotation when it's created from that embed.
   This covers multi-tenant embedders that serve tenants on different domains.
2. **Project default:** `rill.yaml`:
   ```yaml
   embeds:
     open_url: https://app.acme.com/analytics/open
   ```

If neither is set, the email has no open link (reports keep sending the attachment, as today).

## Backend changes

### Proto (`proto/rill/admin/v1/api.proto`)

`GetIFrameRequest`:

- `string open_url`: base URL for open links in emails from reports and alerts created by this embed.
  Must be an absolute `http(s)` URL.
- `string open`: the `rill_open` value from an open link.
  Can't be combined with `resource`.

Run `make proto.generate` and `make docs.generate`.

### Admin

- `admin/database/postgres/migrations/0099.sql`: add `open_url TEXT` to `embed_auth_tokens` (the table is new on this branch, so edit it in place).
  Update `InsertEmbedAuthTokenOptions` and `EmbedAuthToken` in `admin/database/database.go`, and the postgres implementation.
- `GetIFrame` (`admin/server/deployment.go`):
  - Validate `open_url` and store it on the embed auth token.
  - When `open` is set: decode it, then look up the resource with `LookupReport`/`LookupAlert` and read the target dashboard from its `explore`/`canvas` annotations (falling back to `web_open_path` for older ones).
    If navigation is disabled, set `overrideResources` to the report or alert plus its target dashboard.
    Add `open=<blob>` to `iframeQuery`.
- `yamlForManagedReport`/`yamlForManagedAlert` (`admin/server/reports.go`, `alerts.go`), embed branch:
  - Set `Annotations.AdminEmbedOpenURL` from the embed auth token's `open_url`.
    `Edit*` keeps the existing annotation unless the editing token has a new `open_url`.
  - Only allow `[ownerEmail]` as email recipients, and reject Slack targets.
- `GetReportMeta`/`GetAlertMeta`: for embed-created resources, don't issue magic tokens or build Rill Cloud URLs.
  The runtime skips calling them for these resources (see below); add a guard so callers can't get Cloud links for them by mistake.

### Runtime

- `runtime/parser/parse_rillyaml.go`: parse `embeds.open_url` into `RillYAML.EmbedOpenURL`.
  Add an `EmbedOpenURL` field to `drivers.Instance`, filled in `runtime.go` the same way as `PublicPaths`, with a registry migration.
- Reconcilers (`runtime/reconcilers/report.go`, `alert.go`): for embed-created resources:
  - Open URL: the `admin_embed_open_url` annotation, else `inst.EmbedOpenURL`.
  - Build the open link in the runtime: `<open url>?rill_open=embedopen.Encode({kind, name, execution_time})`.
    No admin call is needed for links because there's no token.
    Reports still need the admin service for `query_for` attributes; alerts already get theirs from `GetAlertMeta`.
  - Send only `openLink` (plus the attachment for reports).
    Leave out edit, download and unsubscribe links.
    This replaces the attachment-only `embedCreated` branch in `report.go` and the per-recipient URL handling for embed alerts in `alert.go`.
- Security rules (`runtime/security.go`) already allow the owner by email, so the creator's embed token can read the report or alert spec on the open page.
  Anyone else gets an access error.

## Frontend changes

- Move the logic in `web-admin/src/routes/[organization]/[project]/-/{reports/[report],alerts/[alert]}/open/+page.{ts,svelte}`
  into shared modules under `web-admin/src/features/scheduled-reports/open/` and `web-admin/src/features/alerts/open/`.
  They take the resource spec, execution time and a target URL builder:
  - Cloud: `/{org}/{project}/explore/{name}`, `/{org}/{project}/-/share/{token}/...`, `/{org}/{project}/canvas/{name}`.
  - Embed: `/-/embed/explore/{name}`, `/-/embed/canvas/{name}`.
  The cloud routes keep their current behavior.
- New routes: `web-admin/src/routes/-/embed/reports/[report]/open/` and `web-admin/src/routes/-/embed/alerts/[alert]/open/`.
  - Fetch the resource with the runtime client; the cloud routes get it from the project layout's `parent()`, which doesn't exist in embed.
  - `ignoreFilters: false`: without a magic token, no filters are baked into the token.
  - AI reports: show an error that they aren't supported in embed.
  - Canvas reports: redirect to `/-/embed/canvas/{name}` with the state from `web_open_state`, as the cloud route does.
- `web-admin/src/routes/-/embed/+layout.ts`: when `open` is in the URL, decode it and redirect to the matching open route,
  instead of `/-/embed/{type}/{resource}`.
  Decoding mirrors `embedopen.Decode`; keep it in `web-admin/src/features/embeds/`.
- Optional: an "Open" button on the embedded report and alert metadata pages that uses the same routes.

## Embedder integration

Document under the embedding docs:

1. Set `embeds.open_url` in `rill.yaml`, or pass `open_url` to `GetIFrame` on pages where users can create reports and alerts.
2. On the open page, authenticate the user, read `rill_open` from the URL and call `GetIFrame` with `open: rill_open` and the user's identity (the same `user_email` used when they created the report or alert).
3. Render the iframe as usual.

## Testing

- `runtime/pkg/embedopen`: encode/decode round trip and validation.
- Parser: `embeds.open_url` in `rill.yaml`.
- `admin/server/reports_yaml_test.go` (and alerts): annotation copied from the embed token; recipient and Slack restrictions.
- `GetIFrame`: `open` decoding, scoping when navigation is off, rejecting `open` + `resource`, `open_url` validation.
- `runtime/reconcilers/report_test.go` and alert tests: embed-created resources send only the open link built from the annotation, then from `rill.yaml`, and no link when neither is set.
- Frontend: unit tests for the shared open-URL builders and the `open` redirect in the embed layout.

## Rollout

1. Frontend: move the open logic into shared modules, and add the embed open routes and the `open` redirect.
2. Backend: `embedopen` package, `GetIFrame` `open`/`open_url`, `rill.yaml` config, annotation, reconciler link building and recipient restrictions.
3. Docs.
