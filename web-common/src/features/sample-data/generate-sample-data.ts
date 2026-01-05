import {
  runtimeServiceUnpackEmpty,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import {
  MessageContentType,
  MessageType,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { get, writable } from "svelte/store";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants.ts";
import { overlay } from "@rilldata/web-common/layout/overlay-store.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { goto } from "$app/navigation";
import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store.ts";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import OptionCancelToAIAction from "@rilldata/web-common/features/sample-data/OptionCancelToAIAction.svelte";

export const generatingSampleData = writable(false);
const PROJECT_INIT_TIMEOUT_MS = 10_000;

export async function generateSampleData(
  initializeProject: boolean,
  instanceId: string,
  userPrompt: string,
) {
  const agentPrompt = `Generate a NEW model with fresh data for the following user prompt: ${userPrompt}`;

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
    }

    generatingSampleData.set(true);
    const conversationManager = getConversationManager(instanceId, {
      conversationState: "browserStorage",
      agent: ToolName.DEVELOPER_AGENT,
    });
    conversationManager.enterNewConversationMode();
    const conversation = get(conversationManager.getCurrentConversation());

    const developerChatEnabled = get(featureFlags.dashboardChat);
    const showImportSourcePopup = !developerChatEnabled;

    let created = false;
    let lastReadFile: string | null = null;
    const messages = new Map<string, V1Message>();

    const parseFile = (call: V1Message, result: V1Message) => {
      try {
        const resultContent = JSON.parse(result.contentData!);
        const hasErroredOut =
          !!resultContent.parse_error ||
          resultContent.resources?.some((r) => !!r.reconcile_error);
        if (hasErroredOut) return null;

        const callContent = JSON.parse(call.contentData!);
        return callContent.path as string;
      } catch {
        // json parse errors shouldn't happen. ignore if it ever does.
      }
      return null;
    };

    const handleMessage = (msg: V1Message) => {
      messages.set(msg.id!, msg);
      if (
        msg.type !== MessageType.RESULT ||
        msg.contentType === MessageContentType.ERROR
      ) {
        return;
      }

      switch (msg.tool) {
        // Sometimes AI detects that model is already present.
        case ToolName.READ_FILE: {
          const callMsg = messages.get(msg.parentId ?? "");
          if (!callMsg) break;

          // Keep a copy of the file that was read.
          // LLM can some time read a file and decide not to generate data.
          lastReadFile = parseFile(callMsg, msg);
          break;
        }

        case ToolName.WRITE_FILE: {
          const callMsg = messages.get(msg.parentId ?? "");
          if (!callMsg) break;

          const path = parseFile(callMsg, msg);
          if (!path) break;

          if (showImportSourcePopup) sourceImportedPath.set(path);
          created = true;
          overlay.set(null);
          void goto(`/files${path}`);
          break;
        }
      }
    };
    const handleMessageUnsub = conversation.on("message", handleMessage);

    let cancelled = false;

    conversation.cancelStream();

    if (developerChatEnabled) {
      overlay.set(null);
      sidebarActions.startChat(agentPrompt);
      await waitUntil(() => get(conversation.isStreaming));
    } else {
      overlay.set({
        title: `Hang tight! We're generating the data you requested.`,
        detail: {
          component: OptionCancelToAIAction,
          props: {
            onCancel: () => {
              conversation.cancelStream();
              cancelled = true;
            },
          },
        },
      });
      conversation.draftMessage.set(agentPrompt);
      await conversation.sendMessage({});
    }

    await waitUntil(() => !get(conversation.isStreaming));

    handleMessageUnsub();
    generatingSampleData.set(false);
    if (cancelled) return;
    if (!created) {
      if (lastReadFile) {
        eventBus.emit("notification", {
          message: `Data already present at ${lastReadFile}`,
        });
      } else {
        eventBus.emit("notification", {
          message: "Failed to generate sample data",
        });
      }
      return;
    }
  } catch (err) {
    console.error(err);
    overlay.set(null);
    generatingSampleData.set(false);
    eventBus.emit("notification", {
      message: "Failed to generate sample data. Please try again.",
      type: "error",
    });
  }
}
