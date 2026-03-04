# Tech Draft: Comprehensive Telemetry Instrumentation

**Status:** Draft
**Author:** Roy Endo
**Audience:** Platform Team
**Date:** March 2026
**PRD:** Comprehensive Telemetry Instrumentation for Rill

---

## 1. Current State

### What We Have Today

Rill has **two independent telemetry systems** running side-by-side:

| System | Scope | Custom Events | Autocapture | Session Recording |
|--------|-------|---------------|-------------|-------------------|
| **Rill custom telemetry** | web-local, web-admin, CLI, runtime | Yes (~20 events) | No | No |
| **PostHog JS SDK** | web-local, web-admin | No (unused for custom events) | Yes | Yes |

#### Rill Custom Telemetry

Architecture:

```
Event Handler (BehaviourEventHandler, ErrorEventHandler, ActiveEventHandler)
  → MetricsService.dispatch(action, args)
    → EventFactory.buildEvent(commonFields, ...)
      → TelemetryClient.fireEvent(event)
```

Transport:
- **Local (web-local, CLI):** `RillIntakeClient` → `POST /local/track` → CLI proxy → `https://intake.rilldata.io`
- **Cloud (web-admin):** `RillAdminTelemetryClient` → `POST /v1/telemetry/events` → admin server → Kafka
- **CLI/runtime (Go):** `activity.Client` → `IntakeSink` (HTTP) or `KafkaSink`

Current common fields per event:
```
app_name, install_id, client_id, build_id, version, is_dev,
project_id, user_id, organization_id (cloud only),
analytics_enabled, mode, service_name
```

Current common user fields:
```
locale, browser, os, device_model
```

**Existing behavioral events** (~20 total):
- `navigate`, `deploy-intent`, `deploy-success`, `login-start`, `login-success`
- `source-success`, `source-modal`, `source-cancel`, `source-add`
- `ghconnected-start`, `ghconnected-create-repo`, `ghconnected-success`, `ghconnected-overwrite-prompt`, `ghconnected-failure`, `ghconnected-disconnect`
- `user-invite`, `user-domain-whitelist`, `example-add`, `project-empty`
- Error events: `source-error`, `error-boundary`
- Active event: `active` (heartbeat every 60s)

Go-side CLI events: `install-success`, `app-start`, `login-start`, `login-success`, `deploy-start`, `deploy-success`, `ghconnected-start`, `ghconnected-success`, `dataaccess-start`, `dataaccess-success`

#### PostHog JS SDK

Initialized in both `web-local` and `web-admin` with:
- **Autocapture enabled** (clicks, form submissions, pageviews)
- **Session recording** (masked inputs/text)
- **Heatmaps** enabled
- `posthog.identify()` called on cloud login
- Session IDs passed between local → cloud via URL params during deploy flow
- **No custom `posthog.capture()` calls exist anywhere** — PostHog is only used for its built-in autocapture/recording features

Key files:
- `web-common/src/lib/analytics/posthog.ts` — init, identify, session ID helpers
- `web-common/src/metrics/service/MetricsService.ts` — custom event dispatcher
- `web-common/src/metrics/initMetrics.ts` — local init
- `web-admin/src/features/telemetry/initCloudMetrics.ts` — cloud init
- `runtime/pkg/activity/` — Go event client, sinks, event format

### Gaps Identified

1. **No connector funnel tracking** — we track `source-add` and `source-cancel` but not `connector_selected`, `connector_form_started`, `connector_form_submitted`, or `connector_add_error`
2. **No model/SQL editor events** — zero coverage
3. **No dashboard interaction events** — no filter, time range, drill, export, share tracking
4. **No AI feature events** — only PostHog autocapture, no structured custom events
5. **No deploy funnel granularity** — just `deploy-intent` and `deploy-success`, missing org selection, naming, errors
6. **No cloud admin events** — member management, access policies, embed config are untracked
7. **No navigation/UI engagement events** — modals, tabs, tooltips, onboarding, empty state CTAs
8. **Missing global properties** — no `session_id`, `anonymous_id`, `rd_project_id` / `rc_project_id` distinction, `environment`, `platform`, `current_page`
9. **Inconsistent naming** — current events use kebab-case (`source-add`); PRD wants snake_case (`connector_add_success`)
10. **No local-to-cloud project linkage** — `project_id` on local is an MD5 of the directory name; no mapping to cloud `project_id` or `org_id` after deploy

---

## 2. Proposed Approach

### Decision: Extend the Rill Custom Telemetry Pipeline

**Recommendation:** Route all new events through the existing Rill custom telemetry system (`MetricsService` → Intake API / Kafka), not PostHog.

**Rationale:**
- We already own the full pipeline end-to-end: event factories, transport, and storage (intake API + Kafka)
- The factory/handler pattern is well-established and gives us type-safe, auditable event definitions
- Routing events through our own infrastructure avoids vendor lock-in and gives us full control over the data
- The Rill pipeline already handles both local (intake HTTP) and cloud (Kafka) transport with opt-out enforcement
- Product analytics queries (funnels, retention, paths) can be built on top of our own data warehouse

**PostHog: Future Sunset (Out of Scope)**

PostHog is currently used only for autocapture, session recording, and heatmaps — we make zero custom `posthog.capture()` calls today. Long-term, we should plan to sunset PostHog entirely:
- Session recording can be replaced by a self-hosted or alternative solution
- Autocapture is low-signal and will be superseded by the structured semantic events defined in this PRD
- Heatmaps provide marginal value once we have comprehensive click-level tracking
- Removing PostHog eliminates a third-party dependency, reduces page load overhead, and avoids sending user interaction data to an external service

**This is not part of the current effort.** For now, PostHog continues running as-is. A separate follow-up effort should evaluate the sunset timeline and replacement plan for session recording. New events defined in this doc will NOT go through PostHog.

### High-Level Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                      Frontend (Svelte)                         │
│                                                                │
│  Component code                                                │
│       │                                                        │
│       ▼                                                        │
│  track("event_name", { ...eventProps })                        │
│       │                                                        │
│       ▼                                                        │
│  ┌─────────────────────────────────────────┐                   │
│  │  Telemetry Module (new wrapper)         │                   │
│  │                                         │                   │
│  │  - Reads global props from store        │                   │
│  │  - Merges global + event props          │                   │
│  │  - Checks analytics opt-out             │                   │
│  │  - Checks no-duplicate rules            │                   │
│  │  - Dispatches via MetricsService        │                   │
│  └─────────────────────────────────────────┘                   │
│       │                                                        │
│       ▼                                                        │
│  MetricsService.dispatch()                                     │
│       │                                                        │
│       ├──► RillIntakeClient (web-local) ──► /local/track       │
│       │         ──► CLI proxy ──► intake.rilldata.io           │
│       │                                                        │
│       └──► RillAdminTelemetryClient (web-admin) ──► /v1/...    │
│                 ──► admin server ──► Kafka                     │
│                                                                │
│  (PostHog runs separately for autocapture/recording only;      │
│   no custom events routed through PostHog)                     │
└────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────┐
│                   Go (CLI / Runtime)                            │
│                                                                │
│  CLI commands (deploy, start, login, etc.)                     │
│       │                                                        │
│       ▼                                                        │
│  activity.Client.Record()  ──►  IntakeSink / KafkaSink         │
│  (unchanged; same pipeline as frontend events)                 │
└────────────────────────────────────────────────────────────────┘
```

**Scope boundary:** This tech draft covers the frontend (web-common, web-local, web-admin) instrumentation only. Go-side CLI events (e.g. `app-start`, `deploy-start`) continue through the existing `activity.Client` pipeline. All events — frontend and backend — land in the same data store (intake API → warehouse, or Kafka → warehouse).

---

## 3. Implementation Plan

### Phase 1: Foundation — Telemetry Module & Global Properties

**Goal:** Build the central `track()` function and global property store that all future events will use.

#### 3.1 Global Properties Store

Create a new Svelte writable store in `web-common` that holds the properties the PRD requires on every event.

**File:** `web-common/src/lib/analytics/telemetry-store.ts`

```typescript
import { writable, derived, get } from "svelte/store";
import { v4 as uuidv4 } from "uuid";
import { page } from "$app/stores";

interface TelemetryGlobalProps {
  session_id: string;
  user_id: string | null;
  anonymous_id: string;
  rd_project_id: string | null;
  rc_project_id: string | null;
  org_id: string | null;
  rill_version: string;
  environment: "local" | "cloud";
  platform: string;
}

const ANON_ID_KEY = "rill_anonymous_id";

function getOrCreateAnonymousId(): string {
  let id = localStorage.getItem(ANON_ID_KEY);
  if (!id) {
    id = uuidv4();
    localStorage.setItem(ANON_ID_KEY, id);
  }
  return id;
}

function detectPlatform(): string {
  const ua = navigator.userAgent.toLowerCase();
  if (ua.includes("mac")) return "macos";
  if (ua.includes("win")) return "windows";
  if (ua.includes("linux")) return "linux";
  return "unknown";
}

export const telemetryProps = writable<TelemetryGlobalProps>({
  session_id: uuidv4(),           // New UUID per app load
  user_id: null,
  anonymous_id: getOrCreateAnonymousId(),
  rd_project_id: null,
  rc_project_id: null,
  org_id: null,
  rill_version: "",
  environment: "local",
  platform: detectPlatform(),
});

// Updater functions called during app init and auth flows
export function setTelemetryUser(userId: string) { ... }
export function setTelemetryLocalProject(rdProjectId: string) { ... }
export function setTelemetryCloudProject(rcProjectId: string, orgId: string) { ... }
export function setTelemetryVersion(version: string) { ... }
export function setTelemetryEnvironment(env: "local" | "cloud") { ... }
```

**Where updaters get called:**
- `web-local/src/routes/+layout.svelte` — set `environment: "local"`, `rill_version`, `rd_project_id` from `GetMetadataResponse`
- `web-admin/src/routes/+layout.ts` — set `environment: "cloud"`, `user_id`, `org_id`, `rc_project_id` from URL params and user query
- On cloud login (both apps) — set `user_id` and call `posthog.identify()`
- On deploy success — set `rc_project_id` and `org_id` to link local→cloud

#### 3.2 Central `track()` Function

**File:** `web-common/src/lib/analytics/track.ts`

```typescript
import { get } from "svelte/store";
import { page } from "$app/stores";
import { telemetryProps } from "./telemetry-store";
import { metricsService } from "@rilldata/web-common/metrics/initMetrics";

let analyticsEnabled = true;

export function setAnalyticsEnabled(enabled: boolean) {
  analyticsEnabled = enabled;
}

export function track(
  eventName: string,
  properties: Record<string, unknown> = {},
) {
  if (!analyticsEnabled) return;
  if (!metricsService) return;

  const global = get(telemetryProps);
  const currentPage = get(page)?.url?.pathname ?? "";

  // Build the event in the standard format expected by the intake/Kafka pipeline.
  // This dispatches through MetricsService → TelemetryClient (RillIntakeClient or
  // RillAdminTelemetryClient depending on environment).
  metricsService.dispatch("trackGenericEvent", [{
    event_name: eventName,
    // Global properties
    ...global,
    current_page: currentPage,
    timestamp: new Date().toISOString(),
    // Event-specific properties
    ...properties,
  }]);
}
```

This requires adding a `trackGenericEvent` action to the `MetricsService` factory pattern (see Section 3.6 below).

**Design decisions:**
- `track()` is the single entry point for all new telemetry events. No scattered `metricsService.dispatch()` calls in component code.
- Global props are read from the store at call time, so they're always current.
- `current_page` is derived from SvelteKit's `$page` store at capture time.
- `analyticsEnabled` is set from the same `GetMetadataResponse.analyticsEnabled` / cloud config that controls the existing system.
- `timestamp` is ISO8601 as the PRD requires.
- Events flow through the same transport as existing events (intake API for local, admin API → Kafka for cloud).

#### 3.3 UI Interaction Helper

For generic `button_clicked`, `modal_opened`, etc. events, provide a helper that extracts UI properties:

**File:** `web-common/src/lib/analytics/track-ui.ts`

```typescript
import { track } from "./track";

interface UIInteractionProps {
  component?: string;
  aria_label?: string;
  element_type?: "button" | "link" | "toggle" | "dropdown" | "input";
  action?: "click" | "open" | "close" | "select" | "submit" | "dismiss";
}

export function trackUI(
  eventName: string,
  props: UIInteractionProps & Record<string, unknown> = {},
) {
  track(eventName, props);
}

// Svelte action for declarative tracking on DOM elements
export function trackClick(node: HTMLElement, params: {
  event?: string;
  component?: string;
  props?: Record<string, unknown>;
}) {
  const handler = () => {
    const ariaLabel = node.getAttribute("aria-label") ?? "";
    const elementType = node.tagName.toLowerCase() === "a" ? "link" : "button";
    trackUI(params.event ?? "button_clicked", {
      aria_label: ariaLabel,
      component: params.component ?? "",
      element_type: elementType,
      action: "click",
      ...params.props,
    });
  };

  node.addEventListener("click", handler);
  return {
    destroy() {
      node.removeEventListener("click", handler);
    },
    update(newParams: typeof params) {
      params = newParams;
    },
  };
}
```

Usage in Svelte components:
```svelte
<script>
  import { trackClick } from "@rilldata/web-common/lib/analytics/track-ui";
</script>

<!-- Declarative: Svelte action -->
<button use:trackClick={{ component: "DeployModal", props: { step: "confirm" } }}
        aria-label="Deploy project">
  Deploy
</button>

<!-- Imperative: direct call for complex flows -->
<button on:click={() => {
  trackUI("connector_form_submitted", {
    component: "ConnectorForm",
    connector_type: selectedConnector,
  });
  handleSubmit();
}}>
  Connect
</button>
```

#### 3.4 Initialization Wiring

**web-local** (`web-local/src/routes/+layout.svelte`):
```typescript
// After existing initMetrics(config):
import { setAnalyticsEnabled } from "@rilldata/web-common/lib/analytics/track";
import {
  setTelemetryVersion,
  setTelemetryEnvironment,
  setTelemetryLocalProject,
} from "@rilldata/web-common/lib/analytics/telemetry-store";

if (shouldSendAnalytics) {
  setAnalyticsEnabled(true);
  setTelemetryVersion(config.version);
  setTelemetryEnvironment("local");
  setTelemetryLocalProject(config.projectPath); // or a stable hash
} else {
  setAnalyticsEnabled(false);
}
```

**web-admin** (`web-admin/src/routes/+layout.ts`):
```typescript
// After existing initCloudMetrics():
import { setAnalyticsEnabled } from "@rilldata/web-common/lib/analytics/track";
import {
  setTelemetryVersion,
  setTelemetryEnvironment,
  setTelemetryUser,
  setTelemetryCloudProject,
} from "@rilldata/web-common/lib/analytics/telemetry-store";

setAnalyticsEnabled(true);
setTelemetryVersion(cloudVersion);
setTelemetryEnvironment("cloud");
if (user) setTelemetryUser(user.id);
// org_id and rc_project_id set in sub-layouts when URL params are available
```

#### 3.5 Identity Flow

Identity resolution is handled entirely within the Rill pipeline — no PostHog identity calls needed for custom events.

```
1. App loads → anonymous_id created/loaded from localStorage (persistent UUID)
2. Telemetry store initialized with anonymous_id, session_id (per-load UUID)
3. Cloud login → setTelemetryUser(user_id)
   → All subsequent events carry both anonymous_id and user_id
   → Backend can merge pre-auth and post-auth events via anonymous_id
4. On deploy_success → setTelemetryCloudProject(rc_project_id, org_id)
   → Links rd_project_id to rc_project_id in event stream
   → Backend can retroactively enrich earlier local events
```

The `anonymous_id` → `user_id` linkage is written as an event attribute, not a PostHog `identify()` call. The data warehouse / analytics layer joins on `anonymous_id` to stitch pre-auth and post-auth sessions.

#### 3.6 Extending MetricsService for Generic Events

The existing `MetricsService` uses a factory pattern where each event type has a dedicated factory method. To support the ~60 new events without creating a factory method for each one, add a `GenericEventFactory`:

**File:** `web-common/src/metrics/service/GenericEventFactory.ts`

```typescript
import { MetricsEventFactory } from "./MetricsEventFactory";
import type { CommonFields, MetricsEvent } from "./MetricsTypes";

export class GenericEventFactory extends MetricsEventFactory {
  public trackGenericEvent(
    commonFields: CommonFields,
    eventProps: Record<string, unknown>,
  ): MetricsEvent {
    const event = this.getBaseMetricsEvent(
      "behavioral",
      eventProps.event_name as string,
      commonFields,
      {}, // no legacy commonUserFields; global props from telemetry store
    ) as MetricsEvent;

    // Merge all event-specific properties
    Object.assign(event, eventProps);
    return event;
  }
}
```

Register it alongside the existing factories during `initMetrics()` and `initCloudMetrics()`:
```typescript
const metricsService = new MetricsService(telemetryClient, [
  new ProductHealthEventFactory(),
  new BehaviourEventFactory(),
  new ErrorEventFactory(),
  new GenericEventFactory(),   // ← new
]);
```

This approach avoids touching existing factory/handler code while providing a flexible dispatch path for all new events.

---

### Phase 2: Event Instrumentation (by Surface)

Each subsection below maps to a PRD event catalog section. Events are listed with the component/file where the `track()` call should be placed.

#### 2.1 App Lifecycle

| Event | Where to Instrument |
|---|---|
| `app_launched` | Keep as Go-side `app-start` via existing pipeline. Optionally also fire from `web-local/+layout.svelte` `onMount` for frontend session start. |
| `project_opened` | `web-local/+layout.svelte` after `GetMetadataResponse` loads; `web-admin` org/project layout `load` |
| `project_created` | Go-side `rill init` (existing pipeline); also in the splash screen "Create project" handler |
| `cloud_login` | `web-admin/+layout.ts` when user is authenticated; also fire from `posthog.identify()` |
| `cloud_logout` | Logout handler in web-admin |

#### 2.2 Connector / Source (Critical PLG Funnel)

These are the highest-priority events. Instrument in the connector/source flow components.

**Key files to instrument:**
- `web-common/src/features/connectors/` — connector selection UI
- `web-common/src/features/sources/` — source add/edit flows
- The specific connector form components for each type

| Event | Component / Location | Key Props |
|---|---|---|
| `connector_page_viewed` | Add Source page `onMount` | `page` |
| `connector_selected` | Connector tile click handler | `connector_type` |
| `connector_form_started` | First input focus/change in connector form | `connector_type` |
| `connector_form_submitted` | Form submit handler | `connector_type` |
| `connector_add_success` | Success callback after source reconcile | `connector_type` |
| `connector_add_error` | Error callback | `connector_type`, `error_code` |
| `connector_add_cancelled` | Modal dismiss / back button | `connector_type`, `step` |
| `connector_edited` | Edit connector save handler | `connector_type` |
| `connector_deleted` | Delete connector confirm handler | `connector_type` |

**Migration note:** Map existing `source-add` → `connector_add_success` and `source-cancel` → `connector_add_cancelled` by firing both the legacy and new events during a transition period.

#### 2.3 Model / SQL Editor

**Key files to instrument:**
- `web-common/src/features/models/` — model creation and editing
- `web-common/src/features/editor/` — SQL editor pane

| Event | Component / Location | Key Props |
|---|---|---|
| `model_created` | File creation handler | `model_type` |
| `model_saved` | File save handler | `model_type`, `has_error` |
| `model_run` | Run/refresh button handler | `model_type` |
| `model_run_success` | Run success callback | `row_count`, `duration_ms` |
| `model_run_error` | Run error callback | `error_code` |
| `model_deleted` | File delete confirm handler | — |
| `sql_editor_opened` | Editor pane mount | — |
| `sql_editor_run` | Run button in SQL editor | `has_error` |

#### 2.4 Metrics View & Dashboard

**Key files to instrument:**
- `web-common/src/features/dashboards/` — explore dashboard interactions
- `web-common/src/features/canvas/` — canvas dashboard interactions
- Filter, time range, comparison, measure, dimension UI components

| Event | Component / Location | Key Props |
|---|---|---|
| `metrics_view_created` | Metrics YAML creation handler | `source` |
| `metrics_view_saved` | Metrics YAML save handler | `measure_count`, `dimension_count` |
| `dashboard_viewed` | Dashboard page `onMount` | `dashboard_name` |
| `dashboard_filter_applied` | Filter add handler | `filter_type`, `dimension` |
| `dashboard_filter_cleared` | Clear all filters handler | — |
| `dashboard_time_range_changed` | Time range selector change | `time_range` |
| `dashboard_comparison_toggled` | Comparison toggle handler | `enabled` |
| `dashboard_measure_selected` | Measure selector change | `measure_name` |
| `dashboard_dimension_drilled` | Dimension value click | `dimension_name` |
| `dashboard_leaderboard_sorted` | Sort column click | `sort_column`, `sort_dir` |
| `dashboard_export_clicked` | Export button click | `export_format` |
| `dashboard_share_clicked` | Share button click | — |

**Note:** Dashboard components are shared between explore and embedded surfaces. The `environment` global prop distinguishes local vs cloud context. For embeds, consider adding an `is_embedded` property (derivable from auth context).

#### 2.5 AI Features

**Key files to instrument:**
- `web-common/src/features/ai/` or wherever AI generation flows live

| Event | Component / Location | Key Props |
|---|---|---|
| `ai_completion_requested` | AI generation trigger | `feature` |
| `ai_completion_success` | AI response received | `feature`, `duration_ms` |
| `ai_completion_error` | AI error/timeout | `feature`, `error_code` |
| `ai_result_accepted` | User accepts AI output | `feature` |
| `ai_result_rejected` | User discards/regenerates | `feature` |
| `ai_prompt_edited` | User modifies prompt before run | `feature` |

#### 2.6 Deployment (Local → Cloud)

**Key files to instrument:**
- `web-local/src/routes/(misc)/deploy/+page.svelte` — deploy entry point
- `web-common/src/features/project/deploy/` — deploy flow components

| Event | Component / Location | Key Props |
|---|---|---|
| `deploy_flow_started` | Deploy button click / CLI trigger | `trigger` |
| `deploy_org_selected` | Org selector confirm | `org_id` |
| `deploy_project_named` | Project name input | — |
| `deploy_submitted` | Deploy confirm | `org_id`, `rc_project_id` |
| `deploy_success` | Deploy success callback | `org_id`, `rc_project_id`, `rd_project_id` |
| `deploy_error` | Deploy error callback | `error_code` |
| `deploy_cancelled` | Deploy dismiss/back | `step` |
| `redeploy_triggered` | Redeploy button | `org_id`, `rc_project_id` |

**Critical:** On `deploy_success`, update the telemetry store:
```typescript
setTelemetryCloudProject(rcProjectId, orgId);
```

This ensures all subsequent local events carry the cloud context (`rc_project_id`, `org_id`).

#### 2.7 Cloud-Specific Events

Instrument in `web-admin` components:

| Event | Component / Location | Key Props |
|---|---|---|
| `cloud_project_viewed` | Project page `onMount` | `org_id`, `rc_project_id` |
| `cloud_project_created` | Create project success | `org_id` |
| `cloud_project_deleted` | Delete project confirm | `org_id`, `rc_project_id` |
| `cloud_member_invited` | Invite member submit | `org_id`, `role` |
| `cloud_member_removed` | Remove member confirm | `org_id` |
| `cloud_access_policy_changed` | Access toggle handler | `org_id`, `rc_project_id`, `access_level` |
| `cloud_embed_configured` | Embed config save | `org_id`, `rc_project_id` |
| `cloud_settings_opened` | Settings page mount | `settings_section` |

#### 2.8 Navigation & UI Engagement

These are generic interaction events for surfaces that don't have semantic events.

| Event | Implementation Approach |
|---|---|
| `button_clicked` | `use:trackClick` Svelte action on buttons without semantic events |
| `nav_item_clicked` | Instrument nav components directly |
| `modal_opened` / `modal_closed` | Instrument shared modal component(s) in web-common |
| `tab_switched` | Instrument shared tab components |
| `tooltip_viewed` | Add 1s debounce timer on tooltip hover (use Svelte action) |
| `onboarding_step_viewed` / `onboarding_step_completed` | Instrument onboarding checklist component |
| `empty_state_cta_clicked` | Instrument empty state CTA buttons |
| `error_banner_viewed` / `error_banner_dismissed` | Instrument toast/banner components |

**No-duplicate rule:** If a button already fires a semantic event (e.g. `connector_form_submitted`), do NOT also fire `button_clicked`. Implement this by having the semantic `track()` call and not adding `use:trackClick` to that element.

---

### Phase 3: Aria-Label Audit

The PRD requires `aria_label` on `button_clicked` events. For this to be useful:

1. **Audit all interactive elements** in `web-common`, `web-local`, and `web-admin` for `aria-label` presence
2. **Add missing `aria-label` attributes** — this is also an accessibility improvement
3. **Naming convention:** Use descriptive, stable labels (e.g. `aria-label="Add connector"` not `aria-label="btn-1"`)

This can be done incrementally alongside event instrumentation. Recommend running an automated lint rule (e.g. `eslint-plugin-jsx-a11y` equivalent for Svelte) to enforce going forward.

---

## 4. Event Naming Convention

**Format:** `noun_verb` or `noun_verb_qualifier`, snake_case

Examples: `connector_add_success`, `dashboard_filter_applied`, `ai_completion_requested`

**Migration:** Existing kebab-case events (`source-add`, `deploy-intent`) remain in the legacy pipeline. New PostHog events use snake_case exclusively. No renaming of legacy events.

---

## 5. Opt-Out Behavior

The existing opt-out mechanism (`analytics: false` in rill.yaml / `analytics_enabled` from GetMetadata) must be respected:

1. `setAnalyticsEnabled(false)` prevents all `track()` calls from firing
2. `MetricsService.dispatch()` already checks `analytics_enabled` and early-returns — this is a second layer of protection
3. Cloud user-level opt-out (if implemented) should call `setAnalyticsEnabled(false)` to suppress all new events
4. PostHog's autocapture/recording is separately gated by the same `analyticsEnabled` check in the layout init — no changes needed there

**PII safeguards** (enforced in `track()` implementation):
- Never pass SQL content, file contents, or data values
- `dashboard_name`, `measure_name`, `dimension_name` are metadata identifiers (resource names), not user data — these are safe to track
- `error_code` should be a structured code, not a raw error message

---

## 6. File Structure

```
web-common/src/lib/analytics/
├── posthog.ts                  # (existing) PostHog init, identify, session ID
├── telemetry-store.ts          # (new) Global properties writable store
├── track.ts                    # (new) Central track() function
├── track-ui.ts                 # (new) UI interaction helpers + Svelte action
└── events/                     # (new) Optional: typed event helpers per surface
    ├── connector-events.ts     #   e.g. trackConnectorSelected(connectorType)
    ├── dashboard-events.ts
    ├── deploy-events.ts
    ├── ai-events.ts
    └── cloud-events.ts
```

The `events/` directory is optional — typed helper functions provide autocomplete and enforce required properties, but `track("event_name", { ... })` works fine for simpler cases. Recommend adding typed helpers for the high-priority funnel events (connectors, deploy) and using raw `track()` for lower-priority UI engagement events.

---

## 7. Answers to Open Questions

### Is PostHog already initialized in `web-common`?
**Yes.** PostHog JS SDK is initialized in both `web-local` and `web-admin` via `initPosthog()` in `web-common/src/lib/analytics/posthog.ts`. It currently handles autocapture, session recording, and heatmaps. **New custom events will NOT go through PostHog** — they route through the Rill custom telemetry pipeline. PostHog continues running as-is for autocapture/recording until a future sunset effort replaces those capabilities.

### Are `rd_project_id` and `rc_project_id` stable identifiers?
**Partially.** Cloud `project_id` is a stable UUID from the admin database. Local `project_id` is currently an **MD5 hash of the project directory name** — this is not ideal because renaming the folder changes the ID. Recommend either:
- (a) Using the existing `install_id` + project path hash (current behavior, good enough), or
- (b) Generating a persistent UUID per project stored in the project's `.rill/` directory (better, but requires a new mechanism)

For the initial rollout, (a) is fine. The `deploy_success` event will link `rd_project_id` to `rc_project_id`, so the cloud ID becomes the stable long-term identifier.

### Should `session_id` be browser session or app session?
**Recommend app session** (new UUID on each page load / `rill start`). This aligns with how PostHog already manages sessions. The `session_id` in our global props supplements PostHog's built-in `$session_id` for cross-referencing with the legacy pipeline if needed. PostHog's own session concept (30-min inactivity timeout) is what the product team will typically use for analysis.

### Server-side event forwarding from Go runtime?
**Not for this phase.** Keep CLI events in the existing `activity.Client` pipeline. The frontend covers all UI interactions. Since both frontend and CLI events now flow through the same Rill pipeline (intake API / Kafka → warehouse), unified funnel analysis (e.g. `rill deploy` CLI → cloud dashboard view) is possible at the warehouse/query layer without any additional plumbing.

### Current opt-out mechanism — does it need extending?
**No.** The current mechanism works as-is. `analytics_enabled: false` prevents `MetricsService` from dispatching, and the new `track()` wrapper checks `analyticsEnabled` before calling dispatch. The existing PostHog init is already gated by the same flag. No new user-facing opt-out surface is needed.

---

## 8. Rollout Plan

| Phase | Scope | Effort Estimate |
|---|---|---|
| **Phase 1** | Foundation: `track()`, telemetry store, init wiring, PostHog identity flow | S–M |
| **Phase 2a** | Connector funnel events (highest PLG priority) | S |
| **Phase 2b** | Deploy funnel events + local→cloud linkage | S |
| **Phase 2c** | Dashboard interaction events | M |
| **Phase 2d** | Model/SQL editor, AI, cloud admin events | M |
| **Phase 2e** | Navigation/UI engagement events + aria-label audit | M–L |
| **Phase 3** | Aria-label comprehensive audit + lint rule | M |

Recommend shipping Phase 1 + 2a + 2b first to unblock PLG funnel analysis, then instrument remaining surfaces incrementally.

---

## 9. Testing & Validation

- **Unit tests:** Verify `track()` merges global props correctly, respects opt-out, and handles missing `metricsService` gracefully
- **Manual validation:** In dev mode, add a console logger to `GenericEventFactory.trackGenericEvent()` to inspect events in the browser console before they're sent to the telemetry client
- **Intake API verification:** After deploy, query the data warehouse to verify events arrive with correct names, global properties, and event-specific properties
- **No-PII audit:** Review all event property values to confirm no SQL, file contents, or user data is captured
- **Opt-out test:** Verify that `analytics: false` suppresses all new events (both legacy and new)
