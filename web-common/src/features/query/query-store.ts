import type { PartialMessage, Struct } from "@bufbuild/protobuf";
import { browser } from "$app/environment";
import { get, writable } from "svelte/store";
import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
import type { V1QueryResolverResponse } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { runtimeServiceQueryResolver } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { SqlExecution } from "./query-utils";

export interface SqlWorksheetState {
  sql: string;
  isExecuting: boolean;
  hasExecuted: boolean;
  result: V1QueryResolverResponse | null;
  error: string | null;
  executionTimeMs: number | null;
  lastExecution: SqlExecution | null;
}

export const SQL_CONSOLE_ROW_LIMIT = 1_000;

const worksheetCache = new Map<string, string>();

export function createSqlWorksheet(connector: string, cacheKey: string) {
  const state = writable<SqlWorksheetState>({
    sql: browser ? (worksheetCache.get(cacheKey) ?? "") : "",
    isExecuting: false,
    hasExecuted: false,
    result: null,
    error: null,
    executionTimeMs: null,
    lastExecution: null,
  });
  let abortController: AbortController | null = null;

  function setSql(sql: string) {
    state.update(($state) => ({ ...$state, sql }));
    if (browser) worksheetCache.set(cacheKey, sql);
  }

  async function execute(client: RuntimeClient, execution: SqlExecution) {
    if (get(state).isExecuting || !execution.sql.trim()) return;

    abortController = new AbortController();
    state.update(($state) => ({
      ...$state,
      isExecuting: true,
      error: null,
    }));
    const startedAt = performance.now();
    const currentController = abortController;

    try {
      const result = await runtimeServiceQueryResolver(
        client,
        {
          resolver: "sql",
          resolverProperties: {
            sql: execution.sql,
            connector,
          } as unknown as PartialMessage<Struct>,
          limit: SQL_CONSOLE_ROW_LIMIT,
        },
        { signal: currentController.signal },
      );

      state.update(($state) => ({
        ...$state,
        isExecuting: false,
        hasExecuted: true,
        result,
        error: null,
        executionTimeMs: Math.round(performance.now() - startedAt),
        lastExecution: execution,
      }));
    } catch (error) {
      if (currentController.signal.aborted) return;

      // Leave result and executionTimeMs untouched so the last successful
      // results stay visible below the error banner.
      state.update(($state) => ({
        ...$state,
        isExecuting: false,
        hasExecuted: true,
        error: extractErrorMessage(error),
      }));
    } finally {
      if (abortController === currentController) abortController = null;
    }
  }

  function cancel() {
    abortController?.abort();
    abortController = null;
    state.update(($state) => ({ ...$state, isExecuting: false }));
  }

  function destroy() {
    abortController?.abort();
    abortController = null;
  }

  return {
    subscribe: state.subscribe,
    setSql,
    execute,
    cancel,
    destroy,
  };
}

export type SqlWorksheet = ReturnType<typeof createSqlWorksheet>;

export function clearSqlWorksheetCache() {
  worksheetCache.clear();
}
