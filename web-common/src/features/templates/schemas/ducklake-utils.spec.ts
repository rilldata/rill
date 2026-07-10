import { describe, it, expect } from "vitest";
import {
  applyDuckLakeFormTransform,
  composeDuckLakeAttach,
  extractDuckLakeAttachSecrets,
  shouldExtractDuckLakeAttachSecrets,
  validateDuckLakeAttach,
} from "./ducklake-utils";
import { ducklakeSchema } from "./ducklake";
import {
  envMappedVarsAndValuesToObject,
  makeTestEnvEditSession,
} from "@rilldata/web-common/features/env-management/test/test-env-store.ts";

describe("composeDuckLakeAttach", () => {
  it("returns empty string when catalog identifier is missing", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(composeDuckLakeAttach({}, envEditSession)).toBe("");
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "   ",
        },
        envEditSession,
      ),
    ).toBe("");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("defaults to duckdb catalog when catalog_type is unset", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        { catalog_duckdb_path: "catalog.ducklake" },
        envEditSession,
      ),
    ).toBe("'ducklake:catalog.ducklake'");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("builds a minimal clause from a duckdb catalog path", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "duckdb_database.ducklake",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:duckdb_database.ducklake'");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("builds a sqlite catalog clause with the sqlite: prefix", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "sqlite",
          catalog_sqlite_path: "catalog.sqlite",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:sqlite:catalog.sqlite'");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("omits missing postgres params when composing the connection string", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "postgres",
          catalog_postgres_dbname: "mydb",
          catalog_postgres_host: "localhost",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:postgres:dbname=mydb host=localhost'");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("builds a mysql catalog clause using the database= key", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "mysql",
          catalog_mysql_database: "mydb",
          catalog_mysql_host: "localhost",
          catalog_mysql_port: "3306",
          catalog_mysql_user: "root",
          catalog_mysql_password: "secret",
        },
        envEditSession,
      ),
    ).toBe(
      "'ducklake:mysql:database=mydb host=localhost port=3306 user=root password={{ .env.DUCKLAKE_CATALOG_MYSQL_PASSWORD }}'",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      DUCKLAKE_CATALOG_MYSQL_PASSWORD: "secret",
    });
  });

  it("substitutes postgres password with an env template reference", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "postgres",
          catalog_postgres_dbname: "mydb",
          catalog_postgres_host: "localhost",
          catalog_postgres_user: "postgres",
          catalog_postgres_password: "secret",
        },
        envEditSession,
      ),
    ).toBe(
      "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD }}'",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      DUCKLAKE_CATALOG_POSTGRES_PASSWORD: "secret",
    });
  });

  it("substitutes mysql password with an env template reference", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "mysql",
          catalog_mysql_database: "mydb",
          catalog_mysql_host: "localhost",
          catalog_mysql_user: "root",
          catalog_mysql_password: "secret",
        },
        envEditSession,
      ),
    ).toBe(
      "'ducklake:mysql:database=mydb host=localhost user=root password={{ .env.DUCKLAKE_CATALOG_MYSQL_PASSWORD }}'",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      DUCKLAKE_CATALOG_MYSQL_PASSWORD: "secret",
    });
  });

  it("omits the password pair when secretRefs is provided but the password is empty", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "postgres",
          catalog_postgres_dbname: "mydb",
          catalog_postgres_host: "localhost",
          catalog_postgres_password: "",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:postgres:dbname=mydb host=localhost'");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("includes an alias when provided", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "duckdb_database.ducklake",
          alias: "my_ducklake",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:duckdb_database.ducklake' AS my_ducklake");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("appends DATA_PATH based on data_path_type", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "duckdb_database.ducklake",
          data_path_type: "local",
          data_path_local: "other_data_path/",
        },
        envEditSession,
      ),
    ).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/')",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("picks the data path for the active storage type only", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    // A stale s3 value should not leak into DATA_PATH when local is active.
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          data_path_type: "local",
          data_path_local: "local/",
          data_path_s3: "s3://stale/",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:c.ducklake' (DATA_PATH 'local/')");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});

    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          data_path_type: "s3",
          data_path_local: "local/",
          data_path_s3: "s3://bucket/",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:c.ducklake' (DATA_PATH 's3://bucket/')");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("emits boolean options regardless of value", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          override_data_path: true,
          create_if_not_exists: true,
        },
        envEditSession,
      ),
    ).toBe(
      "'ducklake:c.ducklake' (OVERRIDE_DATA_PATH true, CREATE_IF_NOT_EXISTS true)",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("emits false and true boolean options", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          override_data_path: false,
          encrypted: true,
        },
        envEditSession,
      ),
    ).toBe("'ducklake:c.ducklake' (OVERRIDE_DATA_PATH false, ENCRYPTED true)");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("emits METADATA_PARAMETERS without wrapping quotes", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          metadata_parameters: "{foo: 'bar'}",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:c.ducklake' (METADATA_PARAMETERS {foo: 'bar'})");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("escapes single quotes inside data path values", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          data_path_type: "local",
          data_path_local: "path/with'quote",
        },
        envEditSession,
      ),
    ).toBe("'ducklake:c.ducklake' (DATA_PATH 'path/with''quote')");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("composes the example from the DuckLake docs", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "duckdb_database.ducklake",
          data_path_type: "local",
          data_path_local: "other_data_path/",
          override_data_path: true,
        },
        envEditSession,
      ),
    ).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/', OVERRIDE_DATA_PATH true)",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("emits every advanced parameter when set", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          alias: "my_ducklake",
          data_path_type: "s3",
          data_path_s3: "s3://bucket/",
          mode: true,
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
        },
        envEditSession,
      ),
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
        "READ_ONLY true, " +
        "OVERRIDE_DATA_PATH false, " +
        "CREATE_IF_NOT_EXISTS false, " +
        "ENCRYPTED true, " +
        "AUTOMATIC_MIGRATION true" +
        ")",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("emits READ_ONLY from the mode toggle in both states", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          mode: true,
        },
        envEditSession,
      ),
    ).toBe("'ducklake:c.ducklake' (READ_ONLY true)");

    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "duckdb",
          catalog_duckdb_path: "c.ducklake",
          mode: false,
        },
        envEditSession,
      ),
    ).toBe("'ducklake:c.ducklake' (READ_ONLY false)");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });
});

describe("applyDuckLakeFormTransform", () => {
  it("returns values unchanged for non-DuckLake schemas", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    const values = { attach: "foo" };
    expect(applyDuckLakeFormTransform(null, values, envEditSession)).toBe(
      values,
    );
    expect(
      applyDuckLakeFormTransform(
        { type: "object", title: "DuckLake", properties: {} },
        values,
        envEditSession,
      ),
    ).toBe(values);
  });

  it("leaves attach alone when in SQL mode", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    const values = { connection_mode: "sql", attach: "user input" };
    expect(
      applyDuckLakeFormTransform(ducklakeSchema, values, envEditSession),
    ).toBe(values);
  });

  it("synthesises attach from params when in parameters mode", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    const result = applyDuckLakeFormTransform(
      ducklakeSchema,
      {
        connection_mode: "parameters",
        catalog_type: "duckdb",
        catalog_duckdb_path: "duckdb_database.ducklake",
        data_path_type: "local",
        data_path_local: "other_data_path/",
      },
      envEditSession,
    );
    expect(result.attach).toBe(
      "'ducklake:duckdb_database.ducklake' (DATA_PATH 'other_data_path/')",
    );
  });

  it("threads secretRefs through to composeDuckLakeAttach", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    const result = applyDuckLakeFormTransform(
      ducklakeSchema,
      {
        connection_mode: "parameters",
        catalog_type: "postgres",
        catalog_postgres_dbname: "mydb",
        catalog_postgres_host: "localhost",
        catalog_postgres_user: "postgres",
        catalog_postgres_password: "secret",
      },
      envEditSession,
    );
    expect(result.attach).toBe(
      "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD }}'",
    );
  });
});

describe("extractDuckLakeAttachSecrets", () => {
  it("returns empty extraction for empty input", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    expect(extractDuckLakeAttachSecrets("", envEditSession)).toEqual("");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("leaves non-credential catalog URIs alone", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    const attach =
      "'ducklake:catalog.ducklake' AS my_ducklake (DATA_PATH 'files/')";
    const rewrittenAttach = extractDuckLakeAttachSecrets(
      attach,
      envEditSession,
    );
    expect(rewrittenAttach).toBe(attach);
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});

    const sqliteAttach = "'ducklake:sqlite:catalog.sqlite' AS x";
    const rewrittenSqliteAttach = extractDuckLakeAttachSecrets(
      sqliteAttach,
      envEditSession,
    );
    expect(rewrittenSqliteAttach).toBe(sqliteAttach);
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("extracts a postgres catalog body into DUCKLAKE_POSTGRES", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    const attach =
      "'ducklake:postgres:dbname=mydb host=localhost password=secret' AS my_ducklake (DATA_PATH 'files/')";
    const rewrittenAttach = extractDuckLakeAttachSecrets(
      attach,
      envEditSession,
    );
    expect(rewrittenAttach).toBe(
      "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES }}' AS my_ducklake (DATA_PATH 'files/')",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      DUCKLAKE_POSTGRES: "dbname=mydb host=localhost password=secret",
    });
  });

  it("extracts mysql and motherduck catalog bodies", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );

    envEditSession.startEdit();
    const mysqlAttach = extractDuckLakeAttachSecrets(
      "'ducklake:mysql:database=x host=y password=z'",
      envEditSession,
    );
    expect(mysqlAttach).toBe("'ducklake:mysql:{{ .env.DUCKLAKE_MYSQL }}'");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      DUCKLAKE_MYSQL: "database=x host=y password=z",
    });

    envEditSession.startEdit();
    const mdAttach = extractDuckLakeAttachSecrets(
      "'ducklake:md:my_db?motherduck_token=abc'",
      envEditSession,
    );
    expect(mdAttach).toEqual("'ducklake:md:{{ .env.DUCKLAKE_MOTHERDUCK }}'");
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      DUCKLAKE_MOTHERDUCK: "my_db?motherduck_token=abc",
    });
  });

  it("suffixes when the base env var is already defined", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
      {},
      {
        DUCKLAKE_POSTGRES: "existing",
      },
    );
    const rewrittenAttach = extractDuckLakeAttachSecrets(
      "'ducklake:postgres:dbname=mydb' AS x",
      envEditSession,
    );
    expect(rewrittenAttach).toBe(
      "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES_1 }}' AS x",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      DUCKLAKE_POSTGRES_1: "dbname=mydb",
    });
  });

  it("is idempotent when the body is already a single env template", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    const attach =
      "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES }}' AS my_ducklake";
    const rewrittenAttach = extractDuckLakeAttachSecrets(
      attach,
      envEditSession,
    );
    expect(rewrittenAttach).toBe(attach);
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({});
  });

  it("allocates distinct env vars when the same driver appears twice", async () => {
    const { envEditSession } = await makeTestEnvEditSession(
      "ducklake",
      undefined,
    );
    const attach =
      "'ducklake:postgres:dbname=a password=1' vs 'ducklake:postgres:dbname=b password=2'";
    const rewrittenAttach = extractDuckLakeAttachSecrets(
      attach,
      envEditSession,
    );
    expect(rewrittenAttach).toBe(
      "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES }}' vs 'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES_1 }}'",
    );
    expect(envMappedVarsAndValuesToObject(envEditSession.entries)).toEqual({
      DUCKLAKE_POSTGRES: "dbname=a password=1",
      DUCKLAKE_POSTGRES_1: "dbname=b password=2",
    });
  });
});

describe("validateDuckLakeAttach", () => {
  it("returns no errors for empty, whitespace, or non-string input", () => {
    expect(validateDuckLakeAttach("")).toEqual([]);
    expect(validateDuckLakeAttach("   ")).toEqual([]);
    expect(validateDuckLakeAttach(undefined)).toEqual([]);
    expect(validateDuckLakeAttach(null)).toEqual([]);
    expect(validateDuckLakeAttach(42)).toEqual([]);
  });

  it("accepts a well-formed duckdb catalog clause", () => {
    expect(
      validateDuckLakeAttach(
        "'ducklake:catalog.ducklake' AS my_ducklake (DATA_PATH 'files/')",
      ),
    ).toEqual([]);
  });

  it("accepts postgres, mysql, and motherduck schemes", () => {
    expect(
      validateDuckLakeAttach("'ducklake:postgres:dbname=x host=y'"),
    ).toEqual([]);
    expect(
      validateDuckLakeAttach("'ducklake:mysql:database=x host=y'"),
    ).toEqual([]);
    expect(
      validateDuckLakeAttach("'ducklake:md:my_db?motherduck_token=abc'"),
    ).toEqual([]);
    expect(validateDuckLakeAttach("'ducklake:sqlite:catalog.sqlite'")).toEqual(
      [],
    );
  });

  it("strips a leading ATTACH wrapper before validating", () => {
    expect(
      validateDuckLakeAttach("ATTACH 'ducklake:catalog.ducklake'"),
    ).toEqual([]);
    expect(
      validateDuckLakeAttach(
        "ATTACH IF NOT EXISTS 'ducklake:catalog.ducklake' AS x (TYPE ducklake);",
      ),
    ).toEqual([]);
  });

  it("flags unbalanced single quotes", () => {
    const errors = validateDuckLakeAttach("'ducklake:catalog.ducklake");
    expect(errors).toContainEqual(
      expect.stringContaining("Unbalanced single quotes"),
    );
  });

  it("flags a missing ducklake: prefix", () => {
    const errors = validateDuckLakeAttach("'catalog.ducklake'");
    expect(errors).toContainEqual(
      expect.stringContaining("Catalog URI must begin with `ducklake:`"),
    );
  });

  it("accepts the alternative (TYPE DUCKLAKE) form without the ducklake: prefix", () => {
    expect(
      validateDuckLakeAttach(
        "'https://blobs.duckdb.org/datalake/tpch-sf3.ducklake' AS my_resource (TYPE ducklake)",
      ),
    ).toEqual([]);

    expect(
      validateDuckLakeAttach("'catalog.ducklake' AS x (TYPE DUCKLAKE)"),
    ).toEqual([]);
  });

  it("flags an empty body after the ducklake: prefix", () => {
    const errors = validateDuckLakeAttach("'ducklake:'");
    expect(errors).toContainEqual(
      expect.stringContaining("has no value after `ducklake:`"),
    );
  });

  it("flags an unknown catalog scheme", () => {
    const errors = validateDuckLakeAttach(
      "'ducklake:potsgres:dbname=x host=y'",
    );
    expect(errors).toContainEqual(
      expect.stringContaining("Unknown catalog scheme `potsgres:`"),
    );
  });

  it("flags an empty body after a known scheme prefix", () => {
    const errors = validateDuckLakeAttach("'ducklake:postgres:'");
    expect(errors).toContainEqual(
      expect.stringContaining("`postgres:` catalog has no body"),
    );
  });
});

describe("shouldExtractDuckLakeAttachSecrets", () => {
  it("only runs for the DuckLake schema", () => {
    expect(shouldExtractDuckLakeAttachSecrets(null, {})).toBe(false);
    expect(shouldExtractDuckLakeAttachSecrets(ducklakeSchema, {})).toBe(true);
  });

  it("skips parameters mode so the composer can emit env refs itself", () => {
    expect(
      shouldExtractDuckLakeAttachSecrets(ducklakeSchema, {
        connection_mode: "parameters",
      }),
    ).toBe(false);
    expect(
      shouldExtractDuckLakeAttachSecrets(ducklakeSchema, {
        connection_mode: "sql",
      }),
    ).toBe(true);
  });
});
