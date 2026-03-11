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
    if (databaseSchemas.isPending || databaseSchemas.error || !connectorName) {
      set({
        loading: databaseSchemas.isPending,
        error: extractErrorMessage(databaseSchemas.error),
      });
      return;
    }

    const nodesByDatabase: Record<string, ConnectorExplorerNode[]> = {};
    const directSchemaNodes: ConnectorExplorerNode[] = [];

    const tableQueryOptions: ReturnType<
      typeof getConnectorServiceListTablesInfiniteQueryOptions
    >[] = [];
    const tableQueryParentNode: ConnectorExplorerNode[] = [];

    databaseSchemas?.data?.databaseSchemas?.forEach(
      ({ database, databaseSchema }) => {
        const node = {
          name: databaseSchema ?? "",
          type: !database
            ? ConnectorExplorerNodeType.Database
            : ConnectorExplorerNodeType.Schema,
          entry: {
            connector: connectorName,
            database,
            databaseSchema,
          },
        } satisfies ConnectorExplorerNode;

        if (!database) {
          directSchemaNodes.push(node);
        } else {
          nodesByDatabase[database] ??= [];
          nodesByDatabase[database].push(node);
        }

        tableQueryOptions.push(
          getConnectorServiceListTablesInfiniteQueryOptions(runtimeClient, {
            connector: connectorName,
            database,
            databaseSchema,
            pageSize: TablesPageSize,
          }),
        );
        tableQueryParentNode.push(node);
      },
    );

    const nodesForDatabases = Object.entries(nodesByDatabase).map(
      ([database, nodes]) =>
        ({
          name: database,
          type: ConnectorExplorerNodeType.Database,
          children: nodes,
          entry: {
            connector: connectorName,
            database,
          },
        }) satisfies ConnectorExplorerNode,
    );
    const allNodes = [...directSchemaNodes, ...nodesForDatabases];

    const tableQueriesStore = derived(
      tableQueryOptions.map((o) => createInfiniteQuery(o, queryClient)),
      (tableQueries) => {
        tableQueries.forEach((q, i) => {
          const tablesInfo =
            (
              q.data as InfiniteData<V1ListTablesResponse, string | undefined>
            )?.pages.flatMap((p) => p.tables ?? []) ?? [];
          tableQueryParentNode[i].children = tablesInfo.map(
            (t) =>
              ({
                name: t.name ?? "",
                type: ConnectorExplorerNodeType.Table,
                entry: {
                  ...tableQueryParentNode[i].entry,
                  table: t.name,
                },
                loading: q.isPending,
                error: q.error ? extractErrorMessage(q.error) : undefined,
              }) satisfies ConnectorExplorerNode,
          );
        });

        return allNodes;
      },
    );

    return tableQueriesStore.subscribe((nodes) => set({ data: nodes }));
  });
}
