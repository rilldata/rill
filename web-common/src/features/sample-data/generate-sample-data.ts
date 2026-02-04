import { runtimeServiceUnpackEmpty } from "@rilldata/web-common/runtime-client";
import { ToolName } from "@rilldata/web-common/features/chat/core/types.ts";
import { get, writable } from "svelte/store";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants.ts";
import { overlay } from "@rilldata/web-common/layout/overlay-store.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";

export const generatingSampleData = writable(false);
const PROJECT_INIT_TIMEOUT_MS = 10_000;

export async function generateSampleData(
  initializeProject: boolean,
  instanceId: string,
  userPrompt: string,
) {
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
      overlay.set(null);
    }

    generatingSampleData.set(true);
    const conversationManager = getConversationManager(instanceId, {
      conversationState: "browserStorage",
      agent: ToolName.DEVELOPER_AGENT,
    });

    const conversation = get(conversationManager.getCurrentConversation());
    conversation.cancelStream();

    // Open the chat panel
    sidebarActions.openChat();

    // Send the message directly.
    // - For project init: pass initProject context so the agent handles read-only OLAP gracefully
    // - For existing projects: prefix the prompt so the agent knows the intent
    const prompt = initializeProject
      ? userPrompt
      : `Generate sample data about: ${userPrompt}`;
    conversation.draftMessage.set(prompt);

    const context = initializeProject
      ? { developerAgentContext: { initProject: true } }
      : {};
    await conversation.sendMessage(context);
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
