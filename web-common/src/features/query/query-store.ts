import { writable, derived, get } from "svelte/store";
import { runtimeServiceQueryResolver } from "@rilldata/web-common/runtime-client";
import type {
  V1QueryResolverResponse,
  V1StructType,
  V1QueryResolverResponseDataItem,
} from "@rilldata/web-common/runtime-client";

export interface QueryConsoleState {
  sql: string;
  connector: string;
  limit: number;
  isExecuting: boolean;
  result: V1QueryResolverResponse | null;
  error: string | null;
  executionTimeMs: number | null;
}

const DEFAULT_LIMIT = 100;

function createQueryConsoleStore() {
  const state = writable<QueryConsoleState>({
    sql: "",
    connector: "",
    limit: DEFAULT_LIMIT,
    isExecuting: false,
    result: null,
    error: null,
    executionTimeMs: null,
  });

  const { subscribe, update } = state;

  function setSql(sql: string) {
    update((s) => ({ ...s, sql }));
  }

  function setConnector(connector: string) {
    update((s) => ({ ...s, connector }));
  }

  function setLimit(limit: number) {
    update((s) => ({ ...s, limit: Math.max(1, limit) }));
  }

  async function executeQuery(instanceId: string) {
    const current = get(state);
    const sql = current.sql.trim();
    if (!sql) return;

    update((s) => ({
      ...s,
      isExecuting: true,
      error: null,
    }));

    const startTime = performance.now();

    try {
      const response = await runtimeServiceQueryResolver(instanceId, {
        resolver: "sql",
        resolverProperties: {
          sql,
          connector: current.connector,
        },
        limit: current.limit,
      });

      const elapsed = Math.round(performance.now() - startTime);

      update((s) => ({
        ...s,
        isExecuting: false,
        result: response,
        error: null,
        executionTimeMs: elapsed,
      }));
    } catch (err: unknown) {
      const elapsed = Math.round(performance.now() - startTime);
      const message =
        (err as { response?: { data?: { message?: string } } })?.response?.data
          ?.message ??
        (err as Error)?.message ??
        "Query execution failed";

      update((s) => ({
        ...s,
        isExecuting: false,
        error: message,
        executionTimeMs: elapsed,
      }));
    }
  }

  function reset() {
    update((s) => ({
      ...s,
      result: null,
      error: null,
      executionTimeMs: null,
    }));
  }

  // Derived stores for convenience
  const schema = derived(state, ($s) => $s.result?.schema ?? null);
  const data = derived(state, ($s) => $s.result?.data ?? null);
  const rowCount = derived(state, ($s) => $s.result?.data?.length ?? 0);

  return {
    subscribe,
    setSql,
    setConnector,
    setLimit,
    executeQuery,
    reset,
    schema,
    data,
    rowCount,
  };
}

export type QueryConsoleStore = ReturnType<typeof createQueryConsoleStore>;

export function createQueryConsole(): QueryConsoleStore {
  return createQueryConsoleStore();
}
