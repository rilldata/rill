import { localStorageStore } from "@rilldata/web-common/lib/store-utils";

interface ConnectorExplorerState {
  showConnectors: boolean;
  connectors: Record<
    string,
    {
      showConnector: boolean;
      databases: Record<
        string,
        {
          showDatabase: boolean;
          databaseSchemas: Record<
            string,
            {
              showDatabaseSchema: boolean;
            }
          >;
        }
      >;
    }
  >;
}

const initialState: ConnectorExplorerState = {
  showConnectors: true,
  connectors: {},
};

function createConnectorExplorerStore() {
  const { subscribe, update } = localStorageStore<ConnectorExplorerState>(
    "connector-explorer-state",
    initialState,
  );

  return {
    subscribe,
    toggleExplorer: () =>
      update((state) => ({ ...state, showConnectors: !state.showConnectors })),
    toggleConnector: (connectorName: string) =>
      update((state) => {
        const connector = state.connectors[connectorName] || {
          showConnector: true,
          databases: {},
        };
        return {
          ...state,
          connectors: {
            ...state.connectors,
            [connectorName]: {
              ...connector,
              showConnector: !connector.showConnector,
            },
          },
        };
      }),
    toggleDatabase: (connectorName: string, databaseName: string) =>
      update((state) => {
        const connector = state.connectors[connectorName] || {
          showConnector: true,
          databases: {},
        };
        const database = connector.databases[databaseName] || {
          showDatabase: true,
          databaseSchemas: {},
        };
        return {
          ...state,
          connectors: {
            ...state.connectors,
            [connectorName]: {
              ...connector,
              databases: {
                ...connector.databases,
                [databaseName]: {
                  ...database,
                  showDatabase: !database.showDatabase,
                },
              },
            },
          },
        };
      }),
    toggleSchema: (
      connectorName: string,
      databaseName: string,
      schemaName: string,
    ) =>
      update((state) => {
        const connector = state.connectors[connectorName] || {
          showConnector: true,
          databases: {},
        };
        const database = connector.databases[databaseName] || {
          showDatabase: true,
          databaseSchemas: {},
        };
        const schema = database.databaseSchemas[schemaName] || {
          showDatabaseSchema: true,
        };
        return {
          ...state,
          connectors: {
            ...state.connectors,
            [connectorName]: {
              ...connector,
              databases: {
                ...connector.databases,
                [databaseName]: {
                  ...database,
                  databaseSchemas: {
                    ...database.databaseSchemas,
                    [schemaName]: {
                      ...schema,
                      showDatabaseSchema: !schema.showDatabaseSchema,
                    },
                  },
                },
              },
            },
          },
        };
      }),
  };
}

export const connectorExplorerStore = createConnectorExplorerStore();
