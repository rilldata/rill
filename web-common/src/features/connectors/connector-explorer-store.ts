import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { derived, get, writable, type Writable } from "svelte/store";

type ConnectorExplorerState = {
  showConnectors: boolean;
  expandedItems: Record<string, boolean>;
};

export class ConnectorExplorerStore {
  allowNavigateToTable: boolean;
  allowContextMenu: boolean;
  allowSelectTable: boolean;
  allowShowSchema: boolean;
  store: Writable<ConnectorExplorerState>;
  onToggleItem:
    | undefined
    | ((
        connector: string,
        database?: string,
        schema?: string,
        table?: string,
      ) => void) = undefined;

  constructor(
    {
      allowNavigateToTable = true,
      allowContextMenu = true,
      allowShowSchema = true,
      allowSelectTable = false,

      showConnectors = true,
      expandedItems = {},
      localStorage = true,
    } = {},
    onToggleItem?: (
      connector: string,
      database?: string,
      schema?: string,
      table?: string,
    ) => void,
  ) {
    this.allowNavigateToTable = allowNavigateToTable;
    this.allowContextMenu = allowContextMenu;
    this.allowShowSchema = allowShowSchema;
    this.allowSelectTable = allowSelectTable;

    if (onToggleItem) this.onToggleItem = onToggleItem;

    this.store = localStorage
      ? localStorageStore<ConnectorExplorerState>("connector-explorer-state", {
          showConnectors,
          expandedItems,
        })
      : writable({ showConnectors, expandedItems });
  }

  createItemIfNotExists(
    connector: string,
    database?: string,
    schema?: string,
    table?: string,
  ) {
    this.store.update((state) => {
      const key = getItemKey(connector, database, schema, table);

      if (key in state.expandedItems) return state; // Item already exists

      return {
        ...state,
        expandedItems: {
          ...state.expandedItems,
          [key]: getDefaultState(connector, database, schema, table),
        },
      };
    });
  }

  duplicateStore(
    onToggleItem?: (
      connector: string,
      database?: string,
      schema?: string,
      table?: string,
    ) => void | Promise<void>,
  ) {
    const state = get(this.store);
    return new ConnectorExplorerStore(
      {
        allowNavigateToTable: false,
        allowContextMenu: false,
        allowShowSchema: false,
        allowSelectTable: true,
        localStorage: false,
        showConnectors: state.showConnectors,
        expandedItems: {},
      },
      onToggleItem ?? this.onToggleItem,
    );
  }

  toggleExplorer = () =>
    this.store.update((state) => ({
      ...state,
      showConnectors: !state.showConnectors,
    }));

  getItem = (
    connector: string,
    database?: string,
    schema?: string,
    table?: string,
  ) => {
    this.createItemIfNotExists(connector, database, schema, table);

    const key = getItemKey(connector, database, schema, table);

    return derived(this.store, ($state) => {
      return $state.expandedItems[key];
    });
  };

  toggleItem = (
    connector: string,
    database?: string,
    schema?: string,
    table?: string,
  ) => {
    if (this.onToggleItem)
      this.onToggleItem(connector, database, schema, table);

    if (table && !this.allowShowSchema) return;

    this.store.update((state) => {
      const key = getItemKey(connector, database, schema, table);
      const currentState =
        state.expandedItems[key] ??
        getDefaultState(connector, database, schema, table);
      return {
        ...state,
        expandedItems: {
          ...state.expandedItems,
          [key]: !currentState,
        },
      };
    });
  };

  // Not used yet. Currently, the reconciler does not track connector renames.
  renameItem = (
    oldConnector: string,
    newConnector: string,
    oldDatabase?: string,
    newDatabase?: string,
    oldSchema?: string,
    newSchema?: string,
  ) =>
    this.store.update((state) => {
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
    });

  deleteItem = (connector: string, database?: string, schema?: string) =>
    this.store.update((state) => {
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
    });
}

export const connectorExplorerStore = new ConnectorExplorerStore();

// Helpers
function getItemKey(
  connector: string,
  database?: string,
  schema?: string,
  table?: string,
): string {
  return [connector, database, schema, table].filter(Boolean).join("|");
}

function getDefaultState(
  connector: string, // Included for API consistency, but not used in this function
  database?: string,
  schema?: string,
  table?: string,
): boolean {
  if (schema || table) return false; // Database Schema or Table
  if (database) return true; // Database
  return true; // Connector
}
