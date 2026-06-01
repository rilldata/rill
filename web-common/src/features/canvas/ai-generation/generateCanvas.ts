import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager";
import { ToolName } from "@rilldata/web-common/features/chat/core/types";
import { developerChatActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
import { pollForFileCreation } from "@rilldata/web-common/features/entity-management/actions/actions.ts";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { navigateToFile } from "@rilldata/web-common/layout/navigation/editor-routing";
import { extractErrorMessage } from "@rilldata/web-common/lib/errors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { V1Resource } from "@rilldata/web-common/runtime-client";
import {
  runtimeServiceGenerateCanvasFile,
  runtimeServiceGenerateMetricsViewFile,
  runtimeServicePutFile,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get, writable } from "svelte/store";
import { overlay } from "../../../layout/overlay-store";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import { getName } from "../../entity-management/name-utils";
import { featureFlags } from "../../feature-flags";
import OptionToCancelAIGeneration from "../../metrics-views/ai-generation/OptionToCancelAIGeneration.svelte";

export const generatingCanvasFilePath = writable<string | null>(null);

/**
 * Creates a metrics view from a table.
 * Used internally by canvas generation functions.
 */
async function createMetricsViewFromTable(
  client: RuntimeClient,
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
  void runtimeServiceGenerateMetricsViewFile(client, {
    connector: connector,
    database: database,
    databaseSchema: databaseSchema,
    table: tableName,
    path: newMetricsViewFilePath,
    useAi: isAiEnabled,
  });

  // Poll until file creation is complete or canceled
  const fileCreated = await pollForFileCreation(
    client,
    newMetricsViewFilePath,
    abortController.signal,
  );

  // If the user canceled the AI request, submit another request with `useAi=false`
  if (!fileCreated) {
    await runtimeServiceGenerateMetricsViewFile(client, {
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
    .getResource(queryClient);

  await waitUntil(() => get(metricsViewResource).data !== undefined, 5000);

  const resource = get(metricsViewResource).data;
  if (!resource) {
    throw new Error("Failed to create a Metrics View resource");
  }

  return resource;
}

/**
 * Creates a Canvas dashboard from a metrics view using AI.
 * TODO: Delete after remvoing developerChat feature flag
 */
export async function createCanvasDashboardFromMetricsView(
  client: RuntimeClient,
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
    void runtimeServiceGenerateCanvasFile(client, {
      metricsViewName: metricsViewName,
      path: canvasFilePath,
      useAi: isAiEnabled,
    });

    // Poll until file creation is complete or canceled
    const fileCreated = await pollForFileCreation(
      client,
      canvasFilePath,
      abortController.signal,
    );

    // If the user canceled the AI request, submit another request with `useAi=false`
    if (!fileCreated) {
      await runtimeServiceGenerateCanvasFile(client, {
        metricsViewName: metricsViewName,
        path: canvasFilePath,
        useAi: false,
      });
    }

    // Navigate to the Canvas file
    await navigateToFile(canvasFilePath);
  } catch (err) {
    eventBus.emit("notification", {
      message: "Failed to create Canvas dashboard for " + metricsViewName,
      detail: extractErrorMessage(err),
    });
  } finally {
    // Always clean up the overlay
    overlay.set(null);
  }
}

/**
 * Creates a Canvas dashboard from a metrics view using the developer agent.
 * Opens the developer agent sidebar, sends generation prompt, and navigates to the created file.
 */
export async function createCanvasDashboardFromMetricsViewWithAgent(
  client: RuntimeClient,
  metricsViewName: string,
): Promise<void> {
  // 1. Generate unique canvas name
  const canvasName = getName(
    `${metricsViewName}_canvas`,
    fileArtifacts.getNamesForKind(ResourceKind.Canvas),
  );
  const canvasFilePath = `/dashboards/${canvasName}.yaml`;

  // 2. Construct prompt for developer agent
  const prompt = `Create a canvas dashboard at ${canvasFilePath} based on the "${metricsViewName}" metrics view. Include appropriate visualizations like KPI grids, charts, and leaderboards based on the available measures and dimensions.`;

  try {
    // Set generating state and create a placeholder file so it appears in the
    // left nav with a loading spinner before the agent has written anything.
    generatingCanvasFilePath.set(canvasFilePath);
    await runtimeServicePutFile(client, {
      path: canvasFilePath,
      blob: "type: canvas\n",
      create: true,
      createOnly: true,
    });

    // 3. Set up file creation detection
    // Get conversation manager and start a new conversation
    const conversationManager = getConversationManager(client, {
      conversationState: "browserStorage",
      agent: ToolName.DEVELOPER_AGENT,
      surface: "developer",
    });

    // Start a new conversation instead of continuing existing one
    conversationManager.enterNewConversationMode();

    const currentConversation = get(
      conversationManager.getCurrentConversation(),
    );

    // 4. Start the chat with the generation prompt
    developerChatActions.startChat(prompt);

    // Wait for the stream to start async through the sidebar action.
    await waitUntil(() => get(currentConversation.isStreaming));

    // Then wait for the stream to end before checking for file creation
    await waitUntil(() => !get(currentConversation.isStreaming), -1);
  } catch (err) {
    console.error("Error generating canvas with agent:", err);
    eventBus.emit("notification", {
      message: "Failed to generate canvas dashboard",
      detail: err instanceof Error ? err.message : String(err),
    });
  } finally {
    generatingCanvasFilePath.set(null);
  }
}

/**
 * Creates a Canvas dashboard from a table (source or model) using the developer agent.
 * First creates a metrics view from the table, then generates the canvas dashboard.
 */
export async function createCanvasDashboardFromTableWithAgent(
  client: RuntimeClient,
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
      client,
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
    void createCanvasDashboardFromMetricsViewWithAgent(client, metricsViewName);
  } catch (err) {
    // Remove overlay on error
    overlay.set(null);

    eventBus.emit("notification", {
      message: "Failed to create Metrics View for " + tableName,
      detail: extractErrorMessage(err),
    });
  }
}

/**
 * Creates a Canvas dashboard from a metrics view using AI, without navigation.
 * Returns the file path of the created canvas, or null if creation failed.
 */
export async function createCanvasDashboardWithoutNavigation(
  client: RuntimeClient,
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
    void runtimeServiceGenerateCanvasFile(client, {
      metricsViewName: metricsViewName,
      path: canvasFilePath,
      useAi: isAiEnabled,
    });

    // Poll until file creation is complete or canceled
    const fileCreated = await pollForFileCreation(
      client,
      canvasFilePath,
      abortController.signal,
      1000,
    );

    // If the user canceled the AI request, submit another request with `useAi=false`
    if (!fileCreated) {
      await runtimeServiceGenerateCanvasFile(client, {
        metricsViewName: metricsViewName,
        path: canvasFilePath,
        useAi: false,
      });
    }

    return canvasFilePath;
  } catch (err) {
    eventBus.emit("notification", {
      message: "Failed to create Canvas dashboard for " + metricsViewName,
      detail: extractErrorMessage(err),
    });
    return null;
  }
}
