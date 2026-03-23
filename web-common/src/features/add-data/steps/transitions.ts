import {
  type AddDataConfig,
  type AddDataState,
  AddDataStep,
  type AddDataTransitionArgs,
  ImportDataStep,
} from "@rilldata/web-common/features/add-data/steps/types.ts";
import { type V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import {
  connectorInfoMap,
  getBackendConnectorName,
  getConnectorSchema,
  getDocsCategory,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { fetchAnalyzeConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";

export async function transitionToNextStep(
  runtimeClient: RuntimeClient,
  current: AddDataState,
  args: AddDataTransitionArgs,
): Promise<AddDataState> {
  const selectedConnector: string | undefined = args.connector;
  let selectedSchema: string | undefined = args.schema;

  let driver: V1ConnectorDriver | null = null;
  if (selectedConnector) {
    const analyzedConnector = await getConnectorDriverForConnector(
      runtimeClient,
      selectedConnector,
    );
    if (analyzedConnector) {
      driver = analyzedConnector.driver!;
      selectedSchema = driver.name!;
    }
  }
  if (!driver && selectedSchema) {
    driver = getConnectorDriverForSchema(selectedSchema);
  }

  console.log("[Transition] from", AddDataStep[current.step]);
  switch (current.step) {
    case AddDataStep.SelectConnector:
      if (selectedConnector) {
        return transitionFromConnector(
          driver!,
          selectedConnector,
          selectedSchema!,
          args,
        );
      } else if (selectedSchema) {
        return transitionFromSchema(driver!, selectedSchema, args);
      } else {
        return current;
      }

    case AddDataStep.CreateConnector:
      if (!selectedConnector) {
        throw new Error("Connector must be specified");
      }

      return transitionFromConnector(
        driver!,
        selectedConnector,
        selectedSchema!,
        args,
      );

    case AddDataStep.ExploreConnector:
    case AddDataStep.CreateModel: {
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

export async function maybeGetConnectorDriver(
  runtimeClient: RuntimeClient,
  schemaName: string | undefined,
  connectorName: string | undefined,
) {
  if (connectorName) {
    const analyzedConnector = await getConnectorDriverForConnector(
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
    ? [ImportDataStep.CreateMetricsView, ImportDataStep.CreateCanvas]
    : [
        ImportDataStep.CreateModel,
        ImportDataStep.CreateMetricsView,
        ImportDataStep.CreateCanvas,
      ];
  return config.importOnly ? [steps[0]] : steps;
}

export function getImportStepsForSource(config: AddDataConfig) {
  return config.importOnly
    ? [ImportDataStep.CreateModel]
    : [
        ImportDataStep.CreateModel,
        ImportDataStep.CreateMetricsView,
        ImportDataStep.CreateCanvas,
      ];
}

function transitionFromSchema(
  driver: V1ConnectorDriver,
  schema: string,
  args: AddDataTransitionArgs,
): AddDataState {
  if (isConnectorType(driver)) {
    console.log("[Transition] To CreateConnector");
    return {
      step: AddDataStep.CreateConnector,
      schema,
    };
  } else {
    console.log("[Transition] To CreateModel");
    return {
      step: AddDataStep.CreateModel,
      schema,
      connector: driver.name!,
      connectorFormValues: args.connectorFormValues ?? {},
    };
  }
}

function transitionFromConnector(
  driver: V1ConnectorDriver,
  connector: string,
  schema: string,
  args: AddDataTransitionArgs,
): AddDataState {
  if (isExplorerType(driver)) {
    console.log("[Transition] To ExploreConnector");
    return {
      step: AddDataStep.ExploreConnector,
      schema,
      connector,
    };
  } else {
    console.log("[Transition] To CreateModel");
    return {
      step: AddDataStep.CreateModel,
      schema,
      connector,
      connectorFormValues: args.connectorFormValues ?? {},
    };
  }
}

export function getConnectorDriverForSchema(
  schemaName: string,
): V1ConnectorDriver | null {
  const connectorInfo = connectorInfoMap.get(schemaName);
  if (!connectorInfo) return null;
  const schema = getConnectorSchema(connectorInfo.name);
  const category = schema?.["x-category"];
  const backendName = getBackendConnectorName(connectorInfo.name);

  return {
    name: backendName,
    displayName: schema?.title ?? connectorInfo.displayName ?? schemaName,
    docsUrl: `https://docs.rilldata.com/developers/build/connectors/${getDocsCategory(category)}/${backendName}`,
    implementsObjectStore: category === "objectStore",
    implementsOlap: category === "olap",
    implementsSqlStore: category === "sqlStore",
    implementsWarehouse: category === "warehouse",
    implementsFileStore: category === "fileStore",
    implementsAi: category === "ai",
  };
}

async function getConnectorDriverForConnector(
  runtimeClient: RuntimeClient,
  connectorName: string,
) {
  const connectors = await fetchAnalyzeConnectors(runtimeClient);
  return connectors.find((r) => r.name === connectorName);
}

function isConnectorType(connectorDriver: V1ConnectorDriver) {
  return (
    connectorDriver?.implementsObjectStore ||
    connectorDriver?.implementsOlap ||
    connectorDriver?.implementsSqlStore ||
    connectorDriver?.implementsWarehouse ||
    connectorDriver?.name === "https"
  );
}

function isExplorerType(connectorDriver: V1ConnectorDriver) {
  return (
    connectorDriver?.implementsOlap ||
    connectorDriver?.implementsSqlStore ||
    connectorDriver?.implementsWarehouse
  );
}

export function isLiveConnectorType(connectorDriver: V1ConnectorDriver) {
  return !!connectorDriver?.implementsOlap;
}
