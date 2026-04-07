import {
  type AddDataConfig,
  type ImportAddDataStep,
  ImportDataStep,
  type ImportFromConfig,
  type ImportStepConfig,
  type ImportToConfig,
} from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
import {
  runtimeServiceCreateTrigger,
  runtimeServiceGenerateCanvasFile,
  runtimeServiceGenerateMetricsViewFile,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import {
  deleteFileArtifact,
  maybeDeleteFileArtifact,
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
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { generateBlobForNewResourceFile } from "@rilldata/web-common/features/entity-management/add/new-files.ts";
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
import type { QueryClient } from "@tanstack/svelte-query";
import { unsetResourceEnvVars } from "@rilldata/web-common/features/connectors/code-utils.ts";
import { maybeGetConnectorDriver } from "@rilldata/web-common/features/add-data/manager/steps/utils.ts";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";

export async function runImportSteps(
  runtimeClient: RuntimeClient,
  addDataConfig: AddDataConfig,
  addDataStep: ImportAddDataStep,
  onProgress: (
    step: ImportDataStep,
    currentFilePath: string | undefined,
  ) => void,
) {
  for (const step of addDataStep.config.importSteps) {
    fireImportStepEvent(addDataConfig, addDataStep, step);
    switch (step) {
      case ImportDataStep.CreateModel:
        onProgress(step, addDataStep.config.importTo.modelPath);
        await runCreateModelStep(runtimeClient, addDataStep.config);
        break;
      case ImportDataStep.CreateMetricsView:
        onProgress(step, addDataStep.config.importTo.metricsViewPath);
        await runCreateMetricsViewStep(runtimeClient, addDataStep.config);
        break;
      case ImportDataStep.CreateDashboard:
        onProgress(step, addDataStep.config.importTo.explorePath);
        await runCreateExploreStep(runtimeClient, addDataStep.config);
        onProgress(step, addDataStep.config.importTo.canvasPath);
        await runCreateCanvasStep(runtimeClient, addDataStep.config);
        break;
    }
  }
  fireImportStepEvent(addDataConfig, addDataStep, ImportDataStep.Done);
  onProgress(ImportDataStep.Done, undefined);
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

export async function cleanupImportStep(
  runtimeClient: RuntimeClient,
  queryClient: QueryClient,
  config: ImportStepConfig,
) {
  const importToConfig = config.importTo;

  let envBlob: string | null = null;
  if (
    importToConfig.modelPath &&
    fileArtifacts.hasFileArtifact(importToConfig.modelPath)
  ) {
    const modelArtifact = fileArtifacts.getFileArtifact(
      importToConfig.modelPath,
    );
    const modelYaml = await modelArtifact.fetchContent();

    // Get the existing env and remove the connector's env vars
    envBlob = await unsetResourceEnvVars(
      runtimeClient,
      queryClient,
      modelYaml ?? "",
    );

    await deleteFileArtifact(runtimeClient, importToConfig.modelPath);
  }

  // Cleanup any generated files.
  await Promise.all(
    [
      importToConfig.metricsViewPath,
      importToConfig.explorePath,
      importToConfig.canvasPath,
    ].map((path) => {
      if (!path) return Promise.resolve();
      return maybeDeleteFileArtifact(runtimeClient, path);
    }),
  );

  if (envBlob) {
    // Update the .env file with the removed env vars
    await runtimeServicePutFile(runtimeClient, {
      path: ".env",
      blob: envBlob,
    });
  }
}

async function runCreateModelStep(
  runtimeClient: RuntimeClient,
  config: ImportStepConfig,
) {
  // Get the connector driver for the connector instance
  const connectorDriver = await maybeGetConnectorDriver(
    runtimeClient,
    undefined,
    config.connector,
  );
  if (!connectorDriver) {
    throw new Error(`Failed to get connector driver for ${config.connector}`);
  }

  // Validate model name and path are generated upstream
  const importToConfig = config.importTo;
  if (!importToConfig.modelName || !importToConfig.modelPath) {
    throw new Error("Model name and path must be generated upstream.");
  }

  // Build the model YAML based on the import source
  let yaml = "";
  const importFromConfig = config.importFrom;
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
          connectorInstanceName: config.connector,
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
          connectorInstanceName: config.connector,
        },
      );
      break;
    }
  }

  if (config.envBlob) {
    // Make sure the file has reconciled before testing the connection
    await runtimeServicePutFileAndWaitForReconciliation(runtimeClient, {
      path: ".env",
      blob: config.envBlob,
      create: true,
      createOnly: false,
    });
  }

  let putFile = true;
  // Determine if the model file already exists and has the same content as the generated YAML.
  // We trigger a refresh if the file exists and has same content, backend doesnt do this automatically for optimization.
  if (fileArtifacts.hasFileArtifact(importToConfig.modelPath)) {
    const fileArtifact = fileArtifacts.getFileArtifact(
      importToConfig.modelPath,
    );
    const existingYaml = await fileArtifact.fetchContent();
    putFile = existingYaml !== yaml;
  }
  if (putFile) {
    // Create the model file with the generated YAML
    await runtimeServicePutFile(runtimeClient, {
      path: importToConfig.modelPath,
      blob: yaml,
      create: true,
      createOnly: false,
    });
  } else {
    // Trigger model refresh to reconcile the model file
    await runtimeServiceCreateTrigger(runtimeClient, {
      models: [{ model: importToConfig.modelName, full: true }],
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
  config: ImportStepConfig,
) {
  // Validate metrics view name and path are generated upstream
  const importToConfig = config.importTo;
  if (!importToConfig.metricsViewName || !importToConfig.metricsViewPath) {
    throw new Error("Metrics view name and path must be generated upstream.");
  }

  let connector = config.connector;
  let table = "";
  let database = "";
  let databaseSchema = "";
  if (
    config.importSteps.includes(ImportDataStep.CreateModel) &&
    importToConfig.modelPath
  ) {
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
  } else if (config.importFrom.from === "table") {
    table = config.importFrom.table;
    database = config.importFrom.database;
    databaseSchema = config.importFrom.schema;
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
  config: ImportStepConfig,
) {
  // Validate explore name and path are generated upstream
  const importToConfig = config.importTo;
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
  config: ImportStepConfig,
) {
  // Validate canvas name and path are generated upstream
  const importToConfig = config.importTo;
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

function fireImportStepEvent(
  addDataConfig: AddDataConfig,
  addDataStep: ImportAddDataStep,
  step: ImportDataStep,
) {
  void behaviourEvent?.fireAddDataStepEvent(
    BehaviourEventAction.ImportStep,
    addDataConfig.medium,
    addDataConfig.space,
    addDataConfig.screen,
    {
      step,
      schema: addDataStep.schema,
      connector: addDataStep.config.connector,
    },
  );
}
