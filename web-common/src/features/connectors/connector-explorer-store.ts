import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { derived, type Readable } from "svelte/store";

interface ConnectorExplorerState {
  showConnectors: boolean;
  expandedItems: Record<string, boolean>;
}

const initialState: ConnectorExplorerState = {
  showConnectors: true,
  expandedItems: {},
};

function createConnectorExplorerStore() {
  const { subscribe, update } = localStorageStore<ConnectorExplorerState>(
    "connector-explorer-state",
    initialState,
  );

  function getItemKey(
    connector: string,
    database?: string,
    schema?: string,
  ): string {
    return [connector, database, schema].filter(Boolean).join("|");
  }

  function getDefaultState(
    connector: string,
    database?: string,
    schema?: string,
  ): boolean {
    if (schema) return false; // Database Schema
    if (database) return true; // Database
    return true; // Connector
  }

  return {
    subscribe,
    toggleExplorer: () =>
      update((state) => ({ ...state, showConnectors: !state.showConnectors })),

    getItemState: (
      connector: string,
      database?: string,
      schema?: string,
    ): Readable<boolean> => {
      const key = getItemKey(connector, database, schema);
      return derived({ subscribe }, ($state) => {
        if (key in $state.expandedItems) {
          return $state.expandedItems[key];
        }
        return getDefaultState(connector, database, schema);
      });
    },

    toggleItem: (connector: string, database?: string, schema?: string) =>
      update((state) => {
        const key = getItemKey(connector, database, schema);
        const currentState =
          state.expandedItems[key] ??
          getDefaultState(connector, database, schema);
        return {
          ...state,
          expandedItems: {
            ...state.expandedItems,
            [key]: !currentState,
          },
        };
      }),
  };
}

export const connectorExplorerStore = createConnectorExplorerStore();
