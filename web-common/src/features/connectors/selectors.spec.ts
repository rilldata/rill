import { describe, expect, it, vi } from "vitest";
import type { RuntimeClient } from "../../runtime-client/v2";
import { fetchAllDatabaseSchemas, groupDatabaseSchemas } from "./selectors";

function makeClient(pages: Record<string, unknown>[]) {
  const listDatabaseSchemas = vi.fn(() => {
    const page = pages[listDatabaseSchemas.mock.calls.length - 1];
    return Promise.resolve({ toJson: () => page });
  });
  const client = {
    instanceId: "default",
    connectorService: { listDatabaseSchemas },
  } as unknown as RuntimeClient;
  return { client, listDatabaseSchemas };
}

describe("fetchAllDatabaseSchemas", () => {
  it("follows nextPageToken until exhausted", async () => {
    const { client, listDatabaseSchemas } = makeClient([
      {
        databaseSchemas: [{ database: "db1", databaseSchema: "s1" }],
        nextPageToken: "100",
      },
      {
        databaseSchemas: [{ database: "db1", databaseSchema: "s2" }],
        nextPageToken: "200",
      },
      {
        databaseSchemas: [{ database: "db2", databaseSchema: "s3" }],
        nextPageToken: "",
      },
    ]);

    const schemas = await fetchAllDatabaseSchemas(client, "snowflake");

    expect(schemas).toEqual([
      { database: "db1", databaseSchema: "s1" },
      { database: "db1", databaseSchema: "s2" },
      { database: "db2", databaseSchema: "s3" },
    ]);
    const tokens = listDatabaseSchemas.mock.calls.map(
      (c) => (c as unknown as [{ pageToken: string }])[0].pageToken,
    );
    expect(tokens).toEqual(["", "100", "200"]);
  });

  it("handles a single page without a token", async () => {
    const { client, listDatabaseSchemas } = makeClient([
      { databaseSchemas: [{ database: "db", databaseSchema: "main" }] },
    ]);

    const schemas = await fetchAllDatabaseSchemas(client, "duckdb");

    expect(schemas).toEqual([{ database: "db", databaseSchema: "main" }]);
    expect(listDatabaseSchemas).toHaveBeenCalledTimes(1);
  });
});

const schemas = [
  { database: "db1", databaseSchema: "s1" },
  { database: "db1", databaseSchema: "s2" },
  { database: "db2", databaseSchema: "s3" },
  { database: "db1", databaseSchema: "s4" },
];

describe("groupDatabaseSchemas", () => {
  it("derives unique databases", () => {
    expect(groupDatabaseSchemas(schemas)).toEqual(["db1", "db2"]);
  });

  it("lists schemas of a database across pages", () => {
    expect(groupDatabaseSchemas(schemas, "db1")).toEqual(["s1", "s2", "s4"]);
  });

  it("dedupes schemas repeated across pages", () => {
    const repeated = [...schemas, { database: "db1", databaseSchema: "s2" }];
    expect(groupDatabaseSchemas(repeated, "db1")).toEqual(["s1", "s2", "s4"]);
  });

  it("treats schemas as top level when there are no databases", () => {
    const flat = [
      { database: "", databaseSchema: "a" },
      { database: "", databaseSchema: "b" },
    ];
    expect(groupDatabaseSchemas(flat)).toEqual(["a", "b"]);
    expect(groupDatabaseSchemas(flat, "a")).toEqual(["a"]);
  });
});
