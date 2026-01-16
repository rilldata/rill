# Project Status Page Redesign - Technical Design

**Date:** January 2025
**Author:** Roy
**Status:** Draft - Seeking Feedback
**PRD Reference:** Oct 2025 Project Status PRD

---

## Executive Summary

This document outlines the technical implementation plan for redesigning the Project Status page in Rill Cloud. The goal is to provide administrators with better visibility into project health, resource dependencies, and debugging capabilities—bridging the gap between CLI functionality and the Cloud UI.

---

## Table of Contents

1. [Current State Analysis](#current-state-analysis)
2. [Feature Breakdown](#feature-breakdown)
3. [Work Streams by Team](#work-streams-by-team)
4. [Phased Implementation Plan](#phased-implementation-plan)
5. [Technical Architecture](#technical-architecture)
6. [API Inventory](#api-inventory)
7. [Open Questions](#open-questions)
8. [Appendix](#appendix)

---

## Current State Analysis

### What Exists Today

| Component | Location | Status |
|-----------|----------|--------|
| Resource table with status | `web-admin/src/features/projects/status/ProjectResourcesTable.svelte` | ✅ Working |
| Deployment status display | `web-admin/src/features/projects/status/ProjectDeploymentStatus.svelte` | ✅ Working |
| Parse error display | `web-admin/src/features/projects/status/ProjectParseErrors.svelte` | ✅ Working |
| Refresh triggers (single/bulk) | `RefreshResourceConfirmDialog.svelte`, `RefreshAllSourcesAndModelsConfirmDialog.svelte` | ✅ Working |
| DAG/Graph visualization | `web-common/src/features/resource-graph/` | ✅ Exists (not on status page) |
| Smart polling with backoff | `web-admin/src/lib/refetch-interval-store.ts` | ✅ Working |

### What's Missing

| Feature | CLI Equivalent | Priority |
|---------|---------------|----------|
| DAG Viewer on status page | N/A | P0 |
| Logs viewer | `rill project logs` | P0 |
| Partition management | `rill project partitions` | P1 |
| Model detail panel | `rill project describe` | P1 |
| Token creation UI | `rill service create` | P2 |
| Project clone/download | `rill project clone` | P2 |
| Usage analytics | `rill project tables` (partial) | P2 |
| SQL/MetricsSQL console | `rill query` | P3 |
| CLI-in-browser | N/A | P3 |

---

## Feature Breakdown

### 1. DAG Viewer Integration

**Goal:** Visualize resource dependencies directly on the status page.

**Current State:**
- Full DAG implementation exists in `web-common/src/features/resource-graph/`
- Uses `@xyflow/svelte` + `@dagrejs/dagre` for layout
- Supports node caching, URL sync, bidirectional traversal
- Already integrated in resource menus (sources, models, metrics views)

**Work Required:**
- Port/integrate `ResourceGraph` component into status page
- Add toggle between list view and graph view
- Connect node clicks to resource detail panel
- Highlight error states on nodes

**Files to Modify:**
- `web-admin/src/routes/[organization]/[project]/-/status/+page.svelte`
- Create: `web-admin/src/features/projects/status/StatusDAGView.svelte`

**Estimated Complexity:** Low (reuse existing code)

---

### 2. Streaming Logs Viewer

**Goal:** View project logs in real-time without CLI.

**Current State:**
- Backend APIs ready: `WatchLogs` (streaming), `GetLogs` (historical)
- CLI implementation in `cli/cmd/project/logs.go`
- No web UI component exists

**Work Required:**
- Create `LogsViewer.svelte` component
- Implement WebSocket/gRPC-web streaming for `WatchLogs`
- Add log level filtering (DEBUG, INFO, WARN, ERROR, FATAL)
- Add search/grep functionality
- Consider HUD-style overlay that can follow user around

**API Contract:**
```typescript
// WatchLogsRequest
{
  instance_id: string;
  replay: boolean;      // replay recent logs on connect
  replay_limit: number; // max logs to replay
  level: LogLevel;      // minimum log level
}

// Log
{
  timestamp: string;
  level: LogLevel;
  message: string;
  payload: string;      // JSON additional data
}
```

**Files to Create:**
- `web-admin/src/features/projects/status/logs/LogsViewer.svelte`
- `web-admin/src/features/projects/status/logs/LogEntry.svelte`
- `web-admin/src/features/projects/status/logs/LogsFilter.svelte`

**Estimated Complexity:** Medium

---

### 3. Model Detail Panel

**Goal:** Show in-depth model information (SQL, type, parameters, execution time).

**Current State:**
- `GetResource` API returns full resource details
- CLI `rill project describe` outputs full protobuf

**Work Required:**
- Create slide-out detail panel component
- Display: model type, SQL, incremental config, partitioned status, refresh times
- Link to partition viewer for incremental models
- Show model test results (requires new API?)

**Data Available from API:**
```typescript
interface ModelState {
  connector: string;
  table: string;
  stage: ModelStateStage;
  result_connector: string;
  result_table: string;
  spec_hash: string;
  refs_hash: string;
  refreshed_on: Timestamp;
  incremental: boolean;
  partitions: number;
  partitions_pending: number;
  partitions_errored: number;
}
```

**Files to Create:**
- `web-admin/src/features/projects/status/detail-panel/ResourceDetailPanel.svelte`
- `web-admin/src/features/projects/status/detail-panel/ModelDetail.svelte`
- `web-admin/src/features/projects/status/detail-panel/SourceDetail.svelte`

**Estimated Complexity:** Medium

---

### 4. Partition Management

**Goal:** View and manage model partitions (list, filter by status, trigger refresh).

**Current State:**
- `GetModelPartitions` API ready
- `CreateTrigger` supports partition-level refresh
- CLI: `rill project partitions [--errored, --pending, --all]`

**Work Required:**
- Create partition list view within model detail panel
- Add filters: all, pending, errored
- Add bulk refresh actions for errored partitions
- Paginate for models with many partitions

**API Contract:**
```typescript
// GetModelPartitionsRequest
{
  instance_id: string;
  model: string;
  pending: boolean;
  errored: boolean;
  page_size: number;
  page_token: string;
}

// Partition
{
  key: string;
  data_json: string;  // partition key values
  watermark: Timestamp;
  executed_on: Timestamp;
  error: string;
  elapsed: Duration;
}
```

**Files to Create:**
- `web-admin/src/features/projects/status/partitions/PartitionList.svelte`
- `web-admin/src/features/projects/status/partitions/PartitionRow.svelte`

**Estimated Complexity:** Medium

---

### 5. Token Management UI

**Goal:** Create and manage service tokens with attributes.

**Current State:**
- `CreateService` and `IssueServiceAuthToken` APIs exist
- CLI: `rill service create <name> --attributes JSON`
- No web UI

**Work Required:**
- **PLAT:** Expose service creation in admin UI API (if not already)
- **APP:** Create token management UI
- Form for: service name, org/project role, JSON attributes
- Display generated token (show once)
- List existing service accounts

**API Contract:**
```typescript
// CreateServiceRequest
{
  name: string;
  org: string;
  org_role_name?: string;
  project?: string;
  project_role_name?: string;
  attributes?: string;  // JSON key-value pairs
}
```

**Estimated Complexity:** Low (UI), Low (PLAT - may already be exposed)

---

### 6. Project Clone/Download

**Goal:** Download project as ZIP without Git/CLI.

**Current State:**
- CLI clones via Git
- No direct download endpoint

**Work Required:**
- **PLAT:** Create ZIP download endpoint that packages project files
- **APP:** Add download button to status page header
- Handle projects with subpaths

**Estimated Complexity:** Low (APP), Medium (PLAT - new endpoint)

---

### 7. Usage Analytics

**Goal:** Show project usage metrics (dashboards, user activity, data usage, query times).

**Current State:**
- `OLAPListTables` provides table info
- Limited usage metrics available

**Work Required:**
- **PLAT:** Build usage aggregation APIs (query counts, response times, user sessions)
- **APP:** Create usage dashboard with charts
- Time-series visualization for trends

**Estimated Complexity:** High (PLAT - new data collection), Medium (APP)

---

## Work Streams by Team

### PLAT (Platform) Team

| Priority | Feature | Current API Status | Work Needed |
|----------|---------|-------------------|-------------|
| P0 | Resource listing | ✅ `ListResources` | None |
| P0 | Logs streaming | ✅ `WatchLogs`/`GetLogs` | None |
| P1 | Partitions | ✅ `GetModelPartitions` | None |
| P1 | Refresh triggers | ✅ `CreateTrigger` | None |
| P1 | Tables info | ✅ `OLAPListTables` | None |
| P2 | Token creation | ✅ `CreateService` | Verify admin UI exposure |
| P2 | Project download | ❌ | **New ZIP download endpoint** |
| P2 | Usage analytics | ⚠️ Partial | **New aggregation APIs** |
| P3 | Model validation | ❌ | **New test results API** |

### APP (Application) Team

| Priority | Feature | Complexity | PLAT Dependency |
|----------|---------|------------|-----------------|
| P0 | DAG Viewer on status page | Low | None |
| P0 | Enhanced resource table (sort/filter) | Low | None |
| P0 | Streaming logs viewer | Medium | None (API ready) |
| P1 | Model detail panel | Medium | None |
| P1 | Partition management UI | Medium | None |
| P2 | Token creation UI | Low | Verify API exposure |
| P2 | Project download button | Low | ZIP endpoint |
| P2 | Usage dashboard | Medium | Usage APIs |
| P3 | SQL console | High | None |
| P3 | CLI-in-browser | High | Security review |

---

## Phased Implementation Plan

### Phase 1: Foundation (APP only - No PLAT blockers)

**Goal:** Improve status page with existing infrastructure.

```
Week 1-2:
├── DAG Viewer Integration
│   ├── Add view toggle (list/graph) to status page
│   ├── Port ResourceGraph component
│   └── Connect to resource detail on click
│
├── Enhanced Resource Table
│   ├── Add column sorting (by type, status, name, last refresh)
│   ├── Add filter chips (by resource type, by status)
│   └── Improve error surfacing (inline error preview)
│
└── Model Detail Drawer
    ├── Create slide-out panel component
    ├── Display model metadata (SQL, type, connector)
    └── Show incremental/partitioned status
```

**Deliverables:**
- Status page with DAG view toggle
- Sortable/filterable resource table
- Click-through to model details

---

### Phase 2: Debugging Tools (APP + minor PLAT)

**Goal:** Enable log viewing and partition management.

```
Week 3-4:
├── Streaming Logs Viewer
│   ├── Create LogsViewer component with gRPC-web streaming
│   ├── Implement log level filtering
│   ├── Add search/grep functionality
│   └── Consider HUD overlay pattern
│
├── Partition Viewer
│   ├── Add partition tab to model detail panel
│   ├── List partitions with status indicators
│   ├── Filter by pending/errored
│   └── Add refresh action for errored partitions
│
└── Parse Error Improvements
    ├── Better error message formatting
    ├── Click-to-file navigation (if editor integration)
    └── Show error count badge on status indicator
```

**Deliverables:**
- Real-time log viewing in browser
- Partition management for incremental models
- Improved error visibility

---

### Phase 3: Advanced Features (PLAT + APP collaboration)

**Goal:** Add token management, project download, usage insights.

```
Week 5-6:
├── Token Management UI
│   ├── PLAT: Verify/expose CreateService in admin API
│   └── APP: Build token creation form with attributes editor
│
├── Project Clone/Download
│   ├── PLAT: Build ZIP download endpoint
│   └── APP: Add download button to status header
│
└── Usage Analytics (if APIs ready)
    ├── PLAT: Build query count, response time aggregation
    └── APP: Create usage dashboard with time-series charts
```

**Deliverables:**
- Token creation without CLI
- One-click project download
- Basic usage metrics (depending on PLAT progress)

---

### Phase 4: Power Features (Needs Design Review)

**Goal:** Advanced debugging and exploration tools.

```
Future:
├── SQL/MetricsSQL Console
│   ├── Embedded CodeMirror editor
│   ├── Query execution against deployed project
│   └── Result visualization
│
└── CLI-in-Browser (Needs security review)
    ├── Restricted command whitelist
    ├── Output streaming
    └── Session management
```

**Deliverables:**
- Ad-hoc query capability
- CLI commands without installation

---

## Technical Architecture

### Component Hierarchy

```
StatusPage
├── StatusHeader
│   ├── ProjectDeploymentStatus (existing)
│   ├── ViewToggle (new: list/graph)
│   ├── RefreshAllButton (existing)
│   └── DownloadProjectButton (new)
│
├── StatusContent
│   ├── [List View]
│   │   ├── FilterBar (new)
│   │   │   ├── TypeFilter
│   │   │   └── StatusFilter
│   │   └── ProjectResourcesTable (enhanced)
│   │
│   └── [Graph View]
│       └── StatusDAGView (new, wraps ResourceGraph)
│
├── ResourceDetailPanel (new, slide-out)
│   ├── ModelDetail
│   │   ├── ModelMetadata
│   │   ├── ModelSQL
│   │   └── PartitionList
│   ├── SourceDetail
│   └── MetricsViewDetail
│
├── LogsHUD (new, overlay)
│   ├── LogsFilter
│   └── LogsViewer
│
└── ParseErrorsPanel (existing, enhanced)
```

### State Management

```typescript
// URL-synced state for deep-linking
interface StatusPageState {
  view: 'list' | 'graph';
  selectedResource?: { kind: string; name: string };
  filters: {
    types: ResourceKind[];
    statuses: ReconcileStatus[];
    search: string;
  };
  logsOpen: boolean;
  logLevel: LogLevel;
}
```

### Reusable Components from web-common

| Component | Location | Use Case |
|-----------|----------|----------|
| ResourceGraph | `web-common/src/features/resource-graph/` | DAG visualization |
| VirtualizedTable | `web-common/src/components/virtualized-table/` | Large resource lists |
| SearchableFilterMenu | `web-common/src/components/searchable-filter-menu/` | Type/status filters |
| Dialog | `web-common/src/components/dialog/` | Confirmation modals |
| CodeMirror presets | `web-common/src/components/editor/presets/` | SQL display |

---

## API Inventory

### Existing APIs (Ready to Use)

| API | Service | Proto Location |
|-----|---------|----------------|
| ListResources | RuntimeService | `proto/rill/runtime/v1/api.proto` |
| GetResource | RuntimeService | `proto/rill/runtime/v1/api.proto` |
| WatchResources | RuntimeService | `proto/rill/runtime/v1/api.proto` |
| GetLogs | RuntimeService | `proto/rill/runtime/v1/api.proto` |
| WatchLogs | RuntimeService | `proto/rill/runtime/v1/api.proto` |
| GetModelPartitions | RuntimeService | `proto/rill/runtime/v1/api.proto` |
| CreateTrigger | RuntimeService | `proto/rill/runtime/v1/api.proto` |
| OLAPListTables | ConnectorService | `proto/rill/runtime/v1/connectors.proto` |
| GetProject | AdminService | `proto/rill/admin/v1/api.proto` |
| CreateService | AdminService | `proto/rill/admin/v1/api.proto` |
| IssueServiceAuthToken | AdminService | `proto/rill/admin/v1/api.proto` |

### APIs Needed (PLAT Work)

| API | Purpose | Notes |
|-----|---------|-------|
| DownloadProject | ZIP download | New endpoint, package project files |
| GetProjectUsage | Usage metrics | Query counts, response times, user sessions |
| GetModelTestResults | Test results | Model validation status |

---

## Open Questions

### For Design/Product

1. **DAG Viewer placement:** Inline toggle on status page vs dedicated route vs modal overlay?
2. **Logs viewer UX:** Full page vs HUD drawer that persists across navigation?
3. **Detail panel style:** Right drawer (like Figma) vs modal vs expand-in-place?
4. **Mobile/responsive:** How should DAG viewer behave on smaller screens?

### For PLAT

1. **ZIP download:** Should this include `.env` files? Git history? Just source files?
2. **Usage APIs:** What metrics are feasible to collect? Storage implications?
3. **Token creation:** Is `CreateService` already exposed in admin UI API or just CLI?

### For APP

1. **URL state:** Should all filter/view state be URL-synced for shareability?
2. **Logs streaming:** gRPC-web vs REST polling for WatchLogs?
3. **Cache strategy:** How to handle resource graph position cache across deployments?

### For Security

1. **CLI-in-browser:** What commands should be allowed? How to sandbox?
2. **Token display:** Show-once pattern? Copy-to-clipboard only?

---

## Appendix

### A. CLI Command Mapping

| CLI Command | Web UI Equivalent | Status |
|-------------|-------------------|--------|
| `rill project status` | Status page | ✅ Exists |
| `rill project describe <resource>` | Resource detail panel | 🟡 Planned |
| `rill project logs` | Logs viewer | 🟡 Planned |
| `rill project partitions` | Partition list | 🟡 Planned |
| `rill project refresh` | Refresh buttons | ✅ Exists |
| `rill project tables` | Usage dashboard | 🟡 Planned |
| `rill project clone` | Download button | 🟡 Planned |
| `rill service create` | Token creation UI | 🟡 Planned |

### B. File Structure (Proposed)

```
web-admin/src/features/projects/status/
├── index.ts
├── ProjectResources.svelte (existing)
├── ProjectResourcesTable.svelte (existing, enhanced)
├── ProjectDeploymentStatus.svelte (existing)
├── ProjectParseErrors.svelte (existing)
├── StatusDAGView.svelte (new)
├── StatusFilterBar.svelte (new)
├── detail-panel/
│   ├── ResourceDetailPanel.svelte
│   ├── ModelDetail.svelte
│   ├── SourceDetail.svelte
│   ├── MetricsViewDetail.svelte
│   └── PartitionList.svelte
├── logs/
│   ├── LogsViewer.svelte
│   ├── LogsHUD.svelte
│   ├── LogEntry.svelte
│   └── LogsFilter.svelte
└── tokens/
    ├── TokenCreateDialog.svelte
    └── TokenList.svelte
```

### C. Design References

- Figma Mocks: https://www.figma.com/design/I79srwZXL0i6dthpE1wzMj/Roy-s-playgronud?node-id=245-6128
- DAG Viewer inspiration: Notion canvas, Vercel deployment maps
- Logs viewer inspiration: Vercel deployment logs

---

## Feedback Requested

1. Does the phasing make sense? Should anything be reordered?
2. Are there missing features from the PRD?
3. Any concerns about the proposed component architecture?
4. PLAT: Are the "APIs Needed" accurate? Any existing APIs I missed?
5. Timeline expectations for each phase?

---

*Last updated: January 2025*
