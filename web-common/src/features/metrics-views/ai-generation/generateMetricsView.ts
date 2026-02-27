import { goto } from "$app/navigation";
import { createCanvasDashboardWithoutNavigation } from "@rilldata/web-common/features/canvas/ai-generation/generateCanvas";
import { pollForFileCreation } from "@rilldata/web-common/features/entity-management/actions";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import {
  ResourceKind,
  resourceIsLoading,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { createResourceFile } from "@rilldata/web-common/features/file-explorer/new-files";
import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { overlay } from "../../../layout/overlay-store";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import { waitUntil } from "../../../lib/waitUtils";
import { behaviourEvent } from "../../../metrics/initMetrics";
import { BehaviourEventMedium } from "../../../metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "../../../metrics/service/MetricsTypes";
import {
  runtimeServiceGenerateMetricsViewFile,
  runtimeServiceGetFile,
  type RuntimeServiceGenerateMetricsViewFileBody,
  type V1GenerateMetricsViewFileResponse,
  type V1Resource,
} from "../../../runtime-client";
import httpClient from "../../../runtime-client/http-client";
import { createYamlModelFromTable } from "../../connectors/code-utils";
import { getName } from "../../entity-management/name-utils";
import { featureFlags } from "../../feature-flags";
import { createAndPreviewExplore } from "../create-and-preview-explore";
import OptionToCancelAIGeneration from "./OptionToCancelAIGeneration.svelte";

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
    const abortController = new AbortController();

    overlay.set({
      title: `Hang tight! ${isAiEnabled ? "AI is" : "We're"} personalizing your ${createExplore ? "dashboard" : "metrics"}`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("AI generation cancelled by user");
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

      // Poll until file creation is complete or canceled
      const fileCreated = await pollForFileCreation(
        instanceId,
        newMetricsViewFilePath,
        abortController.signal,
      );

      // If the user canceled the AI request, submit another request with `useAi=false`
      if (!fileCreated) {
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
          abortController.abort("Dashboard generation cancelled by user");
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
 * Unified function that generates metrics (and optionally explore dashboard) from a table.
 * Handles both OLAP and non-OLAP connectors with appropriate logic for each case.
 */
export async function generateMetricsFromTable(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
  createExplore: boolean,
  isOlapConnector: boolean,
  behaviourEventMedium: BehaviourEventMedium = BehaviourEventMedium.Menu,
  metricsEventSpace: MetricsEventSpace = MetricsEventSpace.LeftPanel,
) {
  if (isOlapConnector) {
    // For OLAP connectors, use direct metrics view generation
    const createMetricsViewFromTable = useCreateMetricsViewFromTableUIAction(
      instanceId,
      connector,
      database,
      databaseSchema,
      table,
      createExplore,
      behaviourEventMedium,
      metricsEventSpace,
    );
    await createMetricsViewFromTable();
  } else {
    // For non-OLAP connectors, follow Rill architecture: Model → Metrics → (Optional) Explore
    await createModelAndMetricsAndExplore(
      instanceId,
      connector,
      database,
      databaseSchema,
      table,
      createExplore,
    );
  }
}

/**
 * Creates a model from a table, then generates a metrics view and optionally an explore dashboard.
 * This is used for non-OLAP connectors that need to follow the Rill architecture:
 * 1. Create model (ingests from source → OLAP)
 * 2. Create metrics view (on top of model)
 * 3. Optionally create explore dashboard (on top of metrics view)
 */
export async function createModelAndMetricsAndExplore(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
  createExplore: boolean = true,
) {
  let isAICancelled = false;
  const abortController = new AbortController();

  const isAiEnabled = get(featureFlags.ai);
  overlay.set({
    title: `Creating your ${createExplore ? "metrics and dashboard" : "metrics"}${isAiEnabled ? " with AI" : ""}...`,
    detail: {
      component: OptionToCancelAIGeneration,
      props: {
        onCancel: () => {
          abortController.abort("Metrics creation cancelled by user");
          isAICancelled = true;
        },
      },
    },
  });

  try {
    // Step 1: Create model that ingests from source to OLAP
    overlay.set({
      title: `Creating model...`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("Model creation cancelled by user");
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
      title: `Creating metrics view${isAiEnabled ? " with AI" : ""}...`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("Metrics view creation cancelled by user");
            isAICancelled = true;
          },
        },
      },
    });

    // Use the backend function with the model name instead of table name
    void runtimeServiceGenerateMetricsViewFile(
      instanceId,
      {
        model: modelName, // Use model name instead of table
        path: metricsViewFilePath,
        useAi: get(featureFlags.ai),
      },
      abortController.signal,
    );

    // Poll every second until the AI generation is complete or canceled
    while (!isAICancelled) {
      await new Promise((resolve) => setTimeout(resolve, 1000));

      try {
        await runtimeServiceGetFile(instanceId, {
          path: metricsViewFilePath,
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
        model: modelName,
        path: metricsViewFilePath,
        useAi: false,
      });
    }

    // Step 4: Wait for metrics view to be ready
    const metricsViewResource = fileArtifacts
      .getFileArtifact(metricsViewFilePath)
      .getResource(queryClient, instanceId);

    await waitUntil(() => get(metricsViewResource).data !== undefined, 10000);

    const resource = get(metricsViewResource).data;
    if (!resource) {
      throw new Error("Failed to create a Metrics View resource");
    }

    // If we're not creating an Explore, navigate to the Metrics View file
    if (!createExplore) {
      const previousScreenName = getScreenNameFromPage();
      await goto(`/files${metricsViewFilePath}`);
      void behaviourEvent?.fireNavigationEvent(
        metricsViewName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        previousScreenName,
        MetricsEventScreenName.MetricsDefinition,
      );
      return;
    }

    // If we are creating an Explore...

    // Update overlay for explore dashboard creation
    overlay.set({
      title: `Creating explore dashboard...`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort(
              "Explore dashboard creation cancelled by user",
            );
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

/**
 * Waits for a metrics view resource to finish reconciling.
 * Reconciliation is complete when the status is IDLE.
 */
async function waitForMetricsViewReconciliation(
  instanceId: string,
  metricsViewFilePath: string,
  timeoutMs: number = 10000,
): Promise<void> {
  const metricsViewResource = fileArtifacts
    .getFileArtifact(metricsViewFilePath)
    .getResource(queryClient, instanceId);

  // Wait for the resource to be fully reconciled
  await waitUntil(() => {
    const resource = get(metricsViewResource).data;
    return resource !== undefined && !resourceIsLoading(resource);
  }, timeoutMs);
}

/**
 * Helper function to create metrics view from table with AI.
 * Returns the metrics view resource after creation.
 */
async function createMetricsViewFromTable(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  tableName: string,
  abortController: AbortController,
): Promise<V1Resource> {
  const isAiEnabled = get(featureFlags.ai);

  const newMetricsViewName = getName(
    `${tableName}_metrics`,
    fileArtifacts.getNamesForKind(ResourceKind.MetricsView),
  );
  const newMetricsViewFilePath = `/metrics/${newMetricsViewName}.yaml`;

  // Request an AI-generated metrics view
  void runtimeServiceGenerateMetricsViewFileWithSignal(
    instanceId,
    {
      connector: connector,
      database: database,
      databaseSchema: databaseSchema,
      table: tableName,
      path: newMetricsViewFilePath,
      useAi: isAiEnabled,
    },
    abortController.signal,
  );

  // Poll until file creation is complete or canceled
  const fileCreated = await pollForFileCreation(
    instanceId,
    newMetricsViewFilePath,
    abortController.signal,
  );

  // If the user canceled the AI request, submit another request with `useAi=false`
  if (!fileCreated) {
    await runtimeServiceGenerateMetricsViewFile(instanceId, {
      connector: connector,
      database: database,
      databaseSchema: databaseSchema,
      table: tableName,
      path: newMetricsViewFilePath,
      useAi: false,
    });
  }

  // Wait for Metrics View resource to be ready
  const metricsViewResource = fileArtifacts
    .getFileArtifact(newMetricsViewFilePath)
    .getResource(queryClient, instanceId);

  await waitUntil(() => get(metricsViewResource).data !== undefined, 5000);

  const resource = get(metricsViewResource).data;
  if (!resource) {
    throw new Error("Failed to create a Metrics View resource");
  }

  // Wait for the metrics view to finish reconciling before returning
  await waitForMetricsViewReconciliation(instanceId, newMetricsViewFilePath);

  return resource;
}

/**
 * Creates an Explore dashboard file without navigation.
 * Returns the file path of the created explore.
 */
export async function createExploreWithoutNavigation(
  queryClient: QueryClient,
  instanceId: string,
  metricsViewResource: V1Resource,
): Promise<string> {
  // Create the Explore file
  const filePath = await createResourceFile(
    ResourceKind.Explore,
    metricsViewResource,
  );

  // Wait until the Explore resource is ready
  const fileArtifact = fileArtifacts.getFileArtifact(filePath);
  const resource = fileArtifact.getResource(queryClient, instanceId);

  await waitUntil(() => {
    return get(resource).data !== undefined;
  }, 10000);

  const name = get(resource).data?.meta?.name?.name;
  if (!name) throw new Error("Failed to create an Explore resource");

  return filePath;
}

/**
 * Wrapper function that creates metrics view and canvas dashboard from a table.
 * Navigates to canvas dashboard when complete.
 *
 * This function is to be called from "Generate Canvas dashboard" CTA when you want
 * to create Canvas dashboard only (without Explore).
 */
export function useCreateMetricsViewWithCanvasUIAction(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  tableName: string,
  behaviourEventMedium: BehaviourEventMedium,
  metricsEventSpace: MetricsEventSpace,
) {
  const isAiEnabled = get(featureFlags.ai);

  return async () => {
    const abortController = new AbortController();

    overlay.set({
      title: `Hang tight! ${isAiEnabled ? "AI is" : "We're"} building your Canvas dashboard`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("AI generation cancelled by user");
          },
        },
      },
    });

    try {
      // Step 1: Create metrics view
      const resource = await createMetricsViewFromTable(
        instanceId,
        connector,
        database,
        databaseSchema,
        tableName,
        abortController,
      );

      const metricsViewName = resource.meta?.name?.name;
      if (!metricsViewName) {
        throw new Error("Failed to get metrics view name");
      }

      // Step 2: Wait a bit for metrics view to fully reconcile
      // This ensures the metrics view is ready before we generate the canvas
      overlay.set({
        title: `Preparing metrics view...`,
        detail: {
          component: OptionToCancelAIGeneration,
          props: {
            onCancel: () => {
              abortController.abort("Canvas creation cancelled by user");
            },
          },
        },
      });

      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Step 3: Create Canvas dashboard
      overlay.set({
        title: `Creating Canvas dashboard${isAiEnabled ? " with AI" : ""}...`,
        detail: {
          component: OptionToCancelAIGeneration,
          props: {
            onCancel: () => {
              abortController.abort("Canvas creation cancelled by user");
            },
          },
        },
      });

      const canvasFilePath = await createCanvasDashboardWithoutNavigation(
        instanceId,
        metricsViewName,
      );

      // Step 4: Navigate to Canvas dashboard
      if (canvasFilePath) {
        await goto(`/files${canvasFilePath}`);
        void behaviourEvent?.fireNavigationEvent(
          metricsViewName,
          behaviourEventMedium,
          metricsEventSpace,
          MetricsEventScreenName.Source,
          MetricsEventScreenName.Canvas,
        );
      } else {
        throw new Error("Failed to create Canvas dashboard");
      }
    } catch (err) {
      eventBus.emit("notification", {
        message: "Failed to create Canvas dashboard for " + tableName,
        detail: err.response?.data?.message ?? err.message,
      });
    }

    // Done, remove the overlay
    overlay.set(null);
  };
}

/**
 * Wrapper function that creates metrics view, explore dashboard, and canvas dashboard from a table.
 * Navigates to canvas dashboard if successful, otherwise falls back to explore dashboard.
 *
 * This function is to be called from "Generate dashboard" CTA when canvas feature is enabled.
 */
export function useCreateMetricsViewWithCanvasAndExploreUIAction(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  tableName: string,
  behaviourEventMedium: BehaviourEventMedium,
  metricsEventSpace: MetricsEventSpace,
) {
  const isAiEnabled = get(featureFlags.ai);

  // Return a function that can be called to create dashboards from a table
  return async () => {
    const abortController = new AbortController();

    overlay.set({
      title: `Hang tight! ${isAiEnabled ? "AI is" : "We're"} building your dashboard`,
      detail: {
        component: OptionToCancelAIGeneration,
        props: {
          onCancel: () => {
            abortController.abort("AI generation cancelled by user");
          },
        },
      },
    });

    let exploreFilePath: string | null = null;
    let canvasFilePath: string | null = null;
    let metricsViewName: string | undefined;

    try {
      // Step 1: Create metrics view
      const resource = await createMetricsViewFromTable(
        instanceId,
        connector,
        database,
        databaseSchema,
        tableName,
        abortController,
      );

      metricsViewName = resource.meta?.name?.name;
      if (!metricsViewName) {
        throw new Error("Failed to get metrics view name");
      }

      // Step 2: Wait a bit for metrics view to fully reconcile
      // This ensures the metrics view is ready before we generate dashboards
      overlay.set({
        title: `Preparing metrics view...`,
        detail: {
          component: OptionToCancelAIGeneration,
          props: {
            onCancel: () => {
              abortController.abort("Dashboard creation cancelled by user");
            },
          },
        },
      });

      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Step 3: Create Explore dashboard (without navigation)
      overlay.set({
        title: `Creating Explore dashboard...`,
        detail: {
          component: OptionToCancelAIGeneration,
          props: {
            onCancel: () => {
              abortController.abort("Dashboard creation cancelled by user");
            },
          },
        },
      });

      exploreFilePath = await createExploreWithoutNavigation(
        queryClient,
        instanceId,
        resource,
      );

      // Step 4: Try to create Canvas dashboard
      overlay.set({
        title: `Creating Canvas dashboard${isAiEnabled ? " with AI" : ""}...`,
        detail: {
          component: OptionToCancelAIGeneration,
          props: {
            onCancel: () => {
              abortController.abort("Canvas creation cancelled by user");
            },
          },
        },
      });

      canvasFilePath = await createCanvasDashboardWithoutNavigation(
        instanceId,
        metricsViewName,
      );

      // Step 5: Navigate to Canvas if successful, otherwise Explore
      if (canvasFilePath) {
        await goto(`/files${canvasFilePath}`);
        void behaviourEvent?.fireNavigationEvent(
          metricsViewName,
          behaviourEventMedium,
          metricsEventSpace,
          MetricsEventScreenName.Source,
          MetricsEventScreenName.Canvas,
        );
      } else if (exploreFilePath) {
        await goto(`/files${exploreFilePath}`);
        void behaviourEvent?.fireNavigationEvent(
          metricsViewName,
          behaviourEventMedium,
          metricsEventSpace,
          MetricsEventScreenName.Source,
          MetricsEventScreenName.Explore,
        );
        eventBus.emit("notification", {
          message: "Canvas dashboard creation failed, opened Explore instead",
        });
      }
    } catch (err) {
      eventBus.emit("notification", {
        message: "Failed to create dashboards for " + tableName,
        detail: err.response?.data?.message ?? err.message,
      });

      // If we have an explore path but canvas failed, navigate to explore
      if (exploreFilePath && metricsViewName) {
        await goto(`/files${exploreFilePath}`);
        void behaviourEvent?.fireNavigationEvent(
          metricsViewName,
          behaviourEventMedium,
          metricsEventSpace,
          MetricsEventScreenName.Source,
          MetricsEventScreenName.Explore,
        );
      }
    }

    // Done, remove the overlay
    overlay.set(null);
  };
}

// Re-export canvas generation functions for backward compatibility
export {
  createCanvasDashboardFromMetricsView,
  createCanvasDashboardFromMetricsViewWithAgent,
  createCanvasDashboardFromTableWithAgent,
} from "@rilldata/web-common/features/canvas/ai-generation/generateCanvas";
