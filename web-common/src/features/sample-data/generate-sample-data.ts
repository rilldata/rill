import { runtimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
import { ToolName } from "@rilldata/web-common/features/chat/core/types.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { get, writable } from "svelte/store";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants.ts";
import { overlay } from "@rilldata/web-common/layout/overlay-store.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";

export const generatingSampleData = writable(false);
const PROJECT_INIT_TIMEOUT_MS = 10_000;

export async function generateSampleData(
  initializeProject: boolean,
  instanceId: string,
  userPrompt: string,
) {
  const agentPrompt = `Generate a new model with fresh data for the following user prompt: ${userPrompt}`;

  try {
    if (initializeProject) {
      overlay.set({
        title: `Hang tight! We're initialising an empty project.`,
      });

      // UnpackEmpty create a new rill.yaml file. In backend it triggers a reset and cancels any pending requests.
      // The way we get around this is by invalidating all queries in WatchFilesClient on a rill.yaml write.
      // On a rill.yaml write, WatchFilesClient also fires `rill-yaml-updated` which acts as a signal here to make sure any new requests are not canceled.
      const projectResetPromise = new Promise<void>((resolve, reject) => {
        const unsub = eventBus.once("rill-yaml-updated", () => resolve());
        setTimeout(() => {
          reject(new Error("Project init timed out"));
          unsub();
        }, PROJECT_INIT_TIMEOUT_MS);
      });

      await runtimeServiceUnpackEmpty(instanceId, {
        displayName: EMPTY_PROJECT_TITLE,
        force: true,
      });

      await projectResetPromise;
      await featureFlags.ready;
      overlay.set(null);
    }

    generatingSampleData.set(true);
    const conversationManager = getConversationManager(instanceId, {
      conversationState: "browserStorage",
      agent: ToolName.DEVELOPER_AGENT,
    });

    // Continue with the current chat. We might want to revisit this based on feedback.
    const conversation = get(conversationManager.getCurrentConversation());
    conversation.cancelStream();

    sidebarActions.startChat(agentPrompt);
    // Wait for the stream to start async through the sidebar action.
    await waitUntil(() => get(conversation.isStreaming));

    // Then wait for the stream to end.
    await waitUntil(() => !get(conversation.isStreaming), -1);
  } catch (err) {
    console.error(err);
    eventBus.emit("notification", {
      message: "Failed to generate sample data. Please try again.",
      type: "error",
    });
  } finally {
    overlay.set(null);
    generatingSampleData.set(false);
  }
}
