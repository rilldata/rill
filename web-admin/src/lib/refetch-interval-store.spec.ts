import { describe, it, expect, beforeEach } from "vitest";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  isResourceReconciling,
  createSmartRefetchInterval,
  smartRefetchIntervalFunc,
  INITIAL_REFETCH_INTERVAL,
  MAX_REFETCH_INTERVAL,
  BACKOFF_FACTOR,
} from "./refetch-interval-store";

// Helpers to build minimal resource/query fixtures

function makeResource(
  overrides: Partial<V1Resource> & { reconcileStatus?: string } = {},
): V1Resource {
  const { reconcileStatus = "RECONCILE_STATUS_IDLE", ...rest } = overrides;
  return {
    meta: { reconcileStatus, name: { kind: "test", name: "r" } },
    ...rest,
  } as unknown as V1Resource;
}

// `initializing` defaults to false to match the wire format: the runtime always
// serializes the field, so it is never undefined on a successful response.
function makeQuery(resources: V1Resource[] | undefined, initializing = false) {
  // Minimal query-shaped object matching what refetchInterval receives.
  return {
    state: {
      data: resources ? { resources, initializing } : undefined,
    },
  } as any;
}

// ─── isResourceReconciling ───────────────────────────────────────────

describe("isResourceReconciling", () => {
  it("returns true for PENDING", () => {
    const r = makeResource({ reconcileStatus: "RECONCILE_STATUS_PENDING" });
    expect(isResourceReconciling(r)).toBe(true);
  });

  it("returns true for RUNNING", () => {
    const r = makeResource({ reconcileStatus: "RECONCILE_STATUS_RUNNING" });
    expect(isResourceReconciling(r)).toBe(true);
  });

  it("returns false for IDLE", () => {
    const r = makeResource({ reconcileStatus: "RECONCILE_STATUS_IDLE" });
    expect(isResourceReconciling(r)).toBe(false);
  });

  it("returns false for undefined status", () => {
    const r = { meta: {} } as V1Resource;
    expect(isResourceReconciling(r)).toBe(false);
  });
});

// ─── smartRefetchIntervalFunc (unfiltered) ───────────────────────────

describe("smartRefetchIntervalFunc", () => {
  it("returns false when no data", () => {
    expect(smartRefetchIntervalFunc(makeQuery(undefined))).toBe(false);
  });

  it("returns false when all resources are idle", () => {
    const q = makeQuery([makeResource(), makeResource()]);
    expect(smartRefetchIntervalFunc(q)).toBe(false);
  });

  it("returns initial interval when a resource starts reconciling", () => {
    const q = makeQuery([
      makeResource({ reconcileStatus: "RECONCILE_STATUS_RUNNING" }),
    ]);
    expect(smartRefetchIntervalFunc(q)).toBe(INITIAL_REFETCH_INTERVAL);
  });

  it("backs off on consecutive reconciling checks", () => {
    const q = makeQuery([
      makeResource({ reconcileStatus: "RECONCILE_STATUS_RUNNING" }),
    ]);
    const first = smartRefetchIntervalFunc(q);
    const second = smartRefetchIntervalFunc(q);
    expect(first).toBe(INITIAL_REFETCH_INTERVAL);
    expect(second).toBe(INITIAL_REFETCH_INTERVAL * BACKOFF_FACTOR);
  });

  it("caps at MAX_REFETCH_INTERVAL", () => {
    const q = makeQuery([
      makeResource({ reconcileStatus: "RECONCILE_STATUS_RUNNING" }),
    ]);
    // Call enough times to exceed max
    let interval: number | false = false;
    for (let i = 0; i < 20; i++) {
      interval = smartRefetchIntervalFunc(q);
    }
    expect(interval).toBe(MAX_REFETCH_INTERVAL);
  });

  it("resets interval after reconciliation completes and restarts", () => {
    const q = makeQuery([
      makeResource({ reconcileStatus: "RECONCILE_STATUS_RUNNING" }),
    ]);
    // Start reconciling
    smartRefetchIntervalFunc(q);
    smartRefetchIntervalFunc(q);

    // Finish reconciling
    q.state.data.resources = [makeResource()];
    expect(smartRefetchIntervalFunc(q)).toBe(false);

    // Start reconciling again
    q.state.data.resources = [
      makeResource({ reconcileStatus: "RECONCILE_STATUS_PENDING" }),
    ];
    expect(smartRefetchIntervalFunc(q)).toBe(INITIAL_REFETCH_INTERVAL);
  });
});

// ─── createSmartRefetchInterval (filtered) ───────────────────────────

describe("createSmartRefetchInterval", () => {
  const isDashboard = (r: V1Resource) => !!r.canvas || !!r.explore;
  let refetchInterval: ReturnType<typeof createSmartRefetchInterval>;

  beforeEach(() => {
    // Each test gets a fresh function instance (fresh WeakMap state)
    refetchInterval = createSmartRefetchInterval(isDashboard);
  });

  it("polls when no data (runtime may be initializing)", () => {
    expect(refetchInterval(makeQuery(undefined))).toBe(MAX_REFETCH_INTERVAL);
  });

  it("returns false when all relevant resources are idle", () => {
    const q = makeQuery([
      makeResource({ explore: {} } as any),
      makeResource({ canvas: {} } as any),
    ]);
    expect(refetchInterval(q)).toBe(false);
  });

  it("polls when a relevant resource is reconciling", () => {
    const q = makeQuery([
      makeResource({
        reconcileStatus: "RECONCILE_STATUS_RUNNING",
        explore: {},
      } as any),
    ]);
    expect(refetchInterval(q)).toBe(INITIAL_REFETCH_INTERVAL);
  });

  it("ignores non-relevant reconciling resources when relevant ones exist", () => {
    const q = makeQuery([
      makeResource({ explore: {} } as any), // idle dashboard
      makeResource({ reconcileStatus: "RECONCILE_STATUS_RUNNING" }), // reconciling model
    ]);
    expect(refetchInterval(q)).toBe(false);
  });

  // ─── No relevant resources yet ───────────────────────────────────
  //
  // The runtime hides the ProjectParser from non-admins, so the response can't
  // be used to tell "still building" from "nothing to show". ListResources
  // reports that in `initializing` instead.

  it("polls when no relevant resources and the runtime is still initializing", () => {
    // Runtime restart: models are still being built, dashboards don't exist yet.
    const q = makeQuery(
      [makeResource({ reconcileStatus: "RECONCILE_STATUS_RUNNING" })],
      true,
    );
    expect(refetchInterval(q)).toBe(MAX_REFETCH_INTERVAL);
  });

  it("stops polling when no relevant resources and the runtime is done initializing", () => {
    const q = makeQuery([makeResource()], false);
    expect(refetchInterval(q)).toBe(false);
  });

  // ─── Empty resource list ─────────────────────────────────────────

  it("polls when the resource list is empty and the runtime is still initializing", () => {
    const q = makeQuery([], true);
    expect(refetchInterval(q)).toBe(MAX_REFETCH_INTERVAL);
  });

  it("stops polling when the resource list is empty and the runtime is done initializing", () => {
    // Security policies denied every resource. ListResources returns 200 with an
    // empty list, which used to be read as "runtime not ready" and polled forever.
    const q = makeQuery([], false);
    expect(refetchInterval(q)).toBe(false);
  });

  // ─── Error state (runtime returning errors) ──────────────────────

  it("polls when query has errored with no prior data", () => {
    // Runtime not ready, ListResources returns 400.
    // query.state.data is undefined (no prior successful fetch).
    const q = makeQuery(undefined);
    q.state.status = "error";
    expect(refetchInterval(q)).not.toBe(false);
  });
});
