import { describe, it, expect } from "vitest";
import {
  applyDuckLakeFormTransform,
  composeDuckLakeAttach,
} from "./ducklake-utils";
import { ducklakeSchema } from "./ducklake";

describe("composeDuckLakeAttach", () => {
  it("returns empty string when catalog identifier is missing", () => {
    expect(composeDuckLakeAttach({})).toBe("");
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "   ",
      }),
    ).toBe("");
  });

  it("defaults to duckdb catalog when catalog_type is unset", () => {
    expect(
      composeDuckLakeAttach({ catalog_duckdb_path: "catalog.ducklake" }),
    ).toBe("'ducklake:catalog.ducklake'");
  });

  it("builds a minimal clause from a duckdb catalog path", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "duckdb_database.ducklake",
      }),
    ).toBe("'ducklake:duckdb_database.ducklake'");
  });

  it("builds a sqlite catalog clause with the sqlite: prefix", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "sqlite",
        catalog_sqlite_path: "catalog.sqlite",
      }),
    ).toBe("'ducklake:sqlite:catalog.sqlite'");
  });

  it("builds a postgres catalog clause from individual fields", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "postgres",
        catalog_postgres_dbname: "mydb",
        catalog_postgres_host: "localhost",
        catalog_postgres_port: "5432",
        catalog_postgres_user: "postgres",
        catalog_postgres_password: "secret",
      }),
    ).toBe(
      "'ducklake:postgres:dbname=mydb host=localhost port=5432 user=postgres password=secret'",
    );
  });

  it("omits missing postgres params when composing the connection string", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "postgres",
        catalog_postgres_dbname: "mydb",
        catalog_postgres_host: "localhost",
      }),
    ).toBe("'ducklake:postgres:dbname=mydb host=localhost'");
  });

  it("builds a mysql catalog clause using the database= key", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "mysql",
        catalog_mysql_database: "mydb",
        catalog_mysql_host: "localhost",
        catalog_mysql_port: "3306",
        catalog_mysql_user: "root",
        catalog_mysql_password: "secret",
      }),
    ).toBe(
      "'ducklake:mysql:database=mydb host=localhost port=3306 user=root password=secret'",
    );
  });

  it("includes an alias when provided", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "duckdb_database.ducklake",
        alias: "my_ducklake",
      }),
    ).toBe("'ducklake:duckdb_database.ducklake' AS my_ducklake");
  });

  it("appends DATA_PATH based on data_path_type", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "duckdb_database.ducklake",
        data_path_type: "local",
        data_path_local: "other_data_path/",
      }),
    ).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/')",
    );
  });

  it("picks the data path for the active storage type only", () => {
    // A stale s3 value should not leak into DATA_PATH when local is active.
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
        data_path_type: "local",
        data_path_local: "local/",
        data_path_s3: "s3://stale/",
      }),
    ).toBe("'ducklake:c.ducklake' (DATA_PATH 'local/')");

    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
        data_path_type: "s3",
        data_path_local: "local/",
        data_path_s3: "s3://bucket/",
      }),
    ).toBe("'ducklake:c.ducklake' (DATA_PATH 's3://bucket/')");
  });

  it("emits boolean options regardless of value", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
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
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
        override_data_path: false,
        encrypted: true,
      }),
    ).toBe("'ducklake:c.ducklake' (OVERRIDE_DATA_PATH false, ENCRYPTED true)");
  });

  it("emits METADATA_PARAMETERS without wrapping quotes", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
        metadata_parameters: "{foo: 'bar'}",
      }),
    ).toBe("'ducklake:c.ducklake' (METADATA_PARAMETERS {foo: 'bar'})");
  });

  it("escapes single quotes inside data path values", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
        data_path_type: "local",
        data_path_local: "path/with'quote",
      }),
    ).toBe("'ducklake:c.ducklake' (DATA_PATH 'path/with''quote')");
  });

  it("composes the example from the DuckLake docs", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "duckdb_database.ducklake",
        data_path_type: "local",
        data_path_local: "other_data_path/",
        override_data_path: true,
      }),
    ).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/', OVERRIDE_DATA_PATH true)",
    );
  });

  it("emits every advanced parameter when set", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
        alias: "my_ducklake",
        data_path_type: "s3",
        data_path_s3: "s3://bucket/",
        override_data_path: false,
        create_if_not_exists: false,
        data_inlining_row_limit: 100,
        encrypted: true,

        meta_parameter_name: "foo",
        metadata_catalog: "meta_cat",
        metadata_parameters: "{a: 1}",
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
      catalog_type: "duckdb",
      catalog_duckdb_path: "duckdb_database.ducklake",
      data_path_type: "local",
      data_path_local: "other_data_path/",
    });
    expect(result.attach).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/')",
    );
  });
});
