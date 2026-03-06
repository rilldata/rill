import { browser } from "$app/environment";
import { writable, derived, get } from "svelte/store";
import { debounce } from "@rilldata/web-common/lib/create-debouncer";
import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
import { runtimeServiceQueryResolver } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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
  try {
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
  } catch {
    // QuotaExceededError or other storage failures; silently ignore
  }
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
    hasExecuted: false,
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

  // Only persist when we have a real connector and project ID
  let unsubPersist: (() => void) | undefined;
  if (defaultConnector && projectId) {
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

  // Per-cell abort controllers for query cancellation
  const abortControllers = new Map<string, AbortController>();

  async function executeCellQuery(
    cellId: string,
    client: RuntimeClient,
    sqlOverride?: string,
  ) {
    const current = get(state);
    const cell = current.cells.find((c) => c.id === cellId);
    if (!cell) return;

    const sqlToRun = (sqlOverride ?? cell.sql).trim();
    if (!sqlToRun) return;

    // Abort any in-flight query for this cell
    abortControllers.get(cellId)?.abort();
    const controller = new AbortController();
    abortControllers.set(cellId, controller);

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
      const response = await runtimeServiceQueryResolver(
        client,
        {
          resolver: "sql",
          resolverProperties: {
            sql: sqlToRun,
            connector: cell.connector,
          },
          ...(cell.limit !== undefined ? { limit: cell.limit } : {}),
        } as Parameters<typeof runtimeServiceQueryResolver>[1],
        { signal: controller.signal },
      );
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
      // Ignore abort errors (user cancelled or re-ran)
      if (controller.signal.aborted) return;

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
    } finally {
      abortControllers.delete(cellId);
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

  function cancelCellQuery(cellId: string) {
    const controller = abortControllers.get(cellId);
    if (!controller) return;
    controller.abort();
    abortControllers.delete(cellId);
    update((s) =>
      updateCell(s, cellId, (c) => ({
        ...c,
        isExecuting: false,
      })),
    );
  }

  function destroy() {
    unsubPersist?.();
    for (const controller of abortControllers.values()) {
      controller.abort();
    }
    abortControllers.clear();
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
    cancelCellQuery,
    focusedSchema,
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
