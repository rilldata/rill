import { describe, it, expect } from "vitest";
import {
  applyDuckLakeFormTransform,
  buildDuckLakeSecretRefs,
  composeDuckLakeAttach,
  extractDuckLakeAttachSecrets,
  shouldExtractDuckLakeAttachSecrets,
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

  it("substitutes postgres password with an env template reference when secretRefs is provided", () => {
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "postgres",
          catalog_postgres_dbname: "mydb",
          catalog_postgres_host: "localhost",
          catalog_postgres_user: "postgres",
          catalog_postgres_password: "secret",
        },
        {
          secretRefs: {
            catalog_postgres_password: "DUCKLAKE_CATALOG_POSTGRES_PASSWORD",
          },
        },
      ),
    ).toBe(
      "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD }}'",
    );
  });

  it("substitutes mysql password with an env template reference when secretRefs is provided", () => {
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "mysql",
          catalog_mysql_database: "mydb",
          catalog_mysql_host: "localhost",
          catalog_mysql_user: "root",
          catalog_mysql_password: "secret",
        },
        {
          secretRefs: {
            catalog_mysql_password: "DUCKLAKE_CATALOG_MYSQL_PASSWORD",
          },
        },
      ),
    ).toBe(
      "'ducklake:mysql:database=mydb host=localhost user=root password={{ .env.DUCKLAKE_CATALOG_MYSQL_PASSWORD }}'",
    );
  });

  it("omits the password pair when secretRefs is provided but the password is empty", () => {
    expect(
      composeDuckLakeAttach(
        {
          catalog_type: "postgres",
          catalog_postgres_dbname: "mydb",
          catalog_postgres_host: "localhost",
          catalog_postgres_password: "",
        },
        {
          secretRefs: {
            catalog_postgres_password: "DUCKLAKE_CATALOG_POSTGRES_PASSWORD",
          },
        },
      ),
    ).toBe("'ducklake:postgres:dbname=mydb host=localhost'");
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
        "READ_ONLY true, " +
        "OVERRIDE_DATA_PATH false, " +
        "CREATE_IF_NOT_EXISTS false, " +
        "ENCRYPTED true, " +
        "AUTOMATIC_MIGRATION true" +
        ")",
    );
  });

  it("emits READ_ONLY from the mode toggle in both states", () => {
    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
        mode: true,
      }),
    ).toBe("'ducklake:c.ducklake' (READ_ONLY true)");

    expect(
      composeDuckLakeAttach({
        catalog_type: "duckdb",
        catalog_duckdb_path: "c.ducklake",
        mode: false,
      }),
    ).toBe("'ducklake:c.ducklake' (READ_ONLY false)");
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

  it("threads secretRefs through to composeDuckLakeAttach", () => {
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
      {
        secretRefs: {
          catalog_postgres_password: "DUCKLAKE_CATALOG_POSTGRES_PASSWORD",
        },
      },
    );
    expect(result.attach).toBe(
      "'ducklake:postgres:dbname=mydb host=localhost user=postgres password={{ .env.DUCKLAKE_CATALOG_POSTGRES_PASSWORD }}'",
    );
  });
});

describe("buildDuckLakeSecretRefs", () => {
  it("returns an empty object for non-DuckLake schemas", () => {
    expect(buildDuckLakeSecretRefs(null, "postgres", "")).toEqual({});
    expect(
      buildDuckLakeSecretRefs(
        { type: "object", title: "Other", properties: {} },
        "duckdb",
        "",
      ),
    ).toEqual({});
  });

  it("resolves env var names for the DuckLake password fields", () => {
    const refs = buildDuckLakeSecretRefs(ducklakeSchema, "duckdb", "");
    expect(refs).toEqual({
      catalog_postgres_password: "DUCKLAKE_CATALOG_POSTGRES_PASSWORD",
      catalog_mysql_password: "DUCKLAKE_CATALOG_MYSQL_PASSWORD",
    });
  });

  it("suffixes env var names to avoid conflicts in an existing .env blob", () => {
    const envBlob =
      "DUCKLAKE_CATALOG_POSTGRES_PASSWORD=existing\n" +
      "DUCKLAKE_CATALOG_MYSQL_PASSWORD=existing";
    const refs = buildDuckLakeSecretRefs(ducklakeSchema, "duckdb", envBlob);
    expect(refs).toEqual({
      catalog_postgres_password: "DUCKLAKE_CATALOG_POSTGRES_PASSWORD_1",
      catalog_mysql_password: "DUCKLAKE_CATALOG_MYSQL_PASSWORD_1",
    });
  });
});

describe("extractDuckLakeAttachSecrets", () => {
  it("returns empty extraction for empty input", () => {
    expect(extractDuckLakeAttachSecrets("", "")).toEqual({
      rewrittenAttach: "",
      extractedSecrets: {},
    });
  });

  it("leaves non-credential catalog URIs alone", () => {
    const attach =
      "'ducklake:catalog.ducklake' AS my_ducklake (DATA_PATH 'files/')";
    const result = extractDuckLakeAttachSecrets(attach, "");
    expect(result.rewrittenAttach).toBe(attach);
    expect(result.extractedSecrets).toEqual({});

    const sqliteAttach = "'ducklake:sqlite:catalog.sqlite' AS x";
    const sqliteResult = extractDuckLakeAttachSecrets(sqliteAttach, "");
    expect(sqliteResult.rewrittenAttach).toBe(sqliteAttach);
    expect(sqliteResult.extractedSecrets).toEqual({});
  });

  it("extracts a postgres catalog body into DUCKLAKE_POSTGRES", () => {
    const attach =
      "'ducklake:postgres:dbname=mydb host=localhost password=secret' AS my_ducklake (DATA_PATH 'files/')";
    const result = extractDuckLakeAttachSecrets(attach, "");
    expect(result.rewrittenAttach).toBe(
      "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES }}' AS my_ducklake (DATA_PATH 'files/')",
    );
    expect(result.extractedSecrets).toEqual({
      DUCKLAKE_POSTGRES: "dbname=mydb host=localhost password=secret",
    });
  });

  it("extracts mysql and motherduck catalog bodies", () => {
    const mysql = extractDuckLakeAttachSecrets(
      "'ducklake:mysql:database=x host=y password=z'",
      "",
    );
    expect(mysql.extractedSecrets).toEqual({
      DUCKLAKE_MYSQL: "database=x host=y password=z",
    });
    expect(mysql.rewrittenAttach).toBe(
      "'ducklake:mysql:{{ .env.DUCKLAKE_MYSQL }}'",
    );

    const md = extractDuckLakeAttachSecrets(
      "'ducklake:md:my_db?motherduck_token=abc'",
      "",
    );
    expect(md.extractedSecrets).toEqual({
      DUCKLAKE_MOTHERDUCK: "my_db?motherduck_token=abc",
    });
    expect(md.rewrittenAttach).toBe(
      "'ducklake:md:{{ .env.DUCKLAKE_MOTHERDUCK }}'",
    );
  });

  it("suffixes when the base env var is already defined", () => {
    const envBlob = "DUCKLAKE_POSTGRES=existing";
    const result = extractDuckLakeAttachSecrets(
      "'ducklake:postgres:dbname=mydb' AS x",
      envBlob,
    );
    expect(result.extractedSecrets).toEqual({
      DUCKLAKE_POSTGRES_1: "dbname=mydb",
    });
    expect(result.rewrittenAttach).toBe(
      "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES_1 }}' AS x",
    );
  });

  it("is idempotent when the body is already a single env template", () => {
    const attach =
      "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES }}' AS my_ducklake";
    const result = extractDuckLakeAttachSecrets(attach, "");
    expect(result.rewrittenAttach).toBe(attach);
    expect(result.extractedSecrets).toEqual({});
  });

  it("allocates distinct env vars when the same driver appears twice", () => {
    const attach =
      "'ducklake:postgres:dbname=a password=1' vs 'ducklake:postgres:dbname=b password=2'";
    const result = extractDuckLakeAttachSecrets(attach, "");
    expect(result.extractedSecrets).toEqual({
      DUCKLAKE_POSTGRES: "dbname=a password=1",
      DUCKLAKE_POSTGRES_1: "dbname=b password=2",
    });
    expect(result.rewrittenAttach).toBe(
      "'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES }}' vs 'ducklake:postgres:{{ .env.DUCKLAKE_POSTGRES_1 }}'",
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
