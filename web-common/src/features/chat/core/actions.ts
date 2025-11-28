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
import { asyncWait, waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { get } from "svelte/store";
import { EMPTY_PROJECT_TITLE } from "@rilldata/web-common/features/welcome/constants.ts";
import { overlay } from "@rilldata/web-common/layout/overlay-store.ts";
import OptionToAIAction from "@rilldata/web-common/features/chat/core/OptionToAIAction.svelte";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { goto } from "$app/navigation";
import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store.ts";

export async function generateModel(
  isInit: boolean,
  instanceId: string,
  prompt: string,
) {
  try {
    if (isInit) {
      overlay.set({
        title: `Hang tight! We're initialising an empty project.`,
      });

      await runtimeServiceUnpackEmpty(instanceId, {
        displayName: EMPTY_PROJECT_TITLE,
        force: true,
      });

      // TODO: be deterministic about this wait instead
      await asyncWait(1000);
    }

    overlay.set({
      title: `Hang tight! We're generating the data you requested.`,
      detail: {
        component: OptionToAIAction,
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
    conversation.draftMessage.set(prompt);

    let created = false;
    const fileCreated = (msg: V1Message, messagePrefix: string) => {
      const content = JSON.parse(msg.contentData!);
      const path = content.path as string;
      if (!path) return null;
      eventBus.emit("notification", {
        message: `${messagePrefix} ${path}`,
      });
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
          try {
            fileCreated(callMsg, "Data already present at");
          } catch {
            // no-op
          }
          break;
        }

        case ToolName.WRITE_FILE: {
          const callMsg = messages.get(msg.parentId!);
          if (!callMsg) break;
          try {
            sourceImportedPath.set(
              fileCreated(callMsg, "Data generated successfully at"),
            );
          } catch {
            // no-op
          }
          break;
        }
      }
    };

    let cancelled = false;

    conversation.cancelStream();

    await conversation.sendMessage({}, { onMessage: handleMessage });

    await waitUntil(() => get(conversation.isStreaming));

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
  }
}
