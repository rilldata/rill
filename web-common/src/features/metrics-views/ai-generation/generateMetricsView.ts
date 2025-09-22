import { goto } from "$app/navigation";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { get } from "svelte/store";
import { overlay } from "../../../layout/overlay-store";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import { waitUntil } from "../../../lib/waitUtils";
import { behaviourEvent } from "../../../metrics/initMetrics";
import type { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  type MetricsEventSpace,
} from "../../../metrics/service/MetricsTypes";
import {
  type RuntimeServiceGenerateMetricsViewFileBody,
  type V1GenerateMetricsViewFileResponse,
  runtimeServiceGenerateMetricsViewFile,
  runtimeServiceGetFile,
} from "../../../runtime-client";
import httpClient from "../../../runtime-client/http-client";
import { getName } from "../../entity-management/name-utils";
import { featureFlags } from "../../feature-flags";
import { createAndPreviewExplore } from "../create-and-preview-explore";
import OptionToCancelAIGeneration from "./OptionToCancelAIGeneration.svelte";
import { createYamlModelFromTable } from "../../connectors/code-utils";

/**
 * TanStack Query does not support mutation cancellation (at least as of v4).
 * Here, we create our own version of `runtimeServiceGenerateMetricsViewFile` that accepts an
 * AbortSignal, which we can use to cancel the request.
 */
const runtimeServiceGenerateMetricsViewFileWithSignal = (
  instanceId: string,
  runtimeServiceGenerateMetricsViewFileBody: RuntimeServiceGenerateMetricsViewFileBody,
  signal: AbortSignal,
) => {
  return httpClient<V1GenerateMetricsViewFileResponse>({
    url: `/v1/instances/${instanceId}/files/generate-metrics-view`,
    method: "post",
    headers: { "Content-Type": "application/json" },
    data: runtimeServiceGenerateMetricsViewFileBody,
    signal,
  });
};

/**
 * Wrapper function that takes care of common UI side effects on top of creating a dashboard from a table.
 *
 * This function is to be called from all `Generate dashboard with AI` CTAs *outside* of the Metrics Editor.
 */
export function useCreateMetricsViewFromTableUIAction(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  tableName: string,
  createExplore: boolean,
  behaviourEventMedium: BehaviourEventMedium,
  metricsEventSpace: MetricsEventSpace,
) {
  const isAiEnabled = get(featureFlags.ai);

  // Return a function that can be called to create a dashboard from a table
  return async () => {
    let isAICancelled = false;
    const abortController = new AbortController();

    overlay.set({
      title: `Hang tight! ${isAiEnabled ? "AI is" : "We're"} personalizing your ${createExplore ? "dashboard" : "metrics"}`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("User canceled the AI generation");
            isAICancelled = true;
          },
        },
      },
    });

    // Get a unique name
    const newMetricsViewName = getName(
      `${tableName}_metrics`,
      fileArtifacts.getNamesForKind(ResourceKind.MetricsView),
    );
    const newMetricsViewFilePath = `/metrics/${newMetricsViewName}.yaml`;

    try {
      // First, request an AI-generated metrics view
      void runtimeServiceGenerateMetricsViewFileWithSignal(
        instanceId,
        {
          connector: connector,
          database: database,
          databaseSchema: databaseSchema,
          table: tableName,
          path: newMetricsViewFilePath,
          useAi: isAiEnabled, // AI isn't enabled during e2e tests
        },
        abortController.signal,
      );

      // Poll every second until the AI generation is complete or canceled
      while (!isAICancelled) {
        await new Promise((resolve) => setTimeout(resolve, 1000));

        try {
          await runtimeServiceGetFile(instanceId, {
            path: newMetricsViewFilePath,
          });

          // success, AI is done
          break;
        } catch {
          // 404 error, AI is not done
        }
      }

      // If the user canceled the AI request, submit another request with `useAi=false`
      if (isAICancelled) {
        await runtimeServiceGenerateMetricsViewFile(instanceId, {
          connector: connector,
          database: database,
          databaseSchema: databaseSchema,
          table: tableName,
          path: newMetricsViewFilePath,
          useAi: false,
        });
      }

      const previousScreenName = getScreenNameFromPage();

      // If we're not creating an Explore, navigate to the Metrics View file
      if (!createExplore) {
        await goto(`/files${newMetricsViewFilePath}`);
        void behaviourEvent?.fireNavigationEvent(
          newMetricsViewName,
          behaviourEventMedium,
          metricsEventSpace,
          previousScreenName,
          MetricsEventScreenName.MetricsDefinition,
        );
        overlay.set(null);
        return;
      }

      // If we are creating an Explore...

      // Get the Metrics View to use as a base for the Explore
      const metricsViewResource = fileArtifacts
        .getFileArtifact(newMetricsViewFilePath)
        .getResource(queryClient, instanceId);

      await waitUntil(() => get(metricsViewResource).data !== undefined, 5000);

      const resource = get(metricsViewResource).data;
      if (!resource) {
        throw new Error("Failed to create a Metrics View resource");
      }

      // Create the Explore file, and navigate to it
      await createAndPreviewExplore(queryClient, instanceId, resource);
    } catch (err) {
      eventBus.emit("notification", {
        message:
          `Failed to create ${createExplore ? "a dashboard" : "metrics"} for ` +
          tableName,
        detail: err.response?.data?.message ?? err.message,
      });
    }

    // Done, remove the overlay
    overlay.set(null);
  };
}

/**
 * Wrapper function that takes care of UI side effects on top of creating a dashboard from a model.
 *
 * This function is to be called from the `Generate dashboard with AI` CTA *inside* of the Metrics Editor.
 */
export async function createDashboardFromTableInMetricsEditor(
  instanceId: string,
  modelName: string,
  filePath: string,
) {
  const isAiEnabled = get(featureFlags.ai);

  const tableName = modelName;
  let isAICancelled = false;
  const abortController = new AbortController();

  overlay.set({
    title: `Hang tight! ${isAiEnabled ? "AI is" : "We're"} personalizing your dashboard`,
    detail: {
      component: OptionToCancelAIGeneration,
      props: {
        onCancel: () => {
          abortController.abort();
          isAICancelled = true;
        },
      },
    },
  });

  try {
    // First, request an AI-generated dashboard
    void runtimeServiceGenerateMetricsViewFileWithSignal(
      instanceId,
      {
        table: tableName,
        path: filePath,
        useAi: isAiEnabled, // AI isn't enabled during e2e tests
      },
      abortController.signal,
    );

    // Poll every second until the AI generation is complete or canceled
    while (!isAICancelled) {
      await new Promise((resolve) => setTimeout(resolve, 1000));

      try {
        const file = await runtimeServiceGetFile(instanceId, {
          path: filePath,
        });
        if (file.blob !== "") {
          // success, AI is done
          break;
        }
      } catch {
        // 404 error, AI is not done
      }
    }

    // If the user canceled the AI request, submit another request with `useAi=false`
    if (isAICancelled) {
      await runtimeServiceGenerateMetricsViewFile(instanceId, {
        table: tableName,
        path: filePath,
        useAi: false,
      });
    }
  } catch (err) {
    eventBus.emit("notification", {
      message: "Failed to create a dashboard for " + tableName,
      detail: err.response?.data?.message ?? err.message,
    });
  }

  // Done, remove the overlay
  overlay.set(null);
}

/**
 * Creates a model from a table, then generates a metrics view and explore dashboard.
 * This is used for non-OLAP connectors that need to follow the Rill architecture:
 * 1. Create model (ingests from source → OLAP)
 * 2. Create metrics view (on top of model)
 * 3. Create explore dashboard (on top of metrics view)
 */
export async function createModelAndMetricsViewAndExplore(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  let isAICancelled = false;
  const abortController = new AbortController();

  overlay.set({
    title: `Creating your metrics and dashboard...`,
    detail: {
      component: OptionToCancelAIGeneration,
      props: {
        onCancel: () => {
          abortController.abort("User canceled the AI generation");
          isAICancelled = true;
        },
      },
    },
  });

  try {
    // Step 1: Create model that ingests from source to OLAP
    overlay.set({
      title: `Creating model (ingesting from source → OLAP)...`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("User canceled the AI generation");
            isAICancelled = true;
          },
        },
      },
    });

    const [, modelName] = await createYamlModelFromTable(
      queryClient,
      connector,
      database,
      databaseSchema,
      table,
    );

    // Step 2: Wait for model to be ready
    const modelResource = fileArtifacts
      .getFileArtifact(`/models/${modelName}.yaml`)
      .getResource(queryClient, instanceId);

    await waitUntil(() => get(modelResource).data !== undefined, 10000);

    // Step 3: Create metrics view using the backend AI generation
    // This will properly analyze the model's schema and generate dimensions/measures
    const metricsViewName = `${table}_metrics`;
    const metricsViewFilePath = `/metrics/${metricsViewName}.yaml`;

    // Check if user cancelled
    if (isAICancelled) {
      throw new Error("User cancelled the operation");
    }

    // Update overlay for metrics view creation
    overlay.set({
      title: `Creating metrics view (on top of model)...`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("User canceled the AI generation");
            isAICancelled = true;
          },
        },
      },
    });

    // Use the backend function with the model name instead of table name
    await runtimeServiceGenerateMetricsViewFile(instanceId, {
      model: modelName, // Use model name instead of table
      path: metricsViewFilePath,
      useAi: get(featureFlags.ai),
    });

    // Step 4: Wait for metrics view to be ready
    const metricsViewResource = fileArtifacts
      .getFileArtifact(metricsViewFilePath)
      .getResource(queryClient, instanceId);

    await waitUntil(() => get(metricsViewResource).data !== undefined, 10000);

    const resource = get(metricsViewResource).data;
    if (!resource) {
      throw new Error("Failed to create a Metrics View resource");
    }

    // Check if user cancelled
    if (isAICancelled) {
      throw new Error("User cancelled the operation");
    }

    // Update overlay for explore dashboard creation
    overlay.set({
      title: `Creating explore dashboard (on top of metrics view)...`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("User canceled the AI generation");
            isAICancelled = true;
          },
        },
      },
    });

    // Step 5: Create explore dashboard
    await createAndPreviewExplore(queryClient, instanceId, resource);
  } catch (err) {
    console.error("Failed to create model and metrics view:", err);
    throw err;
  } finally {
    // Always clean up the overlay
    overlay.set(null);
  }
}
