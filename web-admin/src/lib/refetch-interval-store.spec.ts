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

function makeQuery(resources: V1Resource[] | undefined) {
  // Minimal query-shaped object matching what refetchInterval receives.
  return {
    state: {
      data: resources ? { resources } : undefined,
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

  // ─── Fallback behavior: no relevant resources yet ────────────────

  it("polls when no relevant resources but non-parser resources are reconciling", () => {
    // Simulates runtime restart: models are still being built,
    // dashboards haven't been created yet
    const q = makeQuery([
      makeResource({
        projectParser: {},
        reconcileStatus: "RECONCILE_STATUS_RUNNING",
      } as any),
      makeResource({ reconcileStatus: "RECONCILE_STATUS_RUNNING" }), // model
    ]);
    expect(refetchInterval(q)).toBe(INITIAL_REFETCH_INTERVAL);
  });

  it("stops polling when only ProjectParser is reconciling and no relevant resources", () => {
    // Truly empty project: parser is running but no other resources
    const q = makeQuery([
      makeResource({
        projectParser: {},
        reconcileStatus: "RECONCILE_STATUS_RUNNING",
      } as any),
    ]);
    expect(refetchInterval(q)).toBe(false);
  });

  it("keeps polling when non-parser resources are idle but parser is still reconciling", () => {
    // During wake-up the parser creates resources incrementally;
    // early resources (sources, models) may finish before explores
    // are created. Keep polling so we pick them up.
    const q = makeQuery([
      makeResource({
        projectParser: {},
        reconcileStatus: "RECONCILE_STATUS_RUNNING",
      } as any),
      makeResource(), // idle model
    ]);
    expect(refetchInterval(q)).not.toBe(false);
  });

  it("stops polling when all resources including parser are idle and no relevant resources", () => {
    const q = makeQuery([
      makeResource({
        projectParser: {},
        reconcileStatus: "RECONCILE_STATUS_IDLE",
      } as any),
      makeResource(), // idle model
    ]);
    expect(refetchInterval(q)).toBe(false);
  });

  // ─── Empty resource list (runtime just started) ──────────────────

  it("polls when resource list is completely empty", () => {
    // Runtime just started; not even ProjectParser created yet.
    const q = makeQuery([]);
    expect(refetchInterval(q)).not.toBe(false);
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
