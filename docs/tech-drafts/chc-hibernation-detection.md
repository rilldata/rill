# Tech Draft: ClickHouse Cloud Hibernation Detection

## Problem

ClickHouse Cloud (CHC) services with `can_scale_to_zero: true` may enter an idle/hibernated state after a period of inactivity. When this happens, queries take 30-60s to respond as the cluster wakes up. Users currently have no visibility into whether their CHC cluster is hibernated, leading to confusion when dashboards are slow to load.

## Goal

Display an informational banner in Rill Cloud when the connected CHC service is idle/hibernated. This is purely informational — no auto-wake or intervention, just awareness.

## Design

### 1. Connector YAML Configuration

Users opt in by providing CHC Cloud API credentials in their connector YAML:

```yaml
# connectors/clickhouse.yaml
type: connector
driver: clickhouse
host: "abc123.us-east-1.aws.clickhouse.cloud"
port: 8443
username: "default"
password: "{{ .env.clickhouse_password }}"
can_scale_to_zero: true

# CHC Cloud API credentials for hibernation detection
cloud_api_key_id: "{{ .env.chc_api_key_id }}"
cloud_api_key_secret: "{{ .env.chc_api_key_secret }}"
cloud_org_id: "{{ .env.chc_org_id }}"
cloud_service_id: "{{ .env.chc_service_id }}"
```

Hibernation detection is enabled when all four `cloud_*` fields are set AND `can_scale_to_zero: true`.

### 2. ClickHouse Cloud REST API

**Endpoint:** `GET https://api.clickhouse.cloud/v1/organizations/{org_id}/services/{service_id}`

**Auth:** HTTP Basic Auth with `cloud_api_key_id` as username and `cloud_api_key_secret` as password.

**Response (relevant fields):**
```json
{
  "result": {
    "id": "service-uuid",
    "name": "my-service",
    "state": "idle"
  }
}
```

**`result.state` values:**
| State | Meaning |
|-------|---------|
| `idle` | Hibernated; queries will trigger a wake-up (30-60s) |
| `awaking` | Cluster is waking up |
| `running` | Cluster is active |
| `stopped` | Manually stopped by user |
| `stopping` | In the process of stopping |
| `starting` | In the process of starting |

We care about: `idle` and `awaking` (show banner), `running` (hide banner), `stopped`/`stopping` (show different warning).

### 3. Backend Changes

#### 3a. New config fields in `configProperties` (`runtime/drivers/clickhouse/clickhouse.go`)

```go
// CHC Cloud API credentials for hibernation status checks.
// All four fields must be set (along with CanScaleToZero) to enable status polling.
CloudAPIKeyID     string `mapstructure:"cloud_api_key_id"`
CloudAPIKeySecret string `mapstructure:"cloud_api_key_secret"`
CloudOrgID        string `mapstructure:"cloud_org_id"`
CloudServiceID    string `mapstructure:"cloud_service_id"`
```

Add a helper method:

```go
func (c *configProperties) cloudStatusEnabled() bool {
    return c.CanScaleToZero &&
        c.CloudAPIKeyID != "" &&
        c.CloudAPIKeySecret != "" &&
        c.CloudOrgID != "" &&
        c.CloudServiceID != ""
}
```

#### 3b. Proto changes (`proto/rill/runtime/v1/resources.proto`)

Extend `ConnectorState` to carry the cluster status:

```protobuf
message ConnectorState {
  string spec_hash = 1;
  // Status of the ClickHouse Cloud service (empty for non-CHC connectors).
  // Values: "idle", "awaking", "running", "stopped", etc.
  string cloud_service_status = 2;
}
```

Run `make proto.generate` after this change.

#### 3c. Connector reconciler changes (`runtime/reconcilers/connector.go`)

The connector reconciler currently runs `testConnector` (Ping) on every reconcile. For CHC connectors with cloud status enabled, we add a parallel CHC API check that does NOT wake the cluster.

**Option A — Inline in reconciler (simpler):**

After `testConnector`, if the connector is CHC with cloud status enabled, call the CHC REST API and update `ConnectorState.CloudServiceStatus`. The reconciler already re-runs periodically when the spec hash hasn't changed (line 83-86), so this naturally polls.

**Option B — Background goroutine (more responsive):**

Launch a background goroutine from the ClickHouse driver's `Open` that polls the CHC API on a timer (e.g., every 60-120s) and updates the connector state. This decouples status polling from reconciliation cadence.

**Recommendation: Option A** for simplicity. The reconciler already retests connectors on each cycle. We can set `ReconcileOn` to schedule re-reconciliation at a fixed interval (e.g., 60s) when cloud status is enabled, ensuring regular polling.

**Implementation sketch for Option A:**

```go
func (r *ConnectorReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
    // ... existing code ...

    // After testConnector, check CHC cloud status
    cloudStatus := r.checkCloudStatus(ctx, self.Meta.Name.Name)
    if cloudStatus != "" {
        t.State.CloudServiceStatus = cloudStatus
    }

    // Schedule re-reconciliation for continued polling
    if cloudStatus != "" {
        return runtime.ReconcileResult{
            Retrigger: time.Now().Add(60 * time.Second),
        }
    }

    // ... existing code ...
}

func (r *ConnectorReconciler) checkCloudStatus(ctx context.Context, connectorName string) string {
    // 1. Get connector config
    // 2. Parse configProperties
    // 3. If !cloudStatusEnabled(), return ""
    // 4. Call CHC REST API
    // 5. Return result.state
}
```

**CHC API client (new file `runtime/drivers/clickhouse/cloud_api.go`):**

```go
type CloudServiceStatus struct {
    State string `json:"state"`
}

type CloudServiceResponse struct {
    Result CloudServiceStatus `json:"result"`
}

func checkCloudServiceStatus(ctx context.Context, cfg *configProperties) (string, error) {
    url := fmt.Sprintf("https://api.clickhouse.cloud/v1/organizations/%s/services/%s",
        cfg.CloudOrgID, cfg.CloudServiceID)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return "", err
    }
    req.SetBasicAuth(cfg.CloudAPIKeyID, cfg.CloudAPIKeySecret)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result CloudServiceResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }
    return result.Result.State, nil
}
```

#### 3d. Expose status via existing APIs

The `ConnectorState.cloud_service_status` field is automatically available through `ListResources` and `GetResource` RPCs — no new endpoint needed. The frontend reads it from the connector resource's state.

### 4. Frontend Changes

#### 4a. Banner constant (`web-common/src/components/banner/constants.ts`)

```ts
export const CHCHibernationBannerID = "chc-hibernation";
export const CHCHibernationBannerPriority = 2.5; // Between billing (1) and token (2) — or after token
```

#### 4b. CHC status selector (`web-admin/src/features/projects/status/selectors.ts`)

Add a query that finds the CHC connector resource and extracts its cloud status:

```ts
export function useClickHouseCloudStatus(client: RuntimeClient) {
  return createRuntimeServiceListResources(
    client,
    { kind: ResourceKind.Connector },
    {
      query: {
        select: (data: V1ListResourcesResponse) => {
          const chcConnector = data.resources?.find(
            (r) => r.connector?.spec?.driver === "clickhouse" &&
                   r.connector?.state?.cloudServiceStatus
          );
          return chcConnector?.connector?.state?.cloudServiceStatus ?? null;
        },
        refetchInterval: 30_000, // Poll every 30s to stay current
      },
    },
  );
}
```

#### 4c. Banner manager component

New component: `web-admin/src/features/projects/chc-hibernation/CHCHibernationBanner.svelte`

Uses the existing `eventBus.emit("add-banner", ...)` / `eventBus.emit("remove-banner", ...)` pattern (same as `BillingBannerManager`). When `cloudServiceStatus` is `"idle"` or `"awaking"`, emits a warning banner:

- **idle:** "Your ClickHouse Cloud service is hibernated. Initial queries may take 30-60 seconds while the cluster wakes up."
- **awaking:** "Your ClickHouse Cloud service is waking up. Queries should be responsive shortly."
- **stopped:** "Your ClickHouse Cloud service is stopped. Dashboards will not load until the service is started."

Uses `iconType: "sleep"` for idle (moon icon already exists in Banner.svelte), `"loading"` for awaking, `"alert"` for stopped.

#### 4d. Mount the banner manager

Add `CHCHibernationBanner` to the project layout, similar to how `BillingBannerManager` is mounted. It should be rendered at the project level (inside `[organization]/[project]` layout) since it's project-specific.

### 5. Connector YAML Schema

Update `runtime/parser/schema/project.schema.yaml` to add the four new CHC fields under the clickhouse connector properties, with descriptions.

### 6. Documentation

Update `docs/docs/reference/project-files/connectors.md` to document the new fields under the ClickHouse section.

## Scope and Milestones

### Phase 1 (MVP)
- Config fields in connector YAML
- Proto changes to `ConnectorState`
- CHC API client in ClickHouse driver
- Connector reconciler polling (Option A)
- Frontend banner for idle/awaking/stopped states

### Phase 2 (Future)
- Auto-detect CHC without requiring explicit `cloud_service_id` (use CHC list-services API with host matching)
- Wake-up CTA button in banner (send a lightweight query to trigger wake)
- Status indicator on the project status page resource table (e.g., a sleep icon on the connector row)
- CLI `rill project status` output includes CHC cluster state

## Open Questions

1. **Polling interval:** 60s in the reconciler seems reasonable. The frontend polls every 30s via `refetchInterval`. Is this too aggressive for the CHC API? Their rate limits should be checked.
2. **Credential templating:** The `cloud_api_*` fields use `{{ .env.* }}` templates like other secrets. Are there security concerns with storing CHC API keys in the same env as DB credentials?
3. **Multiple CHC connectors:** If a project has multiple ClickHouse connectors pointing to different CHC services, the banner should either show the worst status or show one banner per connector. Phase 1 can pick the first found.
4. **Embed surface:** Should the hibernation banner also appear in embedded dashboards (iframe)? Probably yes, since the end user will also experience the slowness.

## Files to Change

| File | Change |
|------|--------|
| `runtime/drivers/clickhouse/clickhouse.go` | Add 4 config fields, `cloudStatusEnabled()` helper |
| `runtime/drivers/clickhouse/cloud_api.go` | New file: CHC REST API client |
| `runtime/reconcilers/connector.go` | Call CHC API, set `CloudServiceStatus`, schedule re-reconcile |
| `proto/rill/runtime/v1/resources.proto` | Add `cloud_service_status` to `ConnectorState` |
| `runtime/parser/schema/project.schema.yaml` | Add CHC fields to ClickHouse schema |
| `web-common/src/components/banner/constants.ts` | Add `CHCHibernationBannerID` |
| `web-admin/src/features/projects/status/selectors.ts` | Add `useClickHouseCloudStatus` query |
| `web-admin/src/features/projects/chc-hibernation/CHCHibernationBanner.svelte` | New: banner manager component |
| `web-admin/src/routes/[organization]/[project]/+layout.svelte` | Mount `CHCHibernationBanner` |
| `docs/docs/reference/project-files/connectors.md` | Document new fields |
