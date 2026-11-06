import type { V1Resource } from "@rilldata/web-common/runtime-client";

/**
 * Identifies the table that holds a resource's output.
 * Models and sources write to their connector's default database, so `database` and `databaseSchema` are always empty.
 */
export type ResultTable = {
  connector: string;
  database: string;
  databaseSchema: string;
  tableName: string;
};

/**
 * Resolves the result table of a resource referenced by a model.
 * A model's refs also include its connectors (and can include other kinds without a table), so a ref's name must not be queried as a table.
 * Returns undefined for resources that have no result table: connectors, metrics views, models that have not produced a result yet, and refs that have not loaded.
 */
export function getResultTable(
  resource: V1Resource | undefined,
): ResultTable | undefined {
  const model = resource?.model?.state;
  if (model?.resultConnector && model.resultTable) {
    return {
      connector: model.resultConnector,
      database: "",
      databaseSchema: "",
      tableName: model.resultTable,
    };
  }

  const source = resource?.source?.state;
  if (source?.connector && source.table) {
    return {
      connector: source.connector,
      database: "",
      databaseSchema: "",
      tableName: source.table,
    };
  }

  return undefined;
}
