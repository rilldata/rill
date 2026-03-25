import {
  type ImportAddDataStep,
  ImportDataStep,
  type ImportFromConfig,
  type ImportToConfig,
} from "@rilldata/web-common/features/add-data/steps/types.ts";
import {
  runtimeServiceGenerateCanvasFile,
  runtimeServiceGenerateMetricsViewFile,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import {
  runtimeServicePutFileAndWaitForReconciliation,
  waitForResourceReconciliation,
} from "@rilldata/web-common/features/entity-management/actions.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { get } from "svelte/store";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  compileSourceYAML,
  inferModelNameFromSQL,
} from "@rilldata/web-common/features/sources/sourceUtils.ts";
import { maybeGetConnectorDriver } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { generateBlobForNewResourceFile } from "@rilldata/web-common/features/entity-management/add/new-files.ts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";

export async function runImportStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
): Promise<ImportAddDataStep> {
  const advanceToNextStep = (
    importStep?: ImportDataStep,
  ): ImportAddDataStep => {
    if (!importStep) {
      const curStepIndex = step.config.importSteps.findIndex(
        (is) => is === step.importStep,
      );
      if (curStepIndex === -1) {
        throw new Error("Invalid import step");
      } else if (curStepIndex === step.config.importSteps.length - 1) {
        importStep = ImportDataStep.Done;
      } else {
        importStep = step.config.importSteps[curStepIndex + 1];
      }
    }

    switch (importStep) {
      case ImportDataStep.CreateModel:
        return {
          ...step,
          importStep: ImportDataStep.CreateModel,
          currentFilePath: step.config.importTo.modelPath,
        };

      case ImportDataStep.CreateMetricsView:
        return {
          ...step,
          importStep: ImportDataStep.CreateMetricsView,
          currentFilePath: step.config.importTo.metricsViewPath,
        };

      case ImportDataStep.CreateExplore:
        return {
          ...step,
          importStep: ImportDataStep.CreateExplore,
          currentFilePath: step.config.importTo.explorePath,
        };

      case ImportDataStep.CreateCanvas:
        return {
          ...step,
          importStep: ImportDataStep.CreateCanvas,
          currentFilePath: step.config.importTo.canvasPath,
        };

      case ImportDataStep.Done:
        return {
          ...step,
          importStep: ImportDataStep.Done,
        };
    }

    return step;
  };

  switch (step.importStep) {
    case ImportDataStep.Init:
      if (step.config.importSteps.length === 0) {
        throw new Error("Must specify at least one import step");
      }
      return advanceToNextStep(step.config.importSteps[0]);

    case ImportDataStep.CreateModel:
      await runCreateModelStep(runtimeClient, step);
      break;

    case ImportDataStep.CreateMetricsView:
      await runCreateMetricsViewStep(runtimeClient, step);
      break;

    case ImportDataStep.CreateExplore:
      await runCreateExploreStep(runtimeClient, step);
      break;

    case ImportDataStep.CreateCanvas:
      await runCreateCanvasStep(runtimeClient, step);
      break;
  }

  return advanceToNextStep();
}

export function generateImportToConfig(
  importFromConfig: ImportFromConfig,
  inputModelName?: string,
) {
  const importToConfig: ImportToConfig = {};

  let modelName: string | undefined = inputModelName;
  switch (importFromConfig.from) {
    case "sql":
      modelName = inferModelNameFromSQL(importFromConfig.sql);
      break;

    case "table":
      modelName = importFromConfig.table;
      break;
  }
  if (!modelName) return importToConfig;

  importToConfig.modelName = getName(
    modelName,
    fileArtifacts.getNamesForKind(ResourceKind.Model),
  );
  importToConfig.modelPath = `/models/${importToConfig.modelName}.yaml`;

  importToConfig.metricsViewName = getName(
    `${importToConfig.modelName}_metrics`,
    fileArtifacts.getNamesForKind(ResourceKind.MetricsView),
  );
  importToConfig.metricsViewPath = `/metrics/${importToConfig.metricsViewName}.yaml`;

  importToConfig.exploreName = getName(
    `${importToConfig.metricsViewName}_explore`,
    fileArtifacts.getNamesForKind(ResourceKind.Explore),
  );
  importToConfig.explorePath = `/dashboards/${importToConfig.exploreName}.yaml`;

  importToConfig.canvasName = getName(
    `${importToConfig.metricsViewName}_canvas`,
    fileArtifacts.getNamesForKind(ResourceKind.Canvas),
  );
  importToConfig.canvasPath = `/dashboards/${importToConfig.canvasName}.yaml`;

  return importToConfig;
}

async function runCreateModelStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
) {
  // Get the connector driver for the connector instance
  const connectorDriver = await maybeGetConnectorDriver(
    runtimeClient,
    undefined,
    step.config.connector,
  );
  if (!connectorDriver) {
    throw new Error(
      `Failed to get connector driver for ${step.config.connector}`,
    );
  }

  // Validate model name and path are generated upstream
  const importToConfig = step.config.importTo;
  if (!importToConfig.modelName || !importToConfig.modelPath) {
    throw new Error("Model name and path must be generated upstream.");
  }

  // Build the model YAML based on the import source
  let yaml = "";
  const importFromConfig = step.config.importFrom;
  switch (importFromConfig.from) {
    // Generated using a form directly into yaml.
    case "yaml":
      yaml = importFromConfig.yaml;
      break;

    // User provided a SQL query to generate the model.
    case "sql":
      yaml = compileSourceYAML(
        connectorDriver,
        {
          name: importToConfig.modelName,
          sql: importFromConfig.sql,
        },
        {
          connectorInstanceName: step.config.connector,
        },
      );
      break;

    // User selected a table to generate the model.
    case "table": {
      const fromTableName =
        (importFromConfig.database ? importFromConfig.database + "." : "") +
        (importFromConfig.schema ? importFromConfig.schema + "." : "") +
        importFromConfig.table;
      const sql = `SELECT * FROM ${fromTableName}`;
      yaml = compileSourceYAML(
        connectorDriver,
        {
          name: importToConfig.modelName,
          sql: sql,
        },
        {
          connectorInstanceName: step.config.connector,
        },
      );
      break;
    }
  }

  // Create the model file with the generated YAML
  await runtimeServicePutFile(runtimeClient, {
    path: importToConfig.modelPath,
    blob: yaml,
    create: true,
    createOnly: false,
  });

  if (step.config.envBlob) {
    // Make sure the file has reconciled before testing the connection
    await runtimeServicePutFileAndWaitForReconciliation(runtimeClient, {
      path: ".env",
      blob: step.config.envBlob,
      create: true,
      createOnly: false,
    });
  }

  // Wait for the model to successfully reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    importToConfig.modelName,
    ResourceKind.Model,
  );
}

async function runCreateMetricsViewStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
) {
  // Validate metrics view name and path are generated upstream
  const importToConfig = step.config.importTo;
  if (!importToConfig.metricsViewName || !importToConfig.metricsViewPath) {
    throw new Error("Metrics view name and path must be generated upstream.");
  }

  let connector = step.config.connector;
  let table = "";
  let database = "";
  let databaseSchema = "";
  if (importToConfig.modelPath) {
    // Get the model and use it's sink table/connector
    const modelFileArtifact = fileArtifacts.getFileArtifact(
      importToConfig.modelPath,
    );
    const modelResource = await modelFileArtifact.fetchResource(queryClient);
    if (!modelResource?.model) {
      throw new Error("Failed to get model resource");
    }
    // Use the model's result table and connector as the metrics view's source table/connector.
    // Database and schema do not apply in this case.
    table = modelResource.model.state?.resultTable ?? "";
    connector = modelResource.model.spec?.outputConnector ?? "";
  } else if (step.config.importFrom.from === "table") {
    table = step.config.importFrom.table;
    database = step.config.importFrom.database;
    databaseSchema = step.config.importFrom.schema;
  } else {
    throw new Error(
      "Must specify a model name or table to create a metrics view",
    );
  }

  // Call GenerateMetricsViewFile with the generated file path
  await runtimeServiceGenerateMetricsViewFile(runtimeClient, {
    connector,
    table,
    database,
    databaseSchema,
    path: importToConfig.metricsViewPath,
    useAi: get(featureFlags.ai),
  });
  // Wait for the metrics view to successfully reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    importToConfig.metricsViewName,
    ResourceKind.MetricsView,
  );
}

async function runCreateExploreStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
) {
  // Validate explore name and path are generated upstream
  const importToConfig = step.config.importTo;
  if (!importToConfig.exploreName || !importToConfig.explorePath) {
    throw new Error("Explore name and path must be generated upstream.");
  }

  // Get the metrics view resource for this explore
  if (!importToConfig.metricsViewPath) {
    throw new Error("Metrics view must be specified for this step.");
  }
  const metricsViewFileArtifact = fileArtifacts.getFileArtifact(
    importToConfig.metricsViewPath,
  );
  const metricsViewResource =
    await metricsViewFileArtifact.fetchResource(queryClient);
  if (!metricsViewResource?.metricsView?.state?.validSpec) {
    throw new Error("Failed to get metrics view resource");
  }

  // Generate a blank explore file.
  await runtimeServicePutFile(runtimeClient, {
    path: importToConfig.explorePath,
    blob: generateBlobForNewResourceFile(
      ResourceKind.Explore,
      metricsViewResource,
    ),
    create: true,
    createOnly: true,
  });

  // Wait for explore to reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    importToConfig.exploreName,
    ResourceKind.Explore,
  );
}

async function runCreateCanvasStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
) {
  // Validate canvas name and path are generated upstream
  const importToConfig = step.config.importTo;
  if (!importToConfig.canvasName || !importToConfig.canvasPath) {
    throw new Error("Canvas name and path must be generated upstream.");
  }

  await runtimeServiceGenerateCanvasFile(runtimeClient, {
    metricsViewName: importToConfig.metricsViewName,
    path: importToConfig.canvasPath,
    useAi: get(featureFlags.ai),
  });

  // Wait for canvas to reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    importToConfig.canvasName,
    ResourceKind.Canvas,
  );
}
