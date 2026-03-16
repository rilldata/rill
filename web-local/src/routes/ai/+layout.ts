import { browser } from "$app/environment";
import {
  getLastConversationId,
  setLastConversationId,
} from "@rilldata/web-common/features/chat/layouts/fullpage/fullpage-store";
import { redirect } from "@sveltejs/kit";

// Disable SSR for AI chat routes - required for sessionStorage access
export const ssr = false;

export const load = async ({ params, route, url }) => {
  // Extra guard for browser environment
  if (!browser) {
    return;
  }

  const { conversationId } = params;
  // Use a fixed key for local development since there's no org/project
  const org = "local";
  const project = "dev";

  switch (route.id) {
    case "/ai": {
      // If user explicitly wants a new conversation, clear stored ID and skip redirect logic
      const isExplicitNewConversation = url.searchParams.get("new") === "true";
      if (isExplicitNewConversation) {
        setLastConversationId(org, project, null);
        return;
      }

      // Try to redirect to the last conversation
      const lastConversationId = getLastConversationId(org, project);
      if (lastConversationId) {
        throw redirect(307, `/ai/${lastConversationId}`);
      }

      // No existing conversation found, show new conversation interface
      return;
    }

    case "/ai/[conversationId]": {
      // If conversation ID is missing or empty, redirect to base chat
      if (!conversationId?.trim()) {
        throw redirect(307, `/ai`);
      }

      // Store this conversation ID as the last accessed conversation
      setLastConversationId(org, project, conversationId);

      // Go to the conversation
      return;
    }

    default: {
      // Allow unknown routes to pass through
      return;
    }
  }
};
