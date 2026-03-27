import {
  type AddDataConfig,
  type AddDataState,
  AddDataStep,
  type AddDataTransitionArgs,
  ImportDataStep,
} from "@rilldata/web-common/features/add-data/steps/types.ts";
import { type V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  getConnectorDriverForConnector,
  getConnectorDriverForSchema,
  isConnectorType,
  isExplorerType,
  isLiveConnectorType,
} from "@rilldata/web-common/features/add-data/steps/utils.ts";

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
    if (!analyzedConnector?.driver) {
      throw new Error("Connector driver not found");
    }
    driver = analyzedConnector.driver;
    selectedSchema = driver.name;
  }
  if (!driver && selectedSchema) {
    driver = getConnectorDriverForSchema(selectedSchema);
  }

  switch (current.step) {
    case AddDataStep.SelectConnector:
      if (driver && selectedConnector && selectedSchema) {
        return transitionFromConnector(
          driver,
          selectedConnector,
          selectedSchema,
          args,
        );
      } else if (driver && selectedSchema) {
        return transitionFromSchema(driver, selectedSchema, args);
      } else {
        return current;
      }

    case AddDataStep.CreateConnector:
      if (!selectedConnector) {
        throw new Error("Connector must be specified");
      }
      if (!selectedSchema) {
        throw new Error(
          "Schema is missing for connector: " + selectedConnector,
        );
      }
      if (!driver) {
        throw new Error(
          "Connector driver not found for schema: " + selectedSchema,
        );
      }

      return transitionFromConnector(
        driver,
        selectedConnector,
        selectedSchema,
        args,
      );

    case AddDataStep.ExploreConnector:
    case AddDataStep.CreateModel: {
      if (!args.importConfig) {
        throw new Error("Import config must be specified");
      }

      return {
        step: AddDataStep.Import,
        importStep: ImportDataStep.Init,
        config: args.importConfig,
      };
    }

    // AddDataStep.Import is handled by `runImportStep`
    // AddDataStep.Done is handled by the parent component by ending the flow.
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

function transitionFromSchema(
  driver: V1ConnectorDriver,
  schema: string,
  args: AddDataTransitionArgs,
): AddDataState {
  if (isConnectorType(driver)) {
    return {
      step: AddDataStep.CreateConnector,
      schema,
    };
  } else {
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
    return {
      step: AddDataStep.ExploreConnector,
      schema,
      connector,
    };
  } else {
    return {
      step: AddDataStep.CreateModel,
      schema,
      connector,
      connectorFormValues: args.connectorFormValues ?? {},
    };
  }
}
