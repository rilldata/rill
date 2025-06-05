import { browser } from "$app/environment";
import { get, writable } from "svelte/store";
import { queryClient } from "../../lib/svelte-query/globalQueryClient";
import type { V1Conversation, V1Message } from "../../runtime-client";
import { createRuntimeServiceComplete } from "../../runtime-client";
import { runtime } from "../../runtime-client/runtime-store";
import {
  DEFAULTS,
  getChatOpenState,
  getCurrentConversation,
  getSidebarWidth,
  setChatOpenState,
  setCurrentConversation,
  setSidebarWidth,
} from "./utils/storage";

// Core state stores
export const messages = writable<V1Message[]>([]);
export const currentConversation = writable<V1Conversation | null>(
  getCurrentConversation(),
);
export const sidebarWidth = writable(getSidebarWidth());
export const loading = writable(false);
export const error = writable<string | null>(null);
export const chatOpen = writable(getChatOpenState());
export const complete = createRuntimeServiceComplete({}, queryClient);

// Initialize storage subscriptions
if (browser) {
  // Persist conversation changes to storage
  currentConversation.subscribe((conversation) => {
    setCurrentConversation(conversation);
  });

  // Persist sidebar width changes to storage
  sidebarWidth.subscribe((width) => {
    setSidebarWidth(width);
  });

  // Persist chat open state changes to storage
  chatOpen.subscribe((isOpen) => {
    setChatOpenState(isOpen);
  });
}

// Actions for UI components
export const chatActions = {
  // Send a message and handle conversation creation
  async sendMessage(messageText: string): Promise<void> {
    if (!messageText.trim()) return;

    error.set("");
    loading.set(true);

    // Add user message to UI immediately with a unique temporary ID
    const tempMessageId = crypto.randomUUID();
    const userMessage: V1Message = {
      id: tempMessageId,
      role: "user",
      content: [{ text: messageText }],
    };
    messages.update((msgs) => [...msgs, userMessage]);

    try {
      const current = get(currentConversation);
      const $runtime = get(runtime);
      const wasNewConversation = !current;

      // Call Complete API with current conversation ID (if any)
      const $complete = get(complete);
      const response = await $complete.mutateAsync({
        instanceId: $runtime.instanceId,
        data: {
          conversationId: current?.id,
          messages: [{ role: "user", content: [{ text: messageText }] }],
        },
      });

      // Add all new messages from response to UI (provides full transparency into tool calling)
      const newMessages = response.messages || [];

      // Replace the specific temporary user message with the complete flow from the server
      // This ensures we get proper timestamps and IDs, and shows tool calls/results
      messages.update((msgs) => {
        // Find and remove the specific temporary message by ID
        const filteredMsgs = msgs.filter((msg) => msg.id !== tempMessageId);
        return [...filteredMsgs, ...newMessages];
      });

      // Handle new conversation creation - store the returned conversation ID
      if (wasNewConversation && response.conversationId) {
        // Create conversation title from the user message (similar to backend logic)
        const title = createConversationTitle(messageText);

        // Create a minimal conversation object with the returned ID
        const newConversation: V1Conversation = {
          id: response.conversationId,
          title,
          createdOn: new Date().toISOString(),
          updatedOn: new Date().toISOString(),
          messages: newMessages,
        };

        // Update current conversation with the new conversation ID
        currentConversation.set(newConversation);

        // Invalidate conversations list to refresh it with the new conversation
        // This ensures the sidebar shows the new conversation
        queryClient.invalidateQueries({
          queryKey: ["runtimeServiceListConversations", $runtime.instanceId],
        });
      }
    } catch (e) {
      console.error("Chat: Error sending message", e);

      // Provide more specific error messages
      if (e instanceof Error) {
        if (
          e.message.includes("instance") ||
          e.message.includes("API client")
        ) {
          error.set(e.message);
        } else if (
          e.message.includes("Network Error") ||
          e.message.includes("fetch")
        ) {
          error.set(
            "Could not connect to the server. Please check your connection.",
          );
        } else {
          error.set(`Failed to send message: ${e.message}`);
        }
      } else {
        error.set("Could not get a response from the server.");
      }

      // Remove the specific temporary user message if the request failed
      messages.update((msgs) => msgs.filter((msg) => msg.id !== tempMessageId));
    } finally {
      loading.set(false);
    }
  },

  // Create a new conversation
  createNewConversation(): void {
    messages.set([]);
    currentConversation.set(null);
  },

  // Select a conversation from the list
  selectConversation(conv: V1Conversation): void {
    currentConversation.set(conv);
  },

  // Load messages from conversation data
  loadMessages(conversationMessages: V1Message[]): void {
    messages.set(conversationMessages);
  },

  // Update current conversation (for when new conversation is created)
  updateCurrentConversation(conv: V1Conversation): void {
    const current = get(currentConversation);
    if (!current) {
      currentConversation.set(conv);
    }
  },

  // Update sidebar width with constraints
  updateSidebarWidth(width: number): void {
    const constrainedWidth = Math.max(
      DEFAULTS.MIN_SIDEBAR_WIDTH,
      Math.min(DEFAULTS.MAX_SIDEBAR_WIDTH, width),
    );
    sidebarWidth.set(constrainedWidth);
  },

  // Toggle chat open/closed
  toggleChat(): void {
    chatOpen.update((isOpen) => !isOpen);
  },

  // Close chat
  closeChat(): void {
    chatOpen.set(false);
  },
};

// ===== UTILITY FUNCTIONS =====

// createConversationTitle creates a conversation title from the user message
// Similar to the backend implementation in completion.go
function createConversationTitle(messageText: string): string {
  let title = messageText.trim();

  // Truncate to 50 characters and add ellipsis if needed
  if (title.length > 50) {
    title = title.substring(0, 50) + "...";
  }

  // Replace newlines with spaces
  title = title.replace(/\n/g, " ").replace(/\r/g, " ");

  // Collapse multiple spaces
  while (title.includes("  ")) {
    title = title.replace(/  /g, " ");
  }

  return title || "New Conversation";
}
