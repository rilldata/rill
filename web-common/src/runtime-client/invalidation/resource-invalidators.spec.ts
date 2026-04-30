import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getConnectorServiceOLAPListTablesQueryKey,
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetResourceQueryKey,
  V1ReconcileStatus,
  type V1Resource,
  V1ResourceEvent,
  type V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { QueryClient } from "@tanstack/svelte-query";
import { afterEach, describe, expect, it, vi } from "vitest";

const {
  updateArtifacts,
  deleteResource,
  deleteItem,
  isPending,
  trackIngested,
} = vi.hoisted(() => ({
  updateArtifacts: vi.fn(),
  deleteResource: vi.fn(),
  deleteItem: vi.fn(),
  isPending: vi.fn(() => false),
  trackIngested: vi.fn(),
}));

vi.mock(
  "@rilldata/web-common/features/entity-management/file-artifacts",
  () => ({
    fileArtifacts: { updateArtifacts, deleteResource },
  }),
);

vi.mock(
  "@rilldata/web-common/features/connectors/explorer/connector-explorer-store",
  () => ({
    connectorExplorerStore: { deleteItem },
  }),
);

vi.mock("@rilldata/web-common/features/sources/sources-store", () => ({
  sourceIngestionTracker: { isPending, trackIngested },
}));

import {
  handleResourceEvent,
  type ResourceInvalidatorState,
} from "./resource-invalidators";

const INSTANCE_ID = "inst-1";

// `instanceId` is read off the runtimeClient; the rest of the interface is
// only touched by the "source imported successfully" leaf check, which none
// of these fixtures trigger.
const fakeRuntimeClient = { instanceId: INSTANCE_ID } as RuntimeClient;

interface QueryClientMock {
  invalidateQueries: ReturnType<typeof vi.fn>;
  refetchQueries: ReturnType<typeof vi.fn>;
  resetQueries: ReturnType<typeof vi.fn>;
  removeQueries: ReturnType<typeof vi.fn>;
  getQueryData: ReturnType<typeof vi.fn>;
  setQueryData: ReturnType<typeof vi.fn>;
}

function fakeQueryClient(
  previousResource: V1Resource | undefined = undefined,
): QueryClient & QueryClientMock {
  return {
    invalidateQueries: vi.fn(),
    refetchQueries: vi.fn(),
    resetQueries: vi.fn(),
    removeQueries: vi.fn(),
    getQueryData: vi.fn(() =>
      previousResource ? { resource: previousResource } : undefined,
    ),
    setQueryData: vi.fn(),
  } as unknown as QueryClient & QueryClientMock;
}

function makeState(): ResourceInvalidatorState {
  return { connectorNames: new Set<string>() };
}

function writeEvent(
  kind: string,
  name: string,
  resource: V1Resource,
): V1WatchResourcesResponse {
  return {
    event: V1ResourceEvent.RESOURCE_EVENT_WRITE,
    name: { name, kind },
    resource,
  };
}

function deleteEvent(kind: string, name: string): V1WatchResourcesResponse {
  return {
    event: V1ResourceEvent.RESOURCE_EVENT_DELETE,
    name: { name, kind },
  };
}

function baseMeta(overrides: Partial<V1Resource["meta"]> = {}) {
  return {
    reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_IDLE,
    stateVersion: "1",
    ...overrides,
  };
}

function containsQueryKey(
  spy: ReturnType<typeof vi.fn>,
  key: readonly unknown[],
): boolean {
  return spy.mock.calls.some(
    ([arg]) => JSON.stringify(arg.queryKey) === JSON.stringify(key),
  );
}

afterEach(() => {
  vi.clearAllMocks();
});

describe("handleResourceEvent — setup + guards", () => {
  it("short-circuits when event or name is missing", async () => {
    const qc = fakeQueryClient();
    await handleResourceEvent({}, qc, fakeRuntimeClient, makeState());
    expect(qc.setQueryData).not.toHaveBeenCalled();
  });

  it("sets the new resource in the cache before dispatching", async () => {
    const qc = fakeQueryClient();
    await handleResourceEvent(
      writeEvent(ResourceKind.MetricsView, "mv", { meta: baseMeta() }),
      qc,
      fakeRuntimeClient,
      makeState(),
    );
    expect(qc.setQueryData).toHaveBeenCalledWith(
      getRuntimeServiceGetResourceQueryKey(INSTANCE_ID, {
        name: { name: "mv", kind: ResourceKind.MetricsView },
      }),
      { resource: { meta: baseMeta() } },
    );
  });

  it("returns early for the ProjectParser resource", async () => {
    const qc = fakeQueryClient();
    await handleResourceEvent(
      writeEvent(ResourceKind.ProjectParser, "parser", { meta: baseMeta() }),
      qc,
      fakeRuntimeClient,
      makeState(),
    );
    expect(qc.refetchQueries).not.toHaveBeenCalled();
    expect(updateArtifacts).not.toHaveBeenCalled();
  });

  it("skips invalidations when version advanced but reconcile has not finished", async () => {
    // Previous: stateVersion=1, RECONCILING. New: stateVersion=2, RECONCILING.
    // The legacy gate treats this as an intermediate reconcile step and
    // skips query invalidations until the reconcile transitions to IDLE.
    const previous: V1Resource = {
      meta: baseMeta({
        stateVersion: "1",
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      }),
    };
    const qc = fakeQueryClient(previous);
    const event = writeEvent(ResourceKind.MetricsView, "mv", {
      meta: baseMeta({
        stateVersion: "2",
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      }),
    });

    await handleResourceEvent(event, qc, fakeRuntimeClient, makeState());
    expect(qc.refetchQueries).not.toHaveBeenCalled();
  });
});

describe("Connector writes", () => {
  it("invalidates AnalyzeConnectors and the connector's queries", async () => {
    const qc = fakeQueryClient();
    await handleResourceEvent(
      writeEvent(ResourceKind.Connector, "postgres", { meta: baseMeta() }),
      qc,
      fakeRuntimeClient,
      makeState(),
    );
    expect(
      containsQueryKey(
        qc.invalidateQueries,
        getRuntimeServiceAnalyzeConnectorsQueryKey(INSTANCE_ID),
      ),
    ).toBe(true);
  });
});

describe("Source/Model writes", () => {
  it("invalidates the OLAP tables list when the source table changes", async () => {
    // The invalidation is gated on resourceFinishedReconciling (previous was
    // RECONCILING, new is IDLE) OR stateVersions being equal.
    const previous: V1Resource = {
      meta: baseMeta({
        stateVersion: "1",
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      }),
      source: { state: { connector: "duckdb", table: "old_table" } },
    };
    const qc = fakeQueryClient(previous);
    const event = writeEvent(ResourceKind.Source, "users", {
      meta: baseMeta({ stateVersion: "2" }),
      source: { state: { connector: "duckdb", table: "new_table" } },
    });

    await handleResourceEvent(event, qc, fakeRuntimeClient, makeState());
    expect(
      containsQueryKey(
        qc.invalidateQueries,
        getConnectorServiceOLAPListTablesQueryKey(INSTANCE_ID, {
          connector: "duckdb",
        }),
      ),
    ).toBe(true);
  });

  it("invalidates both old and new OLAP tables lists when the connector changes", async () => {
    const previous: V1Resource = {
      meta: baseMeta({
        stateVersion: "1",
        reconcileStatus: V1ReconcileStatus.RECONCILE_STATUS_RUNNING,
      }),
      model: { state: { resultConnector: "duckdb", resultTable: "old" } },
    };
    const qc = fakeQueryClient(previous);
    const event = writeEvent(ResourceKind.Model, "users", {
      meta: baseMeta({ stateVersion: "2" }),
      model: { state: { resultConnector: "postgres", resultTable: "new" } },
    });

    await handleResourceEvent(event, qc, fakeRuntimeClient, makeState());
    expect(
      containsQueryKey(
        qc.invalidateQueries,
        getConnectorServiceOLAPListTablesQueryKey(INSTANCE_ID, {
          connector: "duckdb",
        }),
      ),
    ).toBe(true);
    expect(
      containsQueryKey(
        qc.invalidateQueries,
        getConnectorServiceOLAPListTablesQueryKey(INSTANCE_ID, {
          connector: "postgres",
        }),
      ),
    ).toBe(true);
  });

  it("detects a new connector and invalidates the AnalyzeConnectors query once", async () => {
    const qc = fakeQueryClient();
    const state = makeState();
    const event = writeEvent(ResourceKind.Source, "users", {
      meta: baseMeta({ stateVersion: "1" }),
      source: { state: { connector: "newbie", table: "t" } },
    });

    await handleResourceEvent(event, qc, fakeRuntimeClient, state);
    expect(state.connectorNames.has("newbie")).toBe(true);

    qc.invalidateQueries.mockClear();
    await handleResourceEvent(event, qc, fakeRuntimeClient, state);
    // Second pass: the connector is already tracked, no extra AnalyzeConnectors invalidation.
    expect(
      containsQueryKey(
        qc.invalidateQueries,
        getRuntimeServiceAnalyzeConnectorsQueryKey(INSTANCE_ID),
      ),
    ).toBe(false);
  });

  it("short-circuits profiling invalidation when table name is missing", async () => {
    const qc = fakeQueryClient();
    const event = writeEvent(ResourceKind.Source, "users", {
      meta: baseMeta({ stateVersion: "1" }),
      source: { state: { connector: "duckdb", table: "" } },
    });

    await handleResourceEvent(event, qc, fakeRuntimeClient, makeState());
    // No model-partitions invalidation (would require tableName + Model kind).
    // And since this is a Source with empty table, profiling is skipped too;
    // the assertion is simply that no extra work happened after the
    // connector-names check.
    expect(qc.invalidateQueries.mock.calls.length).toBeLessThanOrEqual(2);
  });
});

describe("Connector deletes", () => {
  it("invalidates AnalyzeConnectors and removes the item from the explorer store", async () => {
    const qc = fakeQueryClient();
    await handleResourceEvent(
      deleteEvent(ResourceKind.Connector, "postgres"),
      qc,
      fakeRuntimeClient,
      makeState(),
    );
    expect(
      containsQueryKey(
        qc.invalidateQueries,
        getRuntimeServiceAnalyzeConnectorsQueryKey(INSTANCE_ID),
      ),
    ).toBe(true);
    expect(deleteItem).toHaveBeenCalledWith("postgres");
  });
});

describe("Source/Model deletes", () => {
  it("invalidates the previous connector's OLAP tables list", async () => {
    const previous: V1Resource = {
      meta: baseMeta({ stateVersion: "1" }),
      source: { state: { connector: "duckdb", table: "t" } },
    };
    const qc = fakeQueryClient(previous);
    await handleResourceEvent(
      deleteEvent(ResourceKind.Source, "users"),
      qc,
      fakeRuntimeClient,
      makeState(),
    );
    expect(
      containsQueryKey(
        qc.invalidateQueries,
        getConnectorServiceOLAPListTablesQueryKey(INSTANCE_ID, {
          connector: "duckdb",
        }),
      ),
    ).toBe(true);
  });
});

describe("fileArtifacts bookkeeping", () => {
  it("updates fileArtifacts on a WRITE", async () => {
    const qc = fakeQueryClient();
    const event = writeEvent(ResourceKind.MetricsView, "mv", {
      meta: baseMeta(),
    });
    await handleResourceEvent(event, qc, fakeRuntimeClient, makeState());
    expect(updateArtifacts).toHaveBeenCalledWith(event.resource);
  });

  it("deletes from fileArtifacts on a DELETE", async () => {
    const qc = fakeQueryClient();
    await handleResourceEvent(
      deleteEvent(ResourceKind.MetricsView, "mv"),
      qc,
      fakeRuntimeClient,
      makeState(),
    );
    expect(deleteResource).toHaveBeenCalledWith({
      name: "mv",
      kind: ResourceKind.MetricsView,
    });
  });
});
