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

  it("omits boolean options equal to their DuckLake default", () => {
    // OVERRIDE_DATA_PATH default is true, CREATE_IF_NOT_EXISTS default is true
    expect(
      composeDuckLakeAttach({
        catalog: "c.ducklake",
        override_data_path: true,
        create_if_not_exists: true,
      }),
    ).toBe("'ducklake:c.ducklake'");
  });

  it("includes boolean options that differ from the default", () => {
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
        override_data_path: true, // default, should still be implicit
      }),
    ).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/')",
    );
  });
});

describe("applyDuckLakeFormTransform", () => {
  it("returns values unchanged for non-DuckLake schemas", () => {
    const values = { attach: "foo" };
    expect(applyDuckLakeFormTransform(null, values)).toBe(values);
    expect(
      applyDuckLakeFormTransform(
        { type: "object", title: "Other", properties: {} },
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
