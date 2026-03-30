import {
  type AddDataConfig,
  type AddDataState,
  AddDataStep,
  type AddDataTransitionArgs,
  ImportDataStep,
  type ImportStepConfig,
} from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
import { type V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  getConnectorDriverForConnector,
  getConnectorDriverForSchema,
  isConnectorType,
  isExplorerType,
  isLiveConnectorType,
} from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { goto } from "$app/navigation";

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

const NonModelSteps = [
  ImportDataStep.CreateMetricsView,
  ImportDataStep.CreateDashboard,
];
const FullListOfSteps = [ImportDataStep.CreateModel, ...NonModelSteps];

export function getImportStepsForConnector(
  config: AddDataConfig,
  driver: V1ConnectorDriver,
) {
  // Live connectors cannot create models as of now.
  // They will create metrics views directly.
  const steps = isLiveConnectorType(driver) ? NonModelSteps : FullListOfSteps;
  return config.importOnly ? [steps[0]] : steps;
}

export function getImportStepsForSource(config: AddDataConfig) {
  return config.importOnly ? [ImportDataStep.CreateModel] : FullListOfSteps;
}
