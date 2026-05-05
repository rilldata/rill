import type { Conversation } from "@rilldata/web-common/features/chat/core/conversation";
import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager";
import { ToolName } from "@rilldata/web-common/features/chat/core/types";
import { sidebarActions } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { derived, get, type Readable } from "svelte/store";
import type { CustomChartComponent } from "./index";

const componentConversations = new Map<string, Conversation>();

export function clearComponentConversation(componentId: string): void {
  componentConversations.delete(componentId);
}

/**
 * Build a prompt with enough context for the dev agent to locate and edit the component.
 */
function buildPrompt(
  component: CustomChartComponent,
  userPrompt: string,
): string {
  const canvasName = component.parent.name;
  const canvasFilePath = `/dashboards/${canvasName}.yaml`;
  const [, rowIdx, , colIdx] = component.pathInYAML;

  return `In the canvas dashboard file ${canvasFilePath}, update the custom_chart component at row ${rowIdx}, item ${colIdx}. Write the metrics_sql and vega_spec for this chart: ${userPrompt}`;
}

/**
 * Send a prompt to the dev agent sidebar for this custom chart component.
 * Continues the existing conversation if one exists for this component;
 * starts a new conversation on first use.
 */
export function sendToDevAgent(
  client: RuntimeClient,
  component: CustomChartComponent,
  userPrompt: string,
): void {
  const conversationManager = getConversationManager(client, {
    conversationState: "browserStorage",
    agent: ToolName.DEVELOPER_AGENT,
  });

  const existing = componentConversations.get(component.id);
  const currentConversation = get(conversationManager.getCurrentConversation());

  // Start a new conversation only if we don't have one for this component,
  // or if the sidebar has moved to a different conversation
  if (!existing || existing !== currentConversation) {
    conversationManager.enterNewConversationMode();
  }

  const fullPrompt = buildPrompt(component, userPrompt);
  sidebarActions.startChat(fullPrompt);

  // Track the conversation for this component so subsequent calls continue it
  const conversation = get(conversationManager.getCurrentConversation());
  componentConversations.set(component.id, conversation);
}

/**
 * Get a reactive store that indicates whether the dev agent is currently
 * streaming a response for this specific component.
 */
export function getAgentStreamingStore(
  client: RuntimeClient,
  componentId: string,
): Readable<boolean> {
  const conversationManager = getConversationManager(client, {
    conversationState: "browserStorage",
    agent: ToolName.DEVELOPER_AGENT,
  });

  return derived(
    conversationManager.getCurrentConversation(),
    ($conversation, set) => {
      const tracked = componentConversations.get(componentId);
      if (!tracked || tracked !== $conversation) {
        set(false);
        return;
      }
      // Subscribe to the tracked conversation's streaming state
      const unsub = tracked.isStreaming.subscribe(set);
      return unsub;
    },
  );
}
