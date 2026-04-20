import { describe, it, expect } from "vitest";
import {
  applyDuckLakeFormTransform,
  composeDuckLakeAttach,
} from "./ducklake-utils";
import { ducklakeSchema } from "./ducklake";

describe("composeDuckLakeAttach", () => {
  it("returns empty string when catalog is missing", () => {
    expect(composeDuckLakeAttach({})).toBe("");
    expect(composeDuckLakeAttach({ catalog: "   " })).toBe("");
  });

  it("builds a minimal clause from just the catalog", () => {
    expect(composeDuckLakeAttach({ catalog: "duckdb_database.ducklake" })).toBe(
      "'ducklake:duckdb_database.ducklake'",
    );
  });

  it("includes an alias when provided", () => {
    expect(
      composeDuckLakeAttach({
        catalog: "duckdb_database.ducklake",
        alias: "my_ducklake",
      }),
    ).toBe("'ducklake:duckdb_database.ducklake' AS my_ducklake");
  });

  it("appends quoted string options", () => {
    expect(
      composeDuckLakeAttach({
        catalog: "duckdb_database.ducklake",
        data_path: "other_data_path/",
      }),
    ).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/')",
    );
  });

  it("emits boolean options regardless of value", () => {
    expect(
      composeDuckLakeAttach({
        catalog: "c.ducklake",
        override_data_path: true,
        create_if_not_exists: true,
      }),
    ).toBe(
      "'ducklake:c.ducklake' (OVERRIDE_DATA_PATH true, CREATE_IF_NOT_EXISTS true)",
    );
  });

  it("emits false and true boolean options", () => {
    expect(
      composeDuckLakeAttach({
        catalog: "c.ducklake",
        override_data_path: false,
        encrypted: true,
      }),
    ).toBe("'ducklake:c.ducklake' (OVERRIDE_DATA_PATH false, ENCRYPTED true)");
  });

  it("emits METADATA_PARAMETERS without wrapping quotes", () => {
    expect(
      composeDuckLakeAttach({
        catalog: "c.ducklake",
        metadata_parameters: "{foo: 'bar'}",
      }),
    ).toBe("'ducklake:c.ducklake' (METADATA_PARAMETERS {foo: 'bar'})");
  });

  it("escapes single quotes inside values", () => {
    expect(
      composeDuckLakeAttach({
        catalog: "c.ducklake",
        data_path: "path/with'quote",
      }),
    ).toBe("'ducklake:c.ducklake' (DATA_PATH 'path/with''quote')");
  });

  it("composes the example from the DuckLake docs", () => {
    expect(
      composeDuckLakeAttach({
        catalog: "duckdb_database.ducklake",
        data_path: "other_data_path/",
        override_data_path: true,
      }),
    ).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/', OVERRIDE_DATA_PATH true)",
    );
  });

  it("emits every advanced parameter when set", () => {
    expect(
      composeDuckLakeAttach({
        catalog: "c.ducklake",
        alias: "my_ducklake",
        data_path: "s3://bucket/",
        override_data_path: false,
        create_if_not_exists: false,
        data_inlining_row_limit: 100,
        encrypted: true,
        // mode is a separate YAML key, not part of ATTACH

        meta_parameter_name: "foo",
        metadata_catalog: "meta_cat",
        metadata_parameters: "{a: 1}",
        metadata_path: "postgres:dbname=x",
        metadata_schema: "my_schema",
        automatic_migration: true,
        snapshot_time: "2024-01-01",
        snapshot_version: "v1",
      }),
    ).toBe(
      "'ducklake:c.ducklake' AS my_ducklake (" +
        "DATA_PATH 's3://bucket/', " +
        "META_PARAMETER_NAME 'foo', " +
        "METADATA_CATALOG 'meta_cat', " +
        "METADATA_PATH 'postgres:dbname=x', " +
        "METADATA_SCHEMA 'my_schema', " +
        "SNAPSHOT_TIME '2024-01-01', " +
        "SNAPSHOT_VERSION 'v1', " +
        "METADATA_PARAMETERS {a: 1}, " +
        "DATA_INLINING_ROW_LIMIT 100, " +
        "OVERRIDE_DATA_PATH false, " +
        "CREATE_IF_NOT_EXISTS false, " +
        "ENCRYPTED true, " +
        "AUTOMATIC_MIGRATION true" +
        ")",
    );
  });
});

describe("applyDuckLakeFormTransform", () => {
  it("returns values unchanged for non-DuckLake schemas", () => {
    const values = { attach: "foo" };
    expect(applyDuckLakeFormTransform(null, values)).toBe(values);
    expect(
      applyDuckLakeFormTransform(
        { type: "object", title: "DuckLake", properties: {} },
        values,
      ),
    ).toBe(values);
  });

  it("leaves attach alone when in SQL mode", () => {
    const values = { connection_mode: "sql", attach: "user input" };
    expect(applyDuckLakeFormTransform(ducklakeSchema, values)).toBe(values);
  });

  it("synthesises attach from params when in parameters mode", () => {
    const result = applyDuckLakeFormTransform(ducklakeSchema, {
      connection_mode: "parameters",
      catalog: "duckdb_database.ducklake",
      data_path: "other_data_path/",
    });
    expect(result.attach).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/')",
    );
  });
});
