import type { QueryClient } from "@tanstack/svelte-query";
import {
  createSqlModelFromTable,
  createYamlModelFromTable,
} from "../../connectors/code-utils";

export async function createModelFromExplorerSelection(
  queryClient: QueryClient,
  options: {
    connector: string;
    database: string;
    schema: string;
    table: string;
    isModelingSupported: boolean;
  },
): Promise<[string, string]> {
  return options.isModelingSupported
    ? await createSqlModelFromTable(
        queryClient,
        options.connector,
        options.database,
        options.schema,
        options.table,
      )
    : await createYamlModelFromTable(
        queryClient,
        options.connector,
        options.database,
        options.schema,
        options.table,
      );
}
