# Tech Draft: Org-Level Admin Console — Multi-Project Status View

**Author:** royendo
**Date:** 2026-03-24
**Status:** Draft

## Problem

Org admins currently have no way to see the health and status of all their projects in one place. The org landing page (`/[organization]`) shows only project names and public/private badges. To check if a project is errored, hibernating, or has reconciliation failures, an admin must click into each project individually and navigate to its status page.

This is a significant gap for organizations managing many projects, making it impossible to quickly identify and triage issues across the org.

### Current state

**Org landing page displays per project:**
- Project name
- Admin/Viewer role badge
- Public/private icon

**Org landing page does NOT display:**
- Deployment status (RUNNING, ERRORED, STOPPED, PENDING, etc.)
- Deployment error messages
- Resource error counts or reconciliation failures
- OLAP/controller health
- Last sync time

**Existing per-project status page** (`/[organization]/[project]/-/status`) shows all of this, but only for one project at a time.

### Root cause: no batch status API

The status data exists in the system but is fragmented across services:

| Data | API | Service | Scope |
|------|-----|---------|-------|
| Project list | `ListProjectsForOrganization` | Admin | Org-wide, returns `Project` only |
| Deployment status | `GetProject` | Admin | Per-project, returns `Deployment` |
| Instance health | `InstanceHealth` | Runtime | Per-deployment, different host |
| Resource status | `ListResources` | Runtime | Per-deployment, different host |

To get a full picture today requires **2N+1 API calls** (1 list + N `GetProject` + N `InstanceHealth`), with the runtime calls going to potentially different hosts per project.

Notably, `ProjectCard.svelte` already calls `GetProject` per card (N+1), but only uses it for the public/private field — not deployment status.

---

## Option A: Frontend-Only Approach (Quick Win)

### Summary

Leverage existing APIs from the frontend. Use TanStack Query's parallel query capabilities to fetch deployment status for all projects client-side, then display a status table on a new org admin page.

### API calls (no backend changes)

1. `ListProjectsForOrganization(org, { pageSize: 1000 })` — already called on org page
2. For each project: `GetProject(org, project)` — returns `Deployment` with `status` and `status_message`
3. Optionally, for running projects: `InstanceHealth(instanceId)` via runtime proxy — returns error counts

### Frontend changes

**New route:** `web-admin/src/routes/[organization]/-/admin/+page.svelte`

Add an "Admin" tab to `OrganizationTabs` (gated on `manageOrg` permission) linking to this page.

**Page contents:** A table with columns:
- Project name (link to project)
- Deployment status (with colored dot, reuse `getStatusDotClass`/`getStatusLabel` from `display-utils.ts`)
- Status message (truncated, shown on hover for errored projects)
- Last updated (`deployment.updatedOn`)
- Public/private

**Implementation pattern:**
```svelte
// Fetch all projects
$: projects = createAdminServiceListProjectsForOrganization(organization, { pageSize: 1000 });

// For each project, fetch deployment status via GetProject
$: projectQueries = $projects.data?.projects?.map((p) =>
  createAdminServiceGetProject(organization, p.name)
) ?? [];
```

### Tradeoffs

| Pro | Con |
|-----|-----|
| No backend changes; ships in days | N+1 API calls from browser |
| Uses existing, tested APIs | Won't scale past ~50 projects |
| Good enough for small/mid orgs | No resource-level health data without runtime calls |
| Can iterate quickly on UX | Each `GetProject` also mints a JWT (unnecessary overhead) |

### Estimated scope
- 1 new route + page component
- 1 tab addition to `OrganizationTabs`
- Reuse of `display-utils.ts` helpers from project status feature

---

## Option B: New Backend API (Proper Solution)

### Summary

Add a new `ListProjectsWithStatus` RPC to the admin service that returns project metadata joined with deployment status in a single query. This eliminates the N+1 problem entirely for deployment status, and optionally supports server-side fan-out for instance health data.

### Proto changes

Add to `proto/rill/admin/v1/api.proto`:

```protobuf
// ListProjectsWithStatus lists all projects in an organization with their
// primary deployment status included. Designed for org admin dashboards.
rpc ListProjectsWithStatus(ListProjectsWithStatusRequest) returns (ListProjectsWithStatusResponse) {
  option (google.api.http) = {get: "/v1/orgs/{org}/projects-with-status"};
}

message ListProjectsWithStatusRequest {
  string org = 1 [(validate.rules).string.min_len = 1];
  uint32 page_size = 2 [(validate.rules).uint32 = {ignore_empty: true, lte: 1000}];
  string page_token = 3;
  // When true, include instance health data (requires server-side fan-out to runtimes).
  // Adds latency; omit for fast initial page loads.
  bool include_health = 4;
}

message ListProjectsWithStatusResponse {
  repeated ProjectWithStatus projects = 1;
  string next_page_token = 2;
}

message ProjectWithStatus {
  Project project = 1;
  // Primary deployment info (nil if no deployment exists, i.e. hibernating).
  DeploymentStatus deployment_status = 2;
  string deployment_status_message = 3;
  google.protobuf.Timestamp deployment_updated_on = 4;
  // Instance health (only populated when include_health=true in request).
  InstanceHealthSummary health = 5;
}

// Lightweight health summary for org-level display.
// Avoids exposing full InstanceHealth (which includes per-metrics-view errors).
message InstanceHealthSummary {
  bool healthy = 1; // true when no errors
  int32 parse_error_count = 2;
  int32 reconcile_error_count = 3;
  string controller_error = 4;
  string olap_error = 5;
}
```

### Backend implementation

**Database layer** (`admin/database/database.go`):

Add a new query method:

```go
// FindProjectsWithDeploymentStatusForOrganization returns projects joined with
// their primary deployment's status. This avoids N+1 queries for org admin views.
FindProjectsWithDeploymentStatusForOrganization(ctx context.Context, orgID, afterProjectName string, limit int) ([]*ProjectWithDeploymentStatus, error)
```

**SQL** (in `admin/database/postgres/postgres.go`):

```sql
SELECT
  p.*,
  d.status AS deployment_status,
  d.status_message AS deployment_status_message,
  d.updated_on AS deployment_updated_on
FROM projects p
LEFT JOIN deployments d ON d.id = p.primary_deployment_id
WHERE p.org_id = $1 AND lower(p.name) > lower($2)
ORDER BY lower(p.name)
LIMIT $3
```

This is a single query — `primary_deployment_id` on the `projects` table already points to the prod deployment, so the join is a simple FK lookup.

**Server handler** (`admin/server/projects.go`):

```go
func (s *Server) ListProjectsWithStatus(ctx context.Context, req *adminv1.ListProjectsWithStatusRequest) (*adminv1.ListProjectsWithStatusResponse, error) {
    // 1. Auth check: require ManageProjects permission (org admin)
    // 2. Query DB with joined deployment status
    // 3. If include_health: fan out InstanceHealth calls via runtime clients (with timeout + concurrency limit)
    // 4. Return results
}
```

For the `include_health` fan-out:
- Use `errgroup` with concurrency limit (e.g., 10 concurrent requests)
- Set a per-request timeout (e.g., 2s) so a single slow runtime doesn't block the response
- Projects with no deployment or unreachable runtimes get `health = nil`
- Cache health results briefly (30-60s) in-memory to handle page refreshes

### Frontend changes

Same as Option A (new route, new tab), but with a single API call:

```svelte
$: projectsWithStatus = createAdminServiceListProjectsWithStatus(organization, {
  pageSize: 100,
  includeHealth: true,
});
```

**Table columns:**
- Project name (link)
- Deployment status (colored dot + label)
- Health indicator (green/yellow/red based on error counts)
- Parse errors count
- Reconcile errors count
- Status message (for errored deployments)
- Last updated

**Drill-down:** Clicking a project row navigates to the existing per-project status page for full details.

### Tradeoffs

| Pro | Con |
|-----|-----|
| Single API call; scales to any org size | Requires proto + Go + frontend work |
| Deployment status is a simple SQL join (fast) | Health fan-out adds complexity |
| Server-side health fan-out is more efficient | Need to handle partial failures gracefully |
| Foundation for future org admin features | Requires `make proto.generate` + client regen |
| Health data cached server-side benefits all clients | |

### Estimated scope
- Proto: 1 RPC, 3 messages
- DB: 1 new query method + SQL
- Server: 1 handler (~100 lines), health fan-out (~50 lines)
- Frontend: 1 route + page component, 1 tab addition
- Generated code: `make proto.generate`, Orval regen

---

## Recommendation

**Ship Option A first, then build Option B.**

Option A can be built in a day or two and immediately unblocks org admins who need visibility. It works fine for orgs with < 50 projects (which covers most current customers). The N+1 pattern is already present in `ProjectCard.svelte`, so this doesn't introduce a new anti-pattern — it just makes better use of data that's already being fetched.

Option B should follow as the proper solution. The core win — the SQL join for deployment status — is straightforward and eliminates the main bottleneck. The health fan-out can be added incrementally behind the `include_health` flag.

### Migration path
1. Ship Option A behind the existing `manageOrg` permission gate
2. Build the `ListProjectsWithStatus` RPC (without health fan-out initially)
3. Swap the frontend to use the new API
4. Add `include_health` support
5. Remove the N+1 `GetProject` calls

---

## Key files reference

### Backend
- `proto/rill/admin/v1/api.proto` — Admin API definitions; `Project` (L3213), `Deployment` (L3254), `DeploymentStatus` enum (L3242)
- `admin/server/projects.go` — `ListProjectsForOrganization` handler (L47)
- `admin/database/database.go` — `Project` struct (L450), `Deployment` struct (L602), DB interface (L81)
- `admin/database/postgres/postgres.go` — SQL queries for project listing (L340)
- `proto/rill/runtime/v1/api.proto` — `InstanceHealth` message (L423)
- `runtime/server/health.go` — Health check implementation

### Frontend
- `web-admin/src/routes/[organization]/+page.svelte` — Current org landing page
- `web-admin/src/features/projects/ProjectCard.svelte` — Current project card (N+1 `GetProject` calls)
- `web-admin/src/features/projects/status/overview/` — Per-project status components (reusable patterns)
- `web-admin/src/features/projects/status/display-utils.ts` — `getStatusDotClass`, `getStatusLabel` helpers
- `web-admin/src/features/projects/status/selectors.ts` — Project status query selectors
