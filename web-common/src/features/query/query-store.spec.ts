import { get } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { V1QueryResolverResponse } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { SqlExecution } from "./query-utils";

vi.mock("@rilldata/web-common/runtime-client/v2/gen/runtime-service", () => ({
  runtimeServiceQueryResolver: vi.fn(),
}));

import { runtimeServiceQueryResolver } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import {
  clearSqlWorksheetCache,
  createSqlWorksheet,
  SQL_CONSOLE_ROW_LIMIT,
} from "./query-store";

const client = { instanceId: "instance" } as RuntimeClient;
const execution: SqlExecution = {
  sql: "SELECT 42",
  from: 8,
  to: 17,
  startLine: 2,
  endLine: 2,
};

describe("createSqlWorksheet", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    clearSqlWorksheetCache();
  });

  it("starts with an empty worksheet", () => {
    const worksheet = createSqlWorksheet("duckdb", "acme/analytics");

    expect(get(worksheet)).toEqual({
      sql: "",
      isExecuting: false,
      hasExecuted: false,
      result: null,
      error: null,
      executionTimeMs: null,
      lastExecution: null,
    });
  });

  it("restores worksheet text from the in-memory cache", () => {
    const worksheet = createSqlWorksheet("duckdb", "acme/analytics");
    worksheet.setSql("SELECT * FROM events");
    worksheet.destroy();

    const restored = createSqlWorksheet("duckdb", "acme/analytics");

    expect(get(restored).sql).toBe("SELECT * FROM events");
    expect(localStorage.length).toBe(0);
  });

  it("separates cached worksheets by project and branch", () => {
    const production = createSqlWorksheet("duckdb", "acme/analytics");
    const branch = createSqlWorksheet("duckdb", "acme/analytics@new-metrics");
    production.setSql("SELECT 'production'");
    branch.setSql("SELECT 'branch'");

    expect(get(createSqlWorksheet("duckdb", "acme/analytics")).sql).toBe(
      "SELECT 'production'",
    );
    expect(
      get(createSqlWorksheet("duckdb", "acme/analytics@new-metrics")).sql,
    ).toBe("SELECT 'branch'");
  });

  it("limits worksheet results to 1,000 rows", async () => {
    const result: V1QueryResolverResponse = {
      schema: { fields: [{ name: "answer" }] },
      data: [{ answer: 42 }],
    };
    vi.mocked(runtimeServiceQueryResolver).mockResolvedValue(result);
    const worksheet = createSqlWorksheet("clickhouse", "acme/analytics");

    await worksheet.execute(client, execution);

    const [calledClient, request, options] = vi.mocked(
      runtimeServiceQueryResolver,
    ).mock.calls[0];
    expect(calledClient).toBe(client);
    expect(request).toMatchObject({
      resolver: "sql",
      resolverProperties: {
        sql: "SELECT 42",
        connector: "clickhouse",
      },
    });
    expect(request.limit).toBe(SQL_CONSOLE_ROW_LIMIT);
    expect(options?.signal).toBeInstanceOf(AbortSignal);
    expect(get(worksheet)).toMatchObject({
      result,
      hasExecuted: true,
      isExecuting: false,
      error: null,
      lastExecution: execution,
    });
  });

  it("does not classify SQL on the client", async () => {
    vi.mocked(runtimeServiceQueryResolver).mockRejectedValue(
      new Error("statement rejected by runtime"),
    );
    const worksheet = createSqlWorksheet("duckdb", "acme/analytics");

    await worksheet.execute(client, {
      ...execution,
      sql: "DELETE FROM events",
    });

    expect(runtimeServiceQueryResolver).toHaveBeenCalledOnce();
    expect(get(worksheet).error).toBe("statement rejected by runtime");
  });

  it("retains the last successful result when a later run fails", async () => {
    const result: V1QueryResolverResponse = { data: [{ value: 1 }] };
    vi.mocked(runtimeServiceQueryResolver)
      .mockResolvedValueOnce(result)
      .mockRejectedValueOnce(new Error("query failed"));
    const worksheet = createSqlWorksheet("duckdb", "acme/analytics");

    await worksheet.execute(client, execution);
    const executionTimeMs = get(worksheet).executionTimeMs;
    await worksheet.execute(client, { ...execution, sql: "SELECT broken" });

    // The retained result keeps its own execution time; the failed run does not
    // overwrite it, so the results header stays consistent with the table.
    expect(get(worksheet)).toMatchObject({
      result,
      error: "query failed",
      lastExecution: execution,
      executionTimeMs,
    });
  });

  it("cancels an in-flight query", async () => {
    vi.mocked(runtimeServiceQueryResolver).mockImplementation(
      (_client, _request, options) =>
        new Promise((_resolve, reject) => {
          options?.signal?.addEventListener("abort", () =>
            reject(new DOMException("Aborted", "AbortError")),
          );
        }),
    );
    const worksheet = createSqlWorksheet("duckdb", "acme/analytics");

    const request = worksheet.execute(client, execution);
    expect(get(worksheet).isExecuting).toBe(true);
    worksheet.cancel();
    await request;

    expect(get(worksheet).isExecuting).toBe(false);
    expect(get(worksheet).error).toBeNull();
  });

  it("does not cache query results across navigation", async () => {
    vi.mocked(runtimeServiceQueryResolver).mockResolvedValue({
      data: [{ answer: 42 }],
    });
    const worksheet = createSqlWorksheet("duckdb", "acme/analytics");
    worksheet.setSql("SELECT 42");
    await worksheet.execute(client, execution);
    worksheet.destroy();

    const restored = createSqlWorksheet("duckdb", "acme/analytics");

    expect(get(restored).sql).toBe("SELECT 42");
    expect(get(restored).result).toBeNull();
    expect(get(restored).hasExecuted).toBe(false);
  });
});
