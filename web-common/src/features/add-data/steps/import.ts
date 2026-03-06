import {
  type ImportAddDataStep,
  ImportDataStep,
} from "@rilldata/web-common/features/add-data/steps/types.ts";
import {
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
import { createResourceFile } from "@rilldata/web-common/features/file-explorer/new-files.ts";
import { splitFolderFileNameAndExtension } from "@rilldata/web-common/features/entity-management/file-path-utils.ts";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

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
            yaml: step.config.yaml,
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

    case ImportDataStep.CreateExplore:
      newImportStep = await runCreateExploreStep(
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

  const filePath = `/models/${step.config.source}.yaml`;
  onNewRoute(`/files${filePath}`);

  await runtimeServicePutFile(runtimeClient, {
    path: filePath,
    blob: modelImportStep.yaml,
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
    useAi: false, // TODO: check feature flags
  });
  // Wait for the metrics view to successfully reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    newMetricsViewName,
    ResourceKind.MetricsView,
  );

  return {
    step: ImportDataStep.CreateExplore,
    metricsViewFilePath: newMetricsViewFilePath,
  };
}

async function runCreateExploreStep(
  runtimeClient: RuntimeClient,
  step: ImportAddDataStep,
  onNewRoute: (newRoute: string) => void,
): Promise<ImportAddDataStep["importStep"]> {
  const exploreImportStep = step.importStep;
  if (exploreImportStep.step !== ImportDataStep.CreateExplore) {
    throw new Error("Invalid explore import step");
  }

  // Get the MetricsView resource used to create the explore from.
  const metricsViewResourceResp = fileArtifacts
    .getFileArtifact(exploreImportStep.metricsViewFilePath)
    .getResource(queryClient);
  await waitUntil(() => get(metricsViewResourceResp).data !== undefined, 5000);
  const metricsViewResource = get(metricsViewResourceResp).data;
  if (!metricsViewResource) {
    throw new Error("Failed to create a Metrics View resource");
  }

  // Create the Explore file
  const exploreFilePath = await createResourceFile(
    runtimeClient,
    ResourceKind.Explore,
    metricsViewResource,
  );

  const [, exploreName] = splitFolderFileNameAndExtension(exploreFilePath);
  onNewRoute(`/explore/${exploreName}`);

  // Wait for explore to reconcile
  await waitForResourceReconciliation(
    runtimeClient,
    exploreName,
    ResourceKind.Explore,
  );

  return {
    step: ImportDataStep.Done,
  };
}
