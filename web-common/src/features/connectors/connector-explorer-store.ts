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
    connector: string, // Included for API consistency, but not used in this function
    database?: string,
    schema?: string,
  ): boolean {
    if (schema) return false; // Database Schema
    if (database) return true; // Database
    return true; // Connector
  }

  function createItemIfNotExists(
    connector: string,
    database?: string,
    schema?: string,
  ) {
    update((state) => {
      const key = getItemKey(connector, database, schema);

      if (key in state.expandedItems) return state; // Item already exists

      return {
        ...state,
        expandedItems: {
          ...state.expandedItems,
          [key]: getDefaultState(connector, database, schema),
        },
      };
    });
  }

  return {
    subscribe,
    toggleExplorer: () =>
      update((state) => ({ ...state, showConnectors: !state.showConnectors })),

    getItem: (
      connector: string,
      database?: string,
      schema?: string,
    ): Readable<boolean> => {
      createItemIfNotExists(connector, database, schema);

      const key = getItemKey(connector, database, schema);

      return derived({ subscribe }, ($state) => {
        return $state.expandedItems[key];
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

    // Not used yet. Currently, the reconciler does not track connector renames.
    renameItem: (
      oldConnector: string,
      newConnector: string,
      oldDatabase?: string,
      newDatabase?: string,
      oldSchema?: string,
      newSchema?: string,
    ) =>
      update((state) => {
        const oldKeyPrefix = getItemKey(oldConnector, oldDatabase, oldSchema);
        const newKeyPrefix = getItemKey(newConnector, newDatabase, newSchema);

        const updatedExpandedItems = Object.fromEntries(
          Object.entries(state.expandedItems).map(([key, value]) => {
            if (key.startsWith(oldKeyPrefix)) {
              const newKey = key.replace(oldKeyPrefix, newKeyPrefix);
              return [newKey, value];
            }
            return [key, value];
          }),
        );

        return {
          ...state,
          expandedItems: updatedExpandedItems,
        };
      }),

    deleteItem: (connector: string, database?: string, schema?: string) =>
      update((state) => {
        const keyPrefix = getItemKey(connector, database, schema);
        const updatedExpandedItems = Object.fromEntries(
          Object.entries(state.expandedItems).filter(
            ([key]) => !key.startsWith(keyPrefix),
          ),
        );
        return {
          ...state,
          expandedItems: updatedExpandedItems,
        };
      }),
  };
}

export const connectorExplorerStore = createConnectorExplorerStore();
