import { browser } from "$app/environment";
import { writable, derived, get } from "svelte/store";
import { debounce } from "@rilldata/web-common/lib/create-debouncer";
import { runtimeServiceQueryResolver } from "@rilldata/web-common/runtime-client";
import type {
  V1QueryResolverResponse,
  V1StructType,
} from "@rilldata/web-common/runtime-client";

export interface CellState {
  id: string;
  sql: string;
  connector: string;
  limit: number | undefined; // undefined = no limit
  isExecuting: boolean;
  result: V1QueryResolverResponse | null;
  error: string | null;
  executionTimeMs: number | null;
  lastRowCount: number | null; // persisted row count from last execution
  collapsed: boolean;
  hasExecuted: boolean; // true after query runs this session
}

export interface NotebookState {
  cells: CellState[];
  focusedCellId: string | null;
}

const DEFAULT_LIMIT = 100;
const STORAGE_KEY_PREFIX = "rill:query-notebook";

interface PersistedCell {
  id: string;
  sql: string;
  connector: string;
  limit: number | undefined;
  collapsed: boolean;
  resultSchema: V1StructType | null;
  resultRowCount: number | null;
  executionTimeMs: number | null;
}

function storageKey(projectId: string): string {
  return projectId ? `${STORAGE_KEY_PREFIX}:${projectId}` : STORAGE_KEY_PREFIX;
}

function loadPersistedCells(projectId: string): PersistedCell[] | null {
  if (!browser) return null;
  try {
    const stored = localStorage.getItem(storageKey(projectId));
    if (!stored) return null;
    const parsed = JSON.parse(stored);
    if (Array.isArray(parsed) && parsed.length > 0) return parsed;
  } catch {
    // ignore corrupt data
  }
  return null;
}

function saveToLocalStorage(projectId: string, cells: CellState[]) {
  if (!browser) return;
  const persisted: PersistedCell[] = cells.map((c) => ({
    id: c.id,
    sql: c.sql,
    connector: c.connector,
    limit: c.limit,
    collapsed: c.collapsed,
    resultSchema: c.result?.schema ?? null,
    resultRowCount: c.result?.data?.length ?? null,
    executionTimeMs: c.executionTimeMs,
  }));
  localStorage.setItem(storageKey(projectId), JSON.stringify(persisted));
}

function hydrateCell(p: PersistedCell): CellState {
  // Restore schema into a minimal result so the inspector can display it
  const hasSchema = p.resultSchema && p.resultSchema.fields?.length;
  return {
    id: p.id,
    sql: p.sql,
    connector: p.connector,
    limit: p.limit,
    collapsed: p.collapsed,
    isExecuting: false,
    result: hasSchema ? { schema: p.resultSchema!, data: [] } : null,
    error: null,
    executionTimeMs: p.executionTimeMs ?? null,
    lastRowCount: p.resultRowCount ?? null,
    hasExecuted: !!hasSchema,
  };
}

function createDefaultCell(connector: string): CellState {
  return {
    id: crypto.randomUUID(),
    sql: "",
    connector,
    limit: DEFAULT_LIMIT,
    isExecuting: false,
    result: null,
    error: null,
    executionTimeMs: null,
    lastRowCount: null,
    collapsed: false,
    hasExecuted: false,
  };
}

/** Extracts a human-readable message from an API or runtime error */
function extractErrorMessage(err: unknown): string {
  if (err && typeof err === "object") {
    // Axios-style error with response.data.message
    if ("response" in err) {
      const resp = (err as Record<string, unknown>).response;
      if (resp && typeof resp === "object" && "data" in resp) {
        const data = (resp as Record<string, unknown>).data;
        if (data && typeof data === "object" && "message" in data) {
          const msg = (data as Record<string, unknown>).message;
          if (typeof msg === "string" && msg) return msg;
        }
      }
    }
    if (err instanceof Error) return err.message;
  }
  return "Query execution failed";
}

function updateCell(
  state: NotebookState,
  cellId: string,
  updater: (cell: CellState) => CellState,
): NotebookState {
  return {
    ...state,
    cells: state.cells.map((c) => (c.id === cellId ? updater(c) : c)),
  };
}

function createNotebookStore(defaultConnector: string, projectId: string) {
  const persisted = loadPersistedCells(projectId);
  const initialCells = persisted
    ? persisted.map(hydrateCell)
    : [createDefaultCell(defaultConnector)];

  const state = writable<NotebookState>({
    cells: initialCells,
    focusedCellId: initialCells[0]?.id ?? null,
  });

  // Only persist when we have a real connector (skip the throwaway initial store)
  let unsubPersist: (() => void) | undefined;
  if (defaultConnector) {
    const debouncedSave = debounce(
      (cells: CellState[]) => saveToLocalStorage(projectId, cells),
      500,
    );
    unsubPersist = state.subscribe(($s) => debouncedSave($s.cells));
  }

  const { subscribe, update } = state;

  function addCell(connector?: string) {
    const cell = createDefaultCell(connector ?? defaultConnector);
    update((s) => ({
      ...s,
      cells: [...s.cells, cell],
      focusedCellId: cell.id,
    }));
    return cell.id;
  }

  function removeCell(cellId: string) {
    update((s) => {
      if (s.cells.length <= 1) return s; // keep at least 1 cell

      const idx = s.cells.findIndex((c) => c.id === cellId);
      const newCells = s.cells.filter((c) => c.id !== cellId);

      let newFocused = s.focusedCellId;
      if (s.focusedCellId === cellId) {
        // Move focus to previous cell, or first if removed was first
        const newIdx = Math.max(0, idx - 1);
        newFocused = newCells[newIdx]?.id ?? newCells[0]?.id ?? null;
      }

      return { cells: newCells, focusedCellId: newFocused };
    });
  }

  function setCellSql(cellId: string, sql: string) {
    update((s) => updateCell(s, cellId, (c) => ({ ...c, sql })));
  }

  function setCellConnector(cellId: string, connector: string) {
    update((s) => updateCell(s, cellId, (c) => ({ ...c, connector })));
  }

  function setCellLimit(cellId: string, limit: number | undefined) {
    update((s) =>
      updateCell(s, cellId, (c) => ({
        ...c,
        limit: limit !== undefined ? Math.max(1, limit) : undefined,
      })),
    );
  }

  function toggleCellCollapsed(cellId: string) {
    update((s) =>
      updateCell(s, cellId, (c) => ({ ...c, collapsed: !c.collapsed })),
    );
  }

  function setFocusedCell(cellId: string) {
    update((s) => ({ ...s, focusedCellId: cellId }));
  }

  async function executeCellQuery(
    cellId: string,
    instanceId: string,
    sqlOverride?: string,
  ) {
    const current = get(state);
    const cell = current.cells.find((c) => c.id === cellId);
    if (!cell || cell.isExecuting) return;

    const sqlToRun = (sqlOverride ?? cell.sql).trim();
    if (!sqlToRun) return;

    update((s) => ({
      ...updateCell(s, cellId, (c) => ({
        ...c,
        isExecuting: true,
        error: null,
      })),
      focusedCellId: cellId,
    }));

    const startTime = performance.now();

    try {
      const body: {
        resolver: string;
        resolverProperties: { sql: string; connector: string };
        limit?: number;
      } = {
        resolver: "sql",
        resolverProperties: {
          sql: sqlToRun,
          connector: cell.connector,
        },
      };

      if (cell.limit !== undefined) {
        body.limit = cell.limit;
      }

      const response = await runtimeServiceQueryResolver(instanceId, body);
      const elapsed = Math.round(performance.now() - startTime);

      update((s) =>
        updateCell(s, cellId, (c) => ({
          ...c,
          isExecuting: false,
          result: response,
          error: null,
          executionTimeMs: elapsed,
          lastRowCount: response.data?.length ?? 0,
          hasExecuted: true,
        })),
      );
    } catch (err: unknown) {
      const elapsed = Math.round(performance.now() - startTime);
      const message = extractErrorMessage(err);

      update((s) =>
        updateCell(s, cellId, (c) => ({
          ...c,
          isExecuting: false,
          error: message,
          executionTimeMs: elapsed,
          hasExecuted: true,
        })),
      );
    }
  }

  // Derived stores for the focused cell
  const focusedCell = derived(state, ($s) => {
    if (!$s.focusedCellId) return null;
    return $s.cells.find((c) => c.id === $s.focusedCellId) ?? null;
  });

  const focusedSchema = derived(
    focusedCell,
    ($c) => $c?.result?.schema ?? null,
  );
  const focusedData = derived(focusedCell, ($c) => $c?.result?.data ?? null);
  const focusedRowCount = derived(focusedCell, ($c) => {
    // Use live data length if available; fall back to persisted row count
    const liveCount = $c?.result?.data?.length;
    if (liveCount !== undefined) return liveCount;
    return $c?.lastRowCount ?? 0;
  });
  const focusedExecutionTimeMs = derived(
    focusedCell,
    ($c) => $c?.executionTimeMs ?? null,
  );

  function destroy() {
    unsubPersist?.();
  }

  return {
    subscribe,
    destroy,
    addCell,
    removeCell,
    setCellSql,
    setCellConnector,
    setCellLimit,
    toggleCellCollapsed,
    setFocusedCell,
    executeCellQuery,
    focusedSchema,
    focusedData,
    focusedRowCount,
    focusedExecutionTimeMs,
  };
}

export type NotebookStore = ReturnType<typeof createNotebookStore>;

export function createNotebook(
  defaultConnector: string,
  projectId: string,
): NotebookStore {
  return createNotebookStore(defaultConnector, projectId);
}
