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
import {
  ConversationManager,
  getConversationManager,
} from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import OptionCancelToAIAction from "@rilldata/web-common/features/sample-data/OptionCancelToAIAction.svelte";
import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation.ts";

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

    const developerChatEnabled = get(featureFlags.developerChat);

    generatingSampleData.set(true);
    const conversationManager = getConversationManager(instanceId, {
      conversationState: "browserStorage",
      agent: ToolName.DEVELOPER_AGENT,
    });

    let created: boolean;
    let lastReadFile: string | null;
    let cancelled: boolean;

    if (developerChatEnabled) {
      ({ created, lastReadFile, cancelled } =
        await generateSampleDataWithDevChat(agentPrompt, conversationManager));
    } else {
      ({ created, lastReadFile, cancelled } =
        await generateSampleDataWithOverlay(agentPrompt, conversationManager));
    }

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
    eventBus.emit("notification", {
      message: "Failed to generate sample data. Please try again.",
      type: "error",
    });
  } finally {
    overlay.set(null);
    generatingSampleData.set(false);
  }
}

export function maybeNavigateToGeneratedFile(
  conversationManager: ConversationManager,
) {
  let listener: FileWriteListener | null = null;

  const conversationStore = conversationManager.getCurrentConversation();

  const storeUnsub = conversationStore.subscribe((conversation) => {
    listener?.cleanup();
    listener = new FileWriteListener(conversation, {
      onWriteFile: (path) => {
        void goto(`/files${path}`);
      },
    });
  });

  return () => {
    storeUnsub();
    listener?.cleanup();
  };
}

async function generateSampleDataWithDevChat(
  agentPrompt: string,
  conversationManager: ConversationManager,
) {
  // Since the user doesn't see the chat when dev chat is not enabled, older chat shouldn't interfere with this prompt.
  conversationManager.enterNewConversationMode();
  const conversation = get(conversationManager.getCurrentConversation());
  conversation.cancelStream();

  let created = false;
  let lastReadFile: string | null = null;

  const listener = new FileWriteListener(conversation, {
    onReadFile: (path) => {
      // Keep a copy of the file that was read.
      // LLM can some time read a file and decide not to generate data.
      lastReadFile = path;
    },
    onWriteFile: (path) => {
      created = true;
      overlay.set(null);
      sourceImportedPath.set(path);
      void goto(`/files${path}`);
    },
  });

  overlay.set(null);
  sidebarActions.startChat(agentPrompt);
  // Wait for the stream to start async through the sidebar action.
  await waitUntil(() => get(conversation.isStreaming));

  // Then wait for the stream to end.
  await waitUntil(() => !get(conversation.isStreaming), -1);

  listener.cleanup();

  return {
    created,
    lastReadFile,
    cancelled: false,
  };
}

async function generateSampleDataWithOverlay(
  agentPrompt: string,
  conversationManager: ConversationManager,
) {
  const conversation = get(conversationManager.getCurrentConversation());
  conversation.cancelStream();

  let cancelled = false;
  let created = false;
  let lastReadFile: string | null = null;

  const listener = new FileWriteListener(conversation, {
    onReadFile: (path) => {
      // Keep a copy of the file that was read.
      // LLM can some time read a file and decide not to generate data.
      lastReadFile = path;
    },
    onWriteFile: (path) => {
      created = true;
      overlay.set(null);
      sourceImportedPath.set(path);
    },
  });

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

  // Wait for the stream to end.
  await waitUntil(() => !get(conversation.isStreaming), -1);

  listener.cleanup();

  return {
    created,
    lastReadFile,
    cancelled,
  };
}

class FileWriteListener {
  private readonly messages = new Map<string, V1Message>();
  private readonly messageUnsub: () => void;

  constructor(
    conversation: Conversation,
    private readonly callbacks: {
      onReadFile?: (filePath: string) => void;
      onWriteFile?: (filePath: string) => void;
    },
  ) {
    this.messageUnsub = conversation.on("message", (m) =>
      this.handleMessage(m),
    );
  }

  public cleanup() {
    this.messageUnsub();
  }

  private handleMessage(msg: V1Message) {
    this.messages.set(msg.id!, msg);

    if (
      msg.type !== MessageType.RESULT ||
      msg.contentType === MessageContentType.ERROR
    ) {
      return;
    }

    switch (msg.tool) {
      // Sometimes AI detects that model is already present.
      case ToolName.READ_FILE: {
        if (!msg.parentId) break;

        const callMsg = this.messages.get(msg.parentId);
        if (!callMsg) break;

        const path = this.parseFile(callMsg, msg);
        if (path) this.callbacks.onReadFile?.(maybePrependSlash(path));

        break;
      }

      case ToolName.WRITE_FILE: {
        const callMsg = this.messages.get(msg.parentId ?? "");
        if (!callMsg) break;

        const path = this.parseFile(callMsg, msg);
        if (path) this.callbacks.onWriteFile?.(maybePrependSlash(path));

        break;
      }
    }
  }

  private parseFile(call: V1Message, result: V1Message) {
    if (!result.contentData || !call.contentData) return null;

    try {
      const resultContent = JSON.parse(result.contentData);
      const hasErroredOut =
        !!resultContent.parse_error ||
        resultContent.resources?.some((r) => !!r.reconcile_error);
      if (hasErroredOut) return null;

      const callContent = JSON.parse(call.contentData);
      return callContent.path as string;
    } catch {
      // json parse errors shouldn't happen. ignore if it ever does.
    }
    return null;
  }
}

function maybePrependSlash(path: string) {
  return path.startsWith("/") ? path : "/" + path;
}
