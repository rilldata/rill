import { describe, expect, it } from "vitest";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { getResultTable } from "./result-table";

describe("getResultTable", () => {
  it("returns a model's result table in the connector that holds it", () => {
    const model: V1Resource = {
      meta: { name: { kind: ResourceKind.Model, name: "orders" } },
      model: {
        spec: { outputConnector: "duckdb" },
        state: { resultConnector: "clickhouse", resultTable: "orders_mart" },
      },
    };

    expect(getResultTable(model)).toEqual({
      connector: "clickhouse",
      database: "",
      databaseSchema: "",
      tableName: "orders_mart",
    });
  });

  it("returns a source's table in its sink connector", () => {
    const source: V1Resource = {
      meta: { name: { kind: ResourceKind.Source, name: "raw_orders" } },
      source: {
        spec: { sourceConnector: "s3", sinkConnector: "duckdb" },
        state: { connector: "duckdb", table: "raw_orders" },
      },
    };

    expect(getResultTable(source)).toEqual({
      connector: "duckdb",
      database: "",
      databaseSchema: "",
      tableName: "raw_orders",
    });
  });

  it("returns undefined for a model that has not produced a result table", () => {
    const model: V1Resource = {
      meta: { name: { kind: ResourceKind.Model, name: "orders" } },
      model: { spec: { outputConnector: "duckdb" }, state: {} },
    };

    expect(getResultTable(model)).toBeUndefined();
  });

  it("returns undefined for a connector, whose name is not a table", () => {
    const connector: V1Resource = {
      meta: { name: { kind: ResourceKind.Connector, name: "duckdb" } },
      connector: { spec: { driver: "duckdb" }, state: {} },
    };

    expect(getResultTable(connector)).toBeUndefined();
  });

  it("returns undefined for a metrics view", () => {
    const metricsView: V1Resource = {
      meta: {
        name: { kind: ResourceKind.MetricsView, name: "orders_metrics" },
      },
      metricsView: {
        spec: { connector: "duckdb", table: "orders" },
        state: {},
      },
    };

    expect(getResultTable(metricsView)).toBeUndefined();
  });

  it("returns undefined for a ref that has not loaded", () => {
    expect(getResultTable(undefined)).toBeUndefined();
  });
});
