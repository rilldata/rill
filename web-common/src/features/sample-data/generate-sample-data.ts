import { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";
import { NEW_CONVERSATION_ID } from "@rilldata/web-common/features/chat/core/utils.ts";
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
import { get } from "svelte/store";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants.ts";
import { overlay } from "@rilldata/web-common/layout/overlay-store.ts";
import OptionCancelToAIAction from "@rilldata/web-common/features/sample-data/OptionCancelToAIAction.svelte";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { goto } from "$app/navigation";
import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store.ts";

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

      const projectResetPromise = new Promise<void>((resolve, reject) => {
        const unsub = eventBus.once("project-reset", () => resolve());
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
    }

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
    const conversation = new Conversation(instanceId, NEW_CONVERSATION_ID, {
      agent: ToolName.DEVELOPER_AGENT,
    });
    const agentPrompt = `Generate a new model for the following user prompt: ${userPrompt}`;
    conversation.draftMessage.set(agentPrompt);

    let created = false;
    const fileCreated = (msg: V1Message) => {
      let path = "";
      try {
        const content = JSON.parse(msg.contentData!);
        path = content.path as string;
      } catch {
        // json parse errors shouldn't happen. ignore if it ever does.
      }
      if (!path) return null;

      created = true;
      overlay.set(null);
      void goto(`/files${path}`);
      return path;
    };

    const messages = new Map<string, V1Message>();
    const handleMessage = (msg: V1Message) => {
      if (created) return; // We already have the file path, no need to process further messages.

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
          const callMsg = messages.get(msg.parentId!);
          if (!callMsg) break;

          const path = fileCreated(callMsg);
          eventBus.emit("notification", {
            message: `Data already present at ${path}`,
          });
          break;
        }

        case ToolName.WRITE_FILE: {
          const callMsg = messages.get(msg.parentId!);
          if (!callMsg) break;

          const path = fileCreated(callMsg);
          sourceImportedPath.set(path);
          break;
        }
      }
    };

    let cancelled = false;

    conversation.cancelStream();

    await conversation.sendMessage({}, { onMessage: handleMessage });

    await waitUntil(() => !get(conversation.isStreaming));

    overlay.set(null);
    if (cancelled) return;
    if (!created) {
      eventBus.emit("notification", {
        message: "Failed to generate sample data",
      });
      return;
    }
  } catch {
    overlay.set(null);
    eventBus.emit("notification", {
      message: "Failed to generate sample data. Please try again.",
      type: "error",
    });
  }
}
