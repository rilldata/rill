import { goto } from "$app/navigation";
import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager";
import { ToolName } from "@rilldata/web-common/features/chat/core/types";
import { extractMessageText } from "@rilldata/web-common/features/chat/core/utils";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
import { pollForFileCreation } from "@rilldata/web-common/features/entity-management/actions";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getScreenNameFromPage } from "@rilldata/web-common/features/file-explorer/telemetry";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import {
  runtimeServiceGenerateCanvasFile,
  runtimeServiceGenerateMetricsViewFile,
  type V1Message,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import { overlay } from "../../../layout/overlay-store";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import { getName } from "../../entity-management/name-utils";
import { featureFlags } from "../../feature-flags";
import OptionToCancelAIGeneration from "../../metrics-views/ai-generation/OptionToCancelAIGeneration.svelte";

export const generatingCanvas = writable(false);

/**
 * Creates a metrics view from a table.
 * Used internally by canvas generation functions.
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
  void runtimeServiceGenerateMetricsViewFile(instanceId, {
    connector: connector,
    database: database,
    databaseSchema: databaseSchema,
    table: tableName,
    path: newMetricsViewFilePath,
    useAi: isAiEnabled,
  });

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

  return resource;
}

/**
 * Creates a Canvas dashboard from a metrics view using AI.
 */
export async function createCanvasDashboardFromMetricsView(
  instanceId: string,
  metricsViewName: string,
) {
  const isAiEnabled = get(featureFlags.ai);
  const abortController = new AbortController();

  overlay.set({
    title: `Creating Canvas dashboard${isAiEnabled ? " with AI" : ""}...`,
    detail: {
      component: OptionToCancelAIGeneration,
      props: {
        onCancel: () => {
          abortController.abort("Canvas dashboard creation cancelled by user");
        },
      },
    },
  });

  // Get a unique name for the canvas dashboard
  const canvasName = getName(
    `${metricsViewName}_canvas`,
    fileArtifacts.getNamesForKind(ResourceKind.Canvas),
  );
  const canvasFilePath = `/dashboards/${canvasName}.yaml`;

  try {
    // Request AI-generated canvas dashboard
    void runtimeServiceGenerateCanvasFile(
      instanceId,
      {
        metricsViewName: metricsViewName,
        path: canvasFilePath,
        useAi: isAiEnabled,
      },
      abortController.signal,
    );

    // Poll until file creation is complete or canceled
    const fileCreated = await pollForFileCreation(
      instanceId,
      canvasFilePath,
      abortController.signal,
    );

    // If the user canceled the AI request, submit another request with `useAi=false`
    if (!fileCreated) {
      await runtimeServiceGenerateCanvasFile(instanceId, {
        metricsViewName: metricsViewName,
        path: canvasFilePath,
        useAi: false,
      });
    }

    // Navigate to the Canvas file
    await goto(`/files${canvasFilePath}`);
  } catch (err) {
    eventBus.emit("notification", {
      message: "Failed to create Canvas dashboard for " + metricsViewName,
      detail: err.response?.data?.message ?? err.message,
    });
  } finally {
    // Always clean up the overlay
    overlay.set(null);
  }
}

/**
 * Helper function to detect if a canvas file was created in a conversation message.
 * Checks if the message text mentions the expected file path.
 */
function isCanvasFileCreated(
  message: V1Message,
  expectedPath: string,
): boolean {
  // Check if the assistant's message mentions the file path
  if (message.role === "assistant") {
    const messageText = extractMessageText(message);
    return messageText.includes(expectedPath);
  }
  return false;
}

/**
 * Creates a Canvas dashboard from a metrics view using the developer agent.
 * Opens the developer agent sidebar, sends generation prompt, and navigates to the created file.
 */
export function createCanvasDashboardFromMetricsViewWithAgent(
  instanceId: string,
  metricsViewName: string,
): void {
  // 1. Generate unique canvas name
  const canvasName = getName(
    `${metricsViewName}_canvas`,
    fileArtifacts.getNamesForKind(ResourceKind.Canvas),
  );
  const canvasFilePath = `/dashboards/${canvasName}.yaml`;

  // 2. Construct prompt for developer agent
  const prompt = `Create a canvas dashboard at ${canvasFilePath} based on the "${metricsViewName}" metrics view. Include appropriate visualizations like KPI grids, charts, and leaderboards based on the available measures and dimensions.`;

  // 3. Set up file creation detection
  // Get conversation manager and start a new conversation
  const conversationManager = getConversationManager(instanceId, {
    conversationState: "browserStorage",
    agent: ToolName.DEVELOPER_AGENT,
  });

  // Start a new conversation instead of continuing existing one
  conversationManager.enterNewConversationMode();

  const currentConversation = get(conversationManager.getCurrentConversation());

  // Set generating state
  generatingCanvas.set(true);

  // Set up timeout fallback (30s)
  const timeoutId = setTimeout(() => {
    generatingCanvas.set(false);
    eventBus.emit("notification", {
      message: "Canvas generation is taking longer than expected",
      detail: "Check the chat sidebar for progress",
    });
  }, 30000);

  const unsubscribe = currentConversation.on("message", (message) => {
    // Check if this is a file write for our canvas
    if (isCanvasFileCreated(message, canvasFilePath)) {
      clearTimeout(timeoutId);
      generatingCanvas.set(false);
      void goto(`/files${canvasFilePath}`);
      void behaviourEvent?.fireNavigationEvent(
        metricsViewName,
        BehaviourEventMedium.Menu,
        MetricsEventSpace.LeftPanel,
        getScreenNameFromPage(),
        MetricsEventScreenName.Canvas,
      );
      unsubscribe();
    }
  });

  // 4. Start the chat with the generation prompt
  sidebarActions.startChat(prompt);
}

/**
 * Creates a Canvas dashboard from a table (source or model) using the developer agent.
 * First creates a metrics view from the table, then generates the canvas dashboard.
 */
export async function createCanvasDashboardFromTableWithAgent(
  instanceId: string,
  connector: string,
  database: string,
  databaseSchema: string,
  tableName: string,
): Promise<void> {
  const isAiEnabled = get(featureFlags.ai);
  const abortController = new AbortController();

  // Show overlay while creating metrics view
  overlay.set({
    title: `Creating Metrics View${isAiEnabled ? " with AI" : ""}...`,
    detail: {
      component: OptionToCancelAIGeneration,
      props: {
        onCancel: () => {
          abortController.abort("Metrics view creation cancelled by user");
        },
      },
    },
  });

  try {
    // Step 1: Create metrics view from table
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
      throw new Error("Failed to get metrics view name from created resource");
    }

    // Remove overlay before starting agent
    overlay.set(null);

    // Step 2: Generate canvas dashboard using developer agent
    // This will open the chat sidebar and handle the rest of the flow
    createCanvasDashboardFromMetricsViewWithAgent(instanceId, metricsViewName);
  } catch (err) {
    // Remove overlay on error
    overlay.set(null);

    eventBus.emit("notification", {
      message: "Failed to create Metrics View for " + tableName,
      detail: err.response?.data?.message ?? err.message,
    });
  }
}

/**
 * Creates a Canvas dashboard from a metrics view using AI, without navigation.
 * Returns the file path of the created canvas, or null if creation failed.
 */
export async function createCanvasDashboardWithoutNavigation(
  instanceId: string,
  metricsViewName: string,
): Promise<string | null> {
  const isAiEnabled = get(featureFlags.ai);
  const abortController = new AbortController();

  // Get a unique name for the canvas dashboard
  const canvasName = getName(
    `${metricsViewName}_canvas`,
    fileArtifacts.getNamesForKind(ResourceKind.Canvas),
  );
  const canvasFilePath = `/dashboards/${canvasName}.yaml`;

  try {
    // Request AI-generated canvas dashboard
    void runtimeServiceGenerateCanvasFile(
      instanceId,
      {
        metricsViewName: metricsViewName,
        path: canvasFilePath,
        useAi: isAiEnabled,
      },
      abortController.signal,
    );

    // Poll until file creation is complete or canceled
    const fileCreated = await pollForFileCreation(
      instanceId,
      canvasFilePath,
      abortController.signal,
      1000,
    );

    // If the user canceled the AI request, submit another request with `useAi=false`
    if (!fileCreated) {
      await runtimeServiceGenerateCanvasFile(instanceId, {
        metricsViewName: metricsViewName,
        path: canvasFilePath,
        useAi: false,
      });
    }

    return canvasFilePath;
  } catch (err) {
    eventBus.emit("notification", {
      message: "Failed to create Canvas dashboard for " + metricsViewName,
      detail: err.response?.data?.message ?? err.message,
    });
    return null;
  }
}
