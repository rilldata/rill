import {
  type ImportAddDataStep,
  ImportDataStep,
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
import { getName } from "@rilldata/web-common/features/entity-management/name-utils.ts";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { get } from "svelte/store";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { compileSourceYAML } from "@rilldata/web-common/features/sources/sourceUtils.ts";
import { maybeGetConnectorDriver } from "@rilldata/web-common/features/add-data/steps/transitions.ts";

export async function runImportStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
  onNewRoute: (newRoute: string) => void,
): Promise<ImportAddDataStep> {
  let newImportStep: ImportAddDataStep["importStep"];

  switch (step.importStep.step) {
    case ImportDataStep.Init:
      if (step.config.importSteps.length === 0) {
        throw new Error("Must specify at least one import step");
      }
      switch (step.config.importSteps[0]) {
        case ImportDataStep.CreateModel:
          newImportStep = {
            step: ImportDataStep.CreateModel,
            source: step.config.source,
            connector: step.config.connector,
            sql: step.config.sql,
            envBlob: step.config.envBlob,
          };
          break;

        case ImportDataStep.CreateMetricsView:
          newImportStep = {
            step: ImportDataStep.CreateMetricsView,
            source: step.config.source,
            sourceSchema: step.config.sourceSchema,
            sourceDatabase: step.config.sourceDatabase,
            connector: step.config.connector,
          };
          break;

        default:
          throw new Error(
            "First step must be one of: CreateModel, CreateMetricsView",
          );
      }
      break;

    case ImportDataStep.CreateModel:
      newImportStep = await runCreateModelStep(runtimeClient, step, onNewRoute);
      break;

    case ImportDataStep.CreateMetricsView:
      newImportStep = await runCreateMetricsViewStep(
        runtimeClient,
        step,
        onNewRoute,
      );
      break;

    case ImportDataStep.CreateCanvas:
      newImportStep = await runCreateCanvasStep(
        runtimeClient,
        step,
        onNewRoute,
      );
      break;

    case ImportDataStep.Done:
      return step;
  }

  return {
    ...step,
    importStep: newImportStep,
  };
}

async function runCreateModelStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
  onNewRoute: (newRoute: string) => void,
): Promise<ImportAddDataStep["importStep"]> {
  const modelImportStep = step.importStep;
  if (modelImportStep.step !== ImportDataStep.CreateModel) {
    throw new Error("Invalid model import step");
  }

  const modelName = getName(
    step.config.source,
    fileArtifacts.getNamesForKind(ResourceKind.Model),
  );
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
  const yaml = compileSourceYAML(
    connectorDriver,
    {
      name: modelName,
      sql: step.config.sql,
      database: step.config.sourceDatabase,
    },
    {
      connectorInstanceName: step.config.connector,
    },
  );

  const filePath = `/models/${step.config.source}.yaml`;
  onNewRoute(`/files${filePath}`);

  await runtimeServicePutFile(runtimeClient, {
    path: filePath,
    blob: yaml,
    create: true,
    createOnly: false,
  });

  if (modelImportStep.envBlob !== null) {
    // Make sure the file has reconciled before testing the connection
    await runtimeServicePutFileAndWaitForReconciliation(runtimeClient, {
      path: ".env",
      blob: modelImportStep.envBlob,
      create: true,
      createOnly: false,
    });
  }

  // Wait for the model to successfully reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    modelImportStep.source,
    ResourceKind.Model,
  );

  if (!step.config.importSteps.includes(ImportDataStep.CreateMetricsView)) {
    return {
      step: ImportDataStep.Done,
    };
  }

  return {
    step: ImportDataStep.CreateMetricsView,
    source: modelImportStep.source,
    sourceSchema: "",
    sourceDatabase: "",
    connector: modelImportStep.connector,
  };
}

async function runCreateMetricsViewStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
  onNewRoute: (newRoute: string) => void,
): Promise<ImportAddDataStep["importStep"]> {
  const metricsViewImportStep = step.importStep;
  if (metricsViewImportStep.step !== ImportDataStep.CreateMetricsView) {
    throw new Error("Invalid metrics view import step");
  }

  // Metrics view generation
  const newMetricsViewName = getName(
    `${metricsViewImportStep.source}_metrics`,
    fileArtifacts.getNamesForKind(ResourceKind.MetricsView),
  );
  const newMetricsViewFilePath = `/metrics/${newMetricsViewName}.yaml`;
  onNewRoute(`/files${newMetricsViewFilePath}`);

  // Call GenerateMetricsViewFile with the generated file path
  await runtimeServiceGenerateMetricsViewFile(runtimeClient, {
    table: metricsViewImportStep.source,
    connector: metricsViewImportStep.connector,
    database: metricsViewImportStep.sourceDatabase,
    databaseSchema: metricsViewImportStep.sourceSchema,
    path: newMetricsViewFilePath,
    useAi: true, // TODO: check feature flags
  });
  // Wait for the metrics view to successfully reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    newMetricsViewName,
    ResourceKind.MetricsView,
  );

  if (!step.config.importSteps.includes(ImportDataStep.CreateCanvas)) {
    return {
      step: ImportDataStep.Done,
    };
  }

  return {
    step: ImportDataStep.CreateCanvas,
    metricsViewFilePath: newMetricsViewFilePath,
  };
}

async function runCreateCanvasStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
  onNewRoute: (newRoute: string) => void,
): Promise<ImportAddDataStep["importStep"]> {
  const canvasImportStep = step.importStep;
  if (canvasImportStep.step !== ImportDataStep.CreateCanvas) {
    throw new Error("Invalid canvas import step");
  }

  // Get the MetricsView resource used to create the explore from.
  const metricsViewResourceResp = fileArtifacts
    .getFileArtifact(canvasImportStep.metricsViewFilePath)
    .getResource(queryClient);
  await waitUntil(() => get(metricsViewResourceResp).data !== undefined, 5000);
  const metricsViewResource = get(metricsViewResourceResp).data;
  if (!metricsViewResource) {
    throw new Error("Failed to create a Metrics View resource");
  }
  const metricsViewName = metricsViewResource.meta?.name?.name;
  if (!metricsViewName) {
    throw new Error("Failed to get MetricsView name");
  }

  // Get a unique name for the canvas dashboard
  const canvasName = getName(
    `${metricsViewName}}_canvas`,
    fileArtifacts.getNamesForKind(ResourceKind.Canvas),
  );
  const canvasFilePath = `/dashboards/${canvasName}.yaml`;

  await runtimeServiceGenerateCanvasFile(runtimeClient, {
    metricsViewName,
    path: canvasFilePath,
    useAi: true, // TODO: check feature flags
  });
  onNewRoute(`/files${canvasFilePath}`);

  // Wait for canvas to reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    canvasName,
    ResourceKind.Canvas,
  );

  return {
    step: ImportDataStep.Done,
  };
}
