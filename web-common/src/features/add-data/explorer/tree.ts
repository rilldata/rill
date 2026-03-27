import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  getConnectorServiceListDatabaseSchemasQueryOptions,
  getConnectorServiceListTablesInfiniteQueryOptions,
  type V1AnalyzedConnector,
  type V1ListTablesResponse,
} from "@rilldata/web-common/runtime-client";
import {
  createQuery,
  createInfiniteQuery,
  type QueryClient,
  type InfiniteData,
} from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import { extractErrorMessage } from "@rilldata/web-common/lib/errors.ts";

export type ConnectorExplorerEntry = {
  connector: string;
  database?: string;
  databaseSchema?: string;
  table?: string;
};
export type ConnectorExplorerNode = {
  name: string;
  type: ConnectorExplorerNodeType;
  entry: ConnectorExplorerEntry;
  children?: ConnectorExplorerNode[];
  loading?: boolean;
  error?: string;
};
export enum ConnectorExplorerNodeType {
  Database,
  Schema,
  Table,
}

const TablesPageSize = 100;

export function filterConnectorExplorerTree(
  runtimeClient: RuntimeClient,
  queryClient: QueryClient,
  connector: V1AnalyzedConnector | undefined,
  searchTextStore: Readable<string>,
) {
  return derived(
    [
      getConnectorExplorerTree(runtimeClient, queryClient, connector),
      searchTextStore,
    ],
    ([connectorExplorerTree, searchText]) => {
      const getFilteredNode = (
        node: ConnectorExplorerNode,
      ): ConnectorExplorerNode | null => {
        const nameMatches = node.name
          .toLowerCase()
          .includes(searchText.toLowerCase());
        const filteredChildren = node.children
          ?.map(getFilteredNode)
          .filter(Boolean) as ConnectorExplorerNode[] | undefined;

        if (!nameMatches && !filteredChildren?.length) {
          return null;
        }

        return {
          ...node,
          children: filteredChildren,
        };
      };

      return {
        data: connectorExplorerTree.data
          ?.map(getFilteredNode)
          .filter(Boolean) as ConnectorExplorerNode[],
        loading: connectorExplorerTree.loading,
        error: connectorExplorerTree.error,
      };
    },
  );
}

export function getConnectorExplorerTree(
  runtimeClient: RuntimeClient,
  queryClient: QueryClient,
  connector: V1AnalyzedConnector | undefined,
): Readable<{
  data?: ConnectorExplorerNode[];
  loading?: boolean;
  error?: string;
}> {
  // Top level query to get all database schemas and tables.
  const databaseSchemasQuery = createQuery(
    getConnectorServiceListDatabaseSchemasQueryOptions(
      runtimeClient,
      {
        connector: connector?.name,
      },
      {
        query: { enabled: !!connector?.name && !connector?.errorMessage },
      },
    ),
    queryClient,
  );

  return derived(databaseSchemasQuery, (databaseSchemas, set) => {
    const connectorName = connector?.name;
    // Early return if the query is still pending or has an error.
    if (databaseSchemas.isPending || databaseSchemas.error || !connectorName) {
      set({
        loading: databaseSchemas.isPending,
        error: databaseSchemas.error
          ? extractErrorMessage(databaseSchemas.error)
          : undefined,
      });
      return;
    }

    // Collect schema entries and corresponding table query options in parallel
    // arrays so the index links each query back to its schema entry.
    type SchemaEntry = {
      name: string;
      nodeType: ConnectorExplorerNodeType;
      entry: ConnectorExplorerEntry;
      tableQueryOption: ReturnType<
        typeof getConnectorServiceListTablesInfiniteQueryOptions
      >;
    };
    const schemaEntries: SchemaEntry[] =
      databaseSchemas?.data?.databaseSchemas?.map(
        ({ database, databaseSchema }) => {
          const tableQueryOption =
            getConnectorServiceListTablesInfiniteQueryOptions(runtimeClient, {
              connector: connectorName,
              database,
              databaseSchema,
              pageSize: TablesPageSize,
            });
          return {
            name: databaseSchema ?? "",
            nodeType: !database
              ? ConnectorExplorerNodeType.Database
              : ConnectorExplorerNodeType.Schema,
            entry: { connector: connectorName, database, databaseSchema },
            tableQueryOption,
          };
        },
      ) ?? [];

    const tableQueriesStore = derived(
      schemaEntries.map((e) =>
        createInfiniteQuery(e.tableQueryOption, queryClient),
      ),
      (tableQueries) => {
        // Build schema nodes with their children from query results
        const schemaNodes = schemaEntries.map(
          ({ name, nodeType, entry }, i) => {
            const q = tableQueries[i];
            const tablesInfo =
              (
                q.data as InfiniteData<V1ListTablesResponse, string | undefined>
              )?.pages.flatMap((p) => p.tables ?? []) ?? [];
            return {
              name,
              type: nodeType,
              entry,
              children: tablesInfo.map(
                (t) =>
                  ({
                    name: t.name ?? "",
                    type: ConnectorExplorerNodeType.Table,
                    entry: { ...entry, table: t.name },
                    loading: q.isPending,
                    error: q.error ? extractErrorMessage(q.error) : undefined,
                  }) satisfies ConnectorExplorerNode,
              ),
            } satisfies ConnectorExplorerNode;
          },
        );

        // Group schema nodes under their parent database nodes.
        const nodesByDatabase: Record<string, ConnectorExplorerNode[]> = {};
        const directSchemaNodes: ConnectorExplorerNode[] = [];
        schemaNodes.forEach((node) => {
          if (!node.entry.database) {
            directSchemaNodes.push(node);
          } else {
            nodesByDatabase[node.entry.database] ??= [];
            nodesByDatabase[node.entry.database].push(node);
          }
        });

        const databaseNodes = Object.entries(nodesByDatabase).map(
          ([database, nodes]) =>
            ({
              name: database,
              type: ConnectorExplorerNodeType.Database,
              children: nodes,
              entry: { connector: connectorName, database },
            }) satisfies ConnectorExplorerNode,
        );

        return [...directSchemaNodes, ...databaseNodes];
      },
    );

    return tableQueriesStore.subscribe((nodes) => set({ data: nodes }));
  });
}
