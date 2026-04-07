import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import {
  connectorInfoMap,
  getBackendConnectorName,
  getConnectorSchema,
  getDocsCategory,
} from "@rilldata/web-common/features/sources/modal/connector-schemas.ts";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { fetchAnalyzeConnectors } from "@rilldata/web-common/features/connectors/selectors.ts";
import {
  type AddDataConfig,
  ImportDataStep,
} from "@rilldata/web-common/features/add-data/manager/steps/types.ts";

export function getConnectorDriverForSchema(
  schemaName: string,
): V1ConnectorDriver | undefined {
  const connectorInfo = connectorInfoMap.get(schemaName);
  if (!connectorInfo) return undefined;
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

export async function getConnectorDriverForConnector(
  runtimeClient: RuntimeClient,
  connectorName: string,
) {
  const connectors = await fetchAnalyzeConnectors(runtimeClient);
  return connectors.find((r) => r.name === connectorName);
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

export function isConnectorType(connectorDriver: V1ConnectorDriver) {
  return (
    connectorDriver?.implementsObjectStore ||
    connectorDriver?.implementsOlap ||
    connectorDriver?.implementsSqlStore ||
    connectorDriver?.implementsWarehouse ||
    connectorDriver?.name === "https"
  );
}

export function isExplorerType(connectorDriver: V1ConnectorDriver) {
  return (
    connectorDriver?.implementsOlap ||
    connectorDriver?.implementsSqlStore ||
    connectorDriver?.implementsWarehouse
  );
}

/**
 * A connector is "live" when it IS the active OLAP engine.
 * Live connectors create metrics views directly against their tables.
 * Non-live connectors ingest data into the OLAP engine first (e.g. into DuckDB).
 */
export function isLiveConnectorType(
  connectorDriver: V1ConnectorDriver,
  activeOlapConnector: string,
) {
  return connectorDriver?.name === activeOlapConnector;
}

const NonModelSteps = [
  ImportDataStep.CreateMetricsView,
  ImportDataStep.CreateDashboard,
];
const FullListOfSteps = [ImportDataStep.CreateModel, ...NonModelSteps];

export function getImportStepsForConnector(
  config: AddDataConfig,
  driver: V1ConnectorDriver,
  activeOlapConnector: string,
) {
  const steps = isLiveConnectorType(driver, activeOlapConnector)
    ? NonModelSteps
    : FullListOfSteps;
  return config.importOnly ? [steps[0]] : steps;
}

export function getImportStepsForSource(config: AddDataConfig) {
  return config.importOnly ? [ImportDataStep.CreateModel] : FullListOfSteps;
}
