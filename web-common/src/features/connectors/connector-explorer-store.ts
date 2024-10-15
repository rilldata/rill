import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";

type ConnectorExplorerState = {
  showConnectors: boolean;
  expandedItems: Record<string, boolean>;
};

type TableInfo = {
  connector: string;
  database?: string;
  schema: string;
  table: string;
};

const init: ConnectorExplorerState = {
  showConnectors: true,
  expandedItems: {},
};

export class ConnectorExplorerStore {
  allowNavigateToTable: boolean;
  allowContextMenu: boolean;
  allowSelectTable: boolean;
  allowShowSchema: boolean;
  selectedTable: Writable<TableInfo | null>;
  store: Writable<ConnectorExplorerState>;

  constructor({
    allowNavigateToTable = true,
    allowContextMenu = true,
    allowShowSchema = true,
    allowSelectTable = false,
    selectedTable = null,
    showConnectors = true,
    expandedItems = {},
    localStorage = true,
  } = {}) {
    this.allowNavigateToTable = allowNavigateToTable;
    this.allowContextMenu = allowContextMenu;
    this.allowShowSchema = allowShowSchema;
    this.allowSelectTable = allowSelectTable;
    this.selectedTable = writable(selectedTable);

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

  duplicateStore() {
    const state = get(this.store);
    return new ConnectorExplorerStore({
      allowNavigateToTable: false,
      allowContextMenu: false,
      allowShowSchema: false,
      allowSelectTable: true,
      selectedTable: null,
      showConnectors: state.showConnectors,
      expandedItems: state.expandedItems,
    });
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

  selectTable = (table: TableInfo) => {
    this.selectedTable.set(table);
  };

  toggleItem = (
    connector: string,
    database?: string,
    schema?: string,
    table?: string,
  ) => {
    if (table && schema && this.allowSelectTable) {
      this.selectTable({ connector, database, schema, table });
      return;
    }

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

// function createConnectorExplorerStore(
//   localStorage: boolean = true,
//   initialState: ConnectorExplorerState = init,
// ) {
//   const { subscribe, update } = localStorage
//     ? localStorageStore<ConnectorExplorerState>(
//         "connector-explorer-state",
//         initialState,
//       )
//     : writable(initialState);

//   function createItemIfNotExists(
//     connector: string,
//     database?: string,
//     schema?: string,
//   ) {
//     update((state) => {
//       const key = getItemKey(connector, database, schema);

//       if (key in state.expandedItems) return state; // Item already exists

//       return {
//         ...state,
//         expandedItems: {
//           ...state.expandedItems,
//           [key]: getDefaultState(connector, database, schema),
//         },
//       };
//     });

//   }

//   return {
//     subscribe,
//     toggleExplorer: () =>
//       update((state) => ({ ...state, showConnectors: !state.showConnectors })),

//     getItem: (
//       connector: string,
//       database?: string,
//       schema?: string,
//     ): Readable<boolean> => {
//       createItemIfNotExists(connector, database, schema);

//       const key = getItemKey(connector, database, schema);

//       return derived({ subscribe }, ($state) => {
//         return $state.expandedItems[key];
//       });
//     },

//     selectTable: (table: TableInfo) =>
//       update((state) => {
//         return {
//           ...state,
//           selectedTable: table,
//         };
//       }),

//     toggleItem: (connector: string, database?: string, schema?: string) =>
//       update((state) => {
//         const key = getItemKey(connector, database, schema);
//         const currentState =
//           state.expandedItems[key] ??
//           getDefaultState(connector, database, schema);
//         return {
//           ...state,
//           expandedItems: {
//             ...state.expandedItems,
//             [key]: !currentState,
//           },
//         };
//       }),

//     // Not used yet. Currently, the reconciler does not track connector renames.
//     renameItem: (
//       oldConnector: string,
//       newConnector: string,
//       oldDatabase?: string,
//       newDatabase?: string,
//       oldSchema?: string,
//       newSchema?: string,
//     ) =>
//       update((state) => {
//         const oldKeyPrefix = getItemKey(oldConnector, oldDatabase, oldSchema);
//         const newKeyPrefix = getItemKey(newConnector, newDatabase, newSchema);

//         const updatedExpandedItems = Object.fromEntries(
//           Object.entries(state.expandedItems).map(([key, value]) => {
//             if (key.startsWith(oldKeyPrefix)) {
//               const newKey = key.replace(oldKeyPrefix, newKeyPrefix);
//               return [newKey, value];
//             }
//             return [key, value];
//           }),
//         );

//         return {
//           ...state,
//           expandedItems: updatedExpandedItems,
//         };
//       }),

//     deleteItem: (connector: string, database?: string, schema?: string) =>
//       update((state) => {
//         const keyPrefix = getItemKey(connector, database, schema);
//         const updatedExpandedItems = Object.fromEntries(
//           Object.entries(state.expandedItems).filter(
//             ([key]) => !key.startsWith(keyPrefix),
//           ),
//         );
//         return {
//           ...state,
//           expandedItems: updatedExpandedItems,
//         };
//       }),
//   };
// }

export const connectorExplorerStore = new ConnectorExplorerStore();

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
