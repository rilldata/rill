import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation";
import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager";
import { ToolName } from "@rilldata/web-common/features/chat/core/types";
import { developerChatActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get } from "svelte/store";

/**
 * Dev-agent integration for standalone component (custom viz) files.
 * Mirrors the per-widget chart agent, but targets a component file path instead
 * of a canvas item. Conversations are tracked per file so follow-up prompts
 * continue in context.
 */

const fileConversations = new Map<string, Conversation>();

export function clearComponentFileConversation(filePath: string): void {
  fileConversations.delete(filePath);
}

/**
 * Ask the dev agent to edit a component file, opening the chat sidebar.
 */
export function sendComponentFilePrompt(
  client: RuntimeClient,
  filePath: string,
  userPrompt: string,
): void {
  const conversationManager = getConversationManager(client, {
    conversationState: "browserStorage",
    agent: ToolName.DEVELOPER_AGENT,
    surface: "developer",
  });

  const existing = fileConversations.get(filePath);
  const currentConversation = get(conversationManager.getCurrentConversation());
  if (!existing || existing !== currentConversation) {
    conversationManager.enterNewConversationMode();
  }

  // develop_file with type "component" is what loads the component authoring instructions, so ask
  // for it by name rather than leaving the routing to the agent.
  const fullPrompt = `Update the component file ${filePath} (a standalone custom viz with declared params) by calling develop_file with type "component". ${userPrompt}`;
  developerChatActions.startChat(fullPrompt);

  const conversation = get(conversationManager.getCurrentConversation());
  fileConversations.set(filePath, conversation);
}
