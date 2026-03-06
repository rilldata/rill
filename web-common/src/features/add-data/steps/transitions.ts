import {
  type AddDataConfig,
  type AddDataState,
  AddDataStep,
  type AddDataTransitionArgs,
  ImportDataStep,
} from "@rilldata/web-common/features/add-data/steps/types.ts";
import {
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  type V1AnalyzeConnectorsResponse,
  type V1ConnectorDriver,
} from "@rilldata/web-common/runtime-client";
import {
  connectors,
  getBackendConnectorName,
  getConnectorSchema,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

export function transitionToNextStep(
  runtimeClient: RuntimeClient,
  current: AddDataState,
  args: AddDataTransitionArgs,
): AddDataState {
  const selectedConnector: string | undefined = args.connector;
  let selectedSchema: string | undefined = args.schema;

  let driver: V1ConnectorDriver | null = null;
  if (selectedConnector) {
    const analyzedConnector = getConnectorDriverForConnector(
      runtimeClient,
      selectedConnector,
    );
    if (analyzedConnector) {
      driver = analyzedConnector.driver!;
      selectedSchema = analyzedConnector.name!;
    }
  }
  if (!driver && selectedSchema) {
    driver = getConnectorDriverForSchema(selectedSchema);
  }

  switch (current.step) {
    case AddDataStep.Select:
      if (selectedConnector) {
        return transitionFromConnector(
          driver!,
          selectedConnector,
          selectedSchema!,
        );
      } else if (selectedSchema) {
        return transitionFromSchema(driver!, selectedSchema);
      } else {
        return current;
      }

    case AddDataStep.Olap:
      return transitionToNextStep(
        runtimeClient,
        { step: AddDataStep.Select },
        { schema: current.schema },
      );

    case AddDataStep.Connector:
      if (!selectedConnector) {
        throw new Error("Connector must be specified");
      }

      return transitionFromConnector(
        driver!,
        selectedConnector,
        selectedSchema!,
      );

    case AddDataStep.Explorer:
    case AddDataStep.Source: {
      if (!args.importConfig) {
        throw new Error("Import config must be specified");
      }

      return {
        step: AddDataStep.Import,
        currentFilePath: "",
        importStep: { step: ImportDataStep.Init },
        config: args.importConfig,
      };
    }
  }

  return current;
}

export function maybeGetConnectorDriver(
  runtimeClient: RuntimeClient,
  schemaName: string | undefined,
  connectorName: string | undefined,
) {
  if (connectorName) {
    const analyzedConnector = getConnectorDriverForConnector(
      runtimeClient,
      connectorName,
    );
    return (
      analyzedConnector?.driver ?? getConnectorDriverForSchema(connectorName)
    );
  }
  if (schemaName) return getConnectorDriverForSchema(schemaName);
  return null;
}

export function getImportStepsForConnector(
  config: AddDataConfig,
  driver: V1ConnectorDriver,
) {
  // Live connectors cannot create models as of now.
  // They will create metrics views directly.
  const steps = isLiveConnectorType(driver)
    ? [ImportDataStep.CreateMetricsView, ImportDataStep.CreateExplore]
    : [
        ImportDataStep.CreateModel,
        ImportDataStep.CreateMetricsView,
        ImportDataStep.CreateExplore,
      ];
  return config.importOnly ? [steps[0]] : steps;
}

export function getImportStepsForSource(config: AddDataConfig) {
  return config.importOnly
    ? [ImportDataStep.CreateModel]
    : [
        ImportDataStep.CreateModel,
        ImportDataStep.CreateMetricsView,
        ImportDataStep.CreateExplore,
      ];
}

function transitionFromSchema(
  driver: V1ConnectorDriver,
  schema: string,
): AddDataState {
  if (isConnectorType(driver)) {
    return {
      step: AddDataStep.Connector,
      schema,
    };
  } else {
    return {
      step: AddDataStep.Source,
      schema,
      connector: driver.name!,
    };
  }
}

function transitionFromConnector(
  driver: V1ConnectorDriver,
  connector: string,
  schema: string,
): AddDataState {
  if (isExplorerType(driver)) {
    return {
      step: AddDataStep.Explorer,
      schema,
      connector,
    };
  } else {
    return {
      step: AddDataStep.Source,
      schema,
      connector,
    };
  }
}

function getConnectorDriverForSchema(
  schemaName: string,
): V1ConnectorDriver | null {
  const connectorInfo = connectors.find((c) => c.name === schemaName);
  if (!connectorInfo) return null;
  const schema = getConnectorSchema(connectorInfo.name);
  const category = schema?.["x-category"];
  const backendName = getBackendConnectorName(connectorInfo.name);

  return {
    name: backendName,
    displayName: connectorInfo.displayName,
    implementsObjectStore: category === "objectStore",
    implementsOlap: category === "olap",
    implementsSqlStore: category === "sqlStore",
    implementsWarehouse: category === "warehouse",
    implementsFileStore: category === "fileStore",
  };
}

function getConnectorDriverForConnector(
  runtimeClient: RuntimeClient,
  connectorName: string,
) {
  const queryKey = getRuntimeServiceAnalyzeConnectorsQueryKey(
    runtimeClient.instanceId,
  );
  const runtimeConnectorsResp =
    queryClient.getQueryData<V1AnalyzeConnectorsResponse>(queryKey);
  const analyzedConnector = runtimeConnectorsResp?.connectors?.find(
    (r) => r.name === connectorName,
  );
  return analyzedConnector;
}

function isConnectorType(connectorDriver: V1ConnectorDriver) {
  return (
    connectorDriver?.implementsFileStore ||
    connectorDriver?.implementsObjectStore ||
    connectorDriver?.implementsOlap ||
    connectorDriver?.implementsSqlStore ||
    (connectorDriver?.implementsWarehouse &&
      connectorDriver?.name !== "salesforce")
  );
}

function isExplorerType(connectorDriver: V1ConnectorDriver) {
  return connectorDriver?.implementsOlap || connectorDriver?.implementsSqlStore;
}

export function isLiveConnectorType(connectorDriver: V1ConnectorDriver) {
  return !!connectorDriver?.implementsOlap;
}
