import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// =============================================================================
// MOCKS
// =============================================================================

// Mock debounce to call the function synchronously (no delay in tests)
vi.mock("@rilldata/web-common/lib/create-debouncer", () => ({
  debounce: (fn: (...args: unknown[]) => void) => fn,
}));

vi.mock("@rilldata/web-common/runtime-client/v2/gen/runtime-service", () => ({
  runtimeServiceQueryResolver: vi.fn(),
}));

import { runtimeServiceQueryResolver } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { V1QueryResolverResponse } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createNotebook, type NotebookState } from "./query-store";

// =============================================================================
// CONSTANTS
// =============================================================================

const DEFAULT_CONNECTOR = "duckdb";
const PROJECT_ID = "test-org/test-project";
const MOCK_CLIENT = { instanceId: "test-instance" } as unknown as RuntimeClient;

// =============================================================================
// HELPERS
// =============================================================================

function getState(store: ReturnType<typeof createNotebook>): NotebookState {
  return get(store);
}

// =============================================================================
// TESTS
// =============================================================================

describe("createNotebook", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  afterEach(() => {
    localStorage.clear();
  });

  // ---------------------------------------------------------------------------
  // Initial state
  // ---------------------------------------------------------------------------

  describe("initial state", () => {
    it("creates with 1 default cell using the given connector", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const state = getState(store);

      expect(state.cells).toHaveLength(1);
      expect(state.cells[0].connector).toBe(DEFAULT_CONNECTOR);
      expect(state.cells[0].sql).toBe("");
      expect(state.cells[0].limit).toBe(100);
      expect(state.cells[0].isExecuting).toBe(false);
      expect(state.cells[0].result).toBeNull();
      expect(state.cells[0].error).toBeNull();
      expect(state.cells[0].collapsed).toBe(false);
      expect(state.cells[0].hasExecuted).toBe(false);
    });

    it("focuses the first cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const state = getState(store);

      expect(state.focusedCellId).toBe(state.cells[0].id);
    });
  });

  // ---------------------------------------------------------------------------
  // addCell
  // ---------------------------------------------------------------------------

  describe("addCell", () => {
    it("appends a new cell and focuses it", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const newId = store.addCell();
      const state = getState(store);

      expect(state.cells).toHaveLength(2);
      expect(state.cells[1].id).toBe(newId);
      expect(state.focusedCellId).toBe(newId);
    });

    it("returns the new cell id", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const newId = store.addCell();

      expect(typeof newId).toBe("string");
      expect(newId.length).toBeGreaterThan(0);
    });

    it("uses the default connector when none is specified", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      store.addCell();
      const state = getState(store);

      expect(state.cells[1].connector).toBe(DEFAULT_CONNECTOR);
    });

    it("uses a custom connector when specified", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      store.addCell("clickhouse");
      const state = getState(store);

      expect(state.cells[1].connector).toBe("clickhouse");
    });
  });

  // ---------------------------------------------------------------------------
  // removeCell
  // ---------------------------------------------------------------------------

  describe("removeCell", () => {
    it("removes the specified cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const id1 = getState(store).cells[0].id;
      const id2 = store.addCell();

      store.removeCell(id2);
      const state = getState(store);

      expect(state.cells).toHaveLength(1);
      expect(state.cells[0].id).toBe(id1);
    });

    it("cannot remove the last remaining cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const onlyId = getState(store).cells[0].id;

      store.removeCell(onlyId);
      const state = getState(store);

      expect(state.cells).toHaveLength(1);
      expect(state.cells[0].id).toBe(onlyId);
    });

    it("moves focus to the previous cell when removing the focused cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      getState(store).cells[0].id;
      const id2 = store.addCell();
      const id3 = store.addCell();

      // id3 is now focused; remove it
      store.removeCell(id3);
      const state = getState(store);

      expect(state.focusedCellId).toBe(id2);
    });

    it("moves focus to first cell when removing the first focused cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const id1 = getState(store).cells[0].id;
      const id2 = store.addCell();

      // Focus the first cell, then remove it
      store.setFocusedCell(id1);
      store.removeCell(id1);
      const state = getState(store);

      expect(state.focusedCellId).toBe(id2);
    });

    it("does not change focus when removing an unfocused cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const id1 = getState(store).cells[0].id;
      store.addCell();
      const id3 = store.addCell();

      // id3 is focused; remove id1
      store.removeCell(id1);
      const state = getState(store);

      expect(state.focusedCellId).toBe(id3);
    });
  });

  // ---------------------------------------------------------------------------
  // setCellSql
  // ---------------------------------------------------------------------------

  describe("setCellSql", () => {
    it("updates the SQL of the specified cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;

      store.setCellSql(cellId, "SELECT 1");
      const state = getState(store);

      expect(state.cells[0].sql).toBe("SELECT 1");
    });
  });

  // ---------------------------------------------------------------------------
  // setCellConnector
  // ---------------------------------------------------------------------------

  describe("setCellConnector", () => {
    it("updates the connector of the specified cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;

      store.setCellConnector(cellId, "postgres");
      const state = getState(store);

      expect(state.cells[0].connector).toBe("postgres");
    });
  });

  // ---------------------------------------------------------------------------
  // setCellLimit
  // ---------------------------------------------------------------------------

  describe("setCellLimit", () => {
    it("sets the limit to the provided value", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;

      store.setCellLimit(cellId, 50);
      const state = getState(store);

      expect(state.cells[0].limit).toBe(50);
    });

    it("clamps to a minimum of 1", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;

      store.setCellLimit(cellId, 0);
      const state = getState(store);

      expect(state.cells[0].limit).toBe(1);
    });

    it("clamps negative values to 1", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;

      store.setCellLimit(cellId, -10);
      const state = getState(store);

      expect(state.cells[0].limit).toBe(1);
    });

    it("sets undefined for no limit", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;

      store.setCellLimit(cellId, undefined);
      const state = getState(store);

      expect(state.cells[0].limit).toBeUndefined();
    });
  });

  // ---------------------------------------------------------------------------
  // toggleCellCollapsed
  // ---------------------------------------------------------------------------

  describe("toggleCellCollapsed", () => {
    it("toggles collapsed from false to true", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;

      store.toggleCellCollapsed(cellId);

      expect(getState(store).cells[0].collapsed).toBe(true);
    });

    it("toggles collapsed back to false", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;

      store.toggleCellCollapsed(cellId);
      store.toggleCellCollapsed(cellId);

      expect(getState(store).cells[0].collapsed).toBe(false);
    });
  });

  // ---------------------------------------------------------------------------
  // setFocusedCell
  // ---------------------------------------------------------------------------

  describe("setFocusedCell", () => {
    it("changes the focused cell", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const id1 = getState(store).cells[0].id;
      store.addCell();

      store.setFocusedCell(id1);

      expect(getState(store).focusedCellId).toBe(id1);
    });
  });

  // ---------------------------------------------------------------------------
  // executeCellQuery
  // ---------------------------------------------------------------------------

  describe("executeCellQuery", () => {
    it("sets result and hasExecuted on success", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT 1");

      const mockResponse = {
        schema: { fields: [{ name: "col1", type: { code: "CODE_INT32" } }] },
        data: [{ col1: 1 }],
      } as V1QueryResolverResponse;
      vi.mocked(runtimeServiceQueryResolver).mockResolvedValue(mockResponse);

      await store.executeCellQuery(cellId, MOCK_CLIENT);
      const cell = getState(store).cells[0];

      expect(cell.result).toEqual(mockResponse);
      expect(cell.hasExecuted).toBe(true);
      expect(cell.isExecuting).toBe(false);
      expect(cell.error).toBeNull();
      expect(cell.executionTimeMs).toBeTypeOf("number");
      expect(cell.lastRowCount).toBe(1);
    });

    it("sets isExecuting to true while query is in flight", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT 1");

      // Use a deferred promise so we can inspect state mid-execution
      let resolveQuery!: (value: unknown) => void;
      vi.mocked(runtimeServiceQueryResolver).mockReturnValue(
        new Promise((resolve) => {
          resolveQuery = resolve;
        }),
      );

      const promise = store.executeCellQuery(cellId, MOCK_CLIENT);

      // While in flight, isExecuting should be true
      expect(getState(store).cells[0].isExecuting).toBe(true);

      resolveQuery({ schema: null, data: [] });
      await promise;

      expect(getState(store).cells[0].isExecuting).toBe(false);
    });

    it("sets error and hasExecuted on failure", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT bad_column");

      vi.mocked(runtimeServiceQueryResolver).mockRejectedValue(
        new Error("Syntax error"),
      );

      await store.executeCellQuery(cellId, MOCK_CLIENT);
      const cell = getState(store).cells[0];

      expect(cell.error).toBe("Syntax error");
      expect(cell.hasExecuted).toBe(true);
      expect(cell.isExecuting).toBe(false);
    });

    it("extracts error message from response.data.message", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT 1");

      const apiError = {
        response: { data: { message: "API-level error" } },
      };
      vi.mocked(runtimeServiceQueryResolver).mockRejectedValue(apiError);

      await store.executeCellQuery(cellId, MOCK_CLIENT);

      expect(getState(store).cells[0].error).toBe("API-level error");
    });

    it("does nothing when cell SQL is empty", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      // SQL is "" by default

      await store.executeCellQuery(cellId, MOCK_CLIENT);

      expect(runtimeServiceQueryResolver).not.toHaveBeenCalled();
      expect(getState(store).cells[0].isExecuting).toBe(false);
    });

    it("does nothing when cell SQL is only whitespace", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "   ");

      await store.executeCellQuery(cellId, MOCK_CLIENT);

      expect(runtimeServiceQueryResolver).not.toHaveBeenCalled();
    });

    it("uses sqlOverride instead of cell SQL when provided", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT original");

      vi.mocked(runtimeServiceQueryResolver).mockResolvedValue({
        schema: undefined,
        data: [],
      });

      await store.executeCellQuery(cellId, MOCK_CLIENT, "SELECT override");

      expect(runtimeServiceQueryResolver).toHaveBeenCalledWith(
        MOCK_CLIENT,
        expect.objectContaining({
          resolverProperties: expect.objectContaining({
            sql: "SELECT override",
          }),
        }),
        expect.objectContaining({ signal: expect.any(AbortSignal) }),
      );
    });

    it("passes connector and limit in the request body", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT 1");
      store.setCellLimit(cellId, 25);

      vi.mocked(runtimeServiceQueryResolver).mockResolvedValue({
        schema: undefined,
        data: [],
      });

      await store.executeCellQuery(cellId, MOCK_CLIENT);

      expect(runtimeServiceQueryResolver).toHaveBeenCalledWith(
        MOCK_CLIENT,
        expect.objectContaining({
          resolver: "sql",
          resolverProperties: {
            sql: "SELECT 1",
            connector: DEFAULT_CONNECTOR,
          },
          limit: 25,
        }),
        expect.objectContaining({ signal: expect.any(AbortSignal) }),
      );
    });

    it("omits limit from the request when cell limit is undefined", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT 1");
      store.setCellLimit(cellId, undefined);

      vi.mocked(runtimeServiceQueryResolver).mockResolvedValue({
        schema: undefined,
        data: [],
      });

      await store.executeCellQuery(cellId, MOCK_CLIENT);

      const callArgs = vi.mocked(runtimeServiceQueryResolver).mock.calls[0][1];
      expect(callArgs).not.toHaveProperty("limit");
    });

    it("focuses the executed cell", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const id1 = getState(store).cells[0].id;
      store.addCell();

      store.setCellSql(id1, "SELECT 1");
      // id2 is focused after addCell; executing id1 should refocus it
      vi.mocked(runtimeServiceQueryResolver).mockResolvedValue({
        schema: undefined,
        data: [],
      });

      await store.executeCellQuery(id1, MOCK_CLIENT);

      expect(getState(store).focusedCellId).toBe(id1);
    });

    it("does nothing for a nonexistent cell id", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);

      await store.executeCellQuery("nonexistent-id", MOCK_CLIENT);

      expect(runtimeServiceQueryResolver).not.toHaveBeenCalled();
    });

    it("aborts previous in-flight query when re-executed", async () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT 1");

      let rejectFirst!: (reason: unknown) => void;
      vi.mocked(runtimeServiceQueryResolver)
        .mockReturnValueOnce(
          new Promise((_resolve, reject) => {
            rejectFirst = reject;
          }),
        )
        .mockResolvedValueOnce({ schema: undefined, data: [] });

      // Fire first execution (will be in-flight)
      const first = store.executeCellQuery(cellId, MOCK_CLIENT);
      expect(getState(store).cells[0].isExecuting).toBe(true);

      // Second call aborts the first and starts a new execution
      const second = store.executeCellQuery(cellId, MOCK_CLIENT);

      // The first call's signal should be aborted
      const firstSignal = vi.mocked(runtimeServiceQueryResolver).mock
        .calls[0][2]?.signal;
      expect(firstSignal?.aborted).toBe(true);

      // Both calls were made
      expect(runtimeServiceQueryResolver).toHaveBeenCalledTimes(2);

      // Resolve/reject the first (should be ignored due to abort)
      rejectFirst(new DOMException("aborted", "AbortError"));
      await first;
      await second;

      // Cell should have completed from the second call
      expect(getState(store).cells[0].isExecuting).toBe(false);
    });
  });

  // ---------------------------------------------------------------------------
  // Per-project localStorage isolation
  // ---------------------------------------------------------------------------

  describe("per-project localStorage", () => {
    it("persists to a project-scoped key", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "SELECT 1");

      const stored = localStorage.getItem(`rill:query-notebook:${PROJECT_ID}`);
      expect(stored).not.toBeNull();
      const parsed = JSON.parse(stored!);
      expect(parsed[0].sql).toBe("SELECT 1");
    });

    it("isolates state between different projects", () => {
      const storeA = createNotebook(DEFAULT_CONNECTOR, "org/project-a");
      const cellA = getState(storeA).cells[0].id;
      storeA.setCellSql(cellA, "SELECT a");

      const storeB = createNotebook(DEFAULT_CONNECTOR, "org/project-b");
      const cellB = getState(storeB).cells[0].id;
      storeB.setCellSql(cellB, "SELECT b");

      const storedA = JSON.parse(
        localStorage.getItem("rill:query-notebook:org/project-a")!,
      );
      const storedB = JSON.parse(
        localStorage.getItem("rill:query-notebook:org/project-b")!,
      );

      expect(storedA[0].sql).toBe("SELECT a");
      expect(storedB[0].sql).toBe("SELECT b");
    });
  });

  // ---------------------------------------------------------------------------
  // destroy
  // ---------------------------------------------------------------------------

  describe("destroy", () => {
    it("stops persisting after destroy is called", () => {
      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cellId = getState(store).cells[0].id;
      store.setCellSql(cellId, "before destroy");

      const key = `rill:query-notebook:${PROJECT_ID}`;
      const before = localStorage.getItem(key);
      expect(before).not.toBeNull();

      store.destroy();
      localStorage.removeItem(key);

      store.setCellSql(cellId, "after destroy");
      expect(localStorage.getItem(key)).toBeNull();
    });
  });

  // ---------------------------------------------------------------------------
  // Hydration
  // ---------------------------------------------------------------------------

  describe("hydration", () => {
    it("restores schema but marks hasExecuted as false (no live data)", () => {
      const key = `rill:query-notebook:${PROJECT_ID}`;
      const persisted = [
        {
          id: "cell-1",
          sql: "SELECT 1",
          connector: DEFAULT_CONNECTOR,
          limit: 100,
          collapsed: false,
          resultSchema: {
            fields: [{ name: "col1", type: { code: "CODE_INT32" } }],
          },
          resultRowCount: 1,
          executionTimeMs: 50,
        },
      ];
      localStorage.setItem(key, JSON.stringify(persisted));

      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cell = getState(store).cells[0];

      expect(cell.hasExecuted).toBe(false);
      expect(cell.result).not.toBeNull();
      expect(cell.result?.schema).not.toBeNull();
      expect(cell.sql).toBe("SELECT 1");
    });

    it("restores hasExecuted as false when no schema was persisted", () => {
      const key = `rill:query-notebook:${PROJECT_ID}`;
      const persisted = [
        {
          id: "cell-2",
          sql: "SELECT 1",
          connector: DEFAULT_CONNECTOR,
          limit: 100,
          collapsed: false,
          resultSchema: null,
          resultRowCount: null,
          executionTimeMs: null,
        },
      ];
      localStorage.setItem(key, JSON.stringify(persisted));

      const store = createNotebook(DEFAULT_CONNECTOR, PROJECT_ID);
      const cell = getState(store).cells[0];

      expect(cell.hasExecuted).toBe(false);
      expect(cell.result).toBeNull();
    });
  });
});
