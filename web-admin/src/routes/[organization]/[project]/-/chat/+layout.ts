import { featureFlags } from "@rilldata/web-common/features/feature-flags.js";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({
  params: { organization, project, conversationId },
  route,
  parent,
  url,
}) => {
  const {
    runtime: { instanceId },
  } = await parent();

  // NOTE: in the future, we'll use user-level `ai` permissions to determine access to the chat page.
  if (!get(featureFlags.chat)) {
    throw redirect(307, `/${organization}/${project}/-/dashboards`);
  }

  // Handle different chat routes
  const routeId = route.id;

  switch (routeId) {
    // Base chat route: redirect to last conversation or new conversation
    case "/[organization]/[project]/-/chat": {
      const lastConversationId = sessionStorage.getItem(
        "current-conversation-id",
      );
      if (lastConversationId && lastConversationId !== "null") {
        throw redirect(
          307,
          `/${organization}/${project}/-/chat/${lastConversationId}`,
        );
      }

      // No existing conversation, redirect to new conversation
      throw redirect(307, `/${organization}/${project}/-/chat/new`);
    }

    // New conversation route
    case "/[organization]/[project]/-/chat/new": {
      return {
        routeType: "new",
      };
    }

    // Conversation ID route
    case "/[organization]/[project]/-/chat/[conversationId]": {
      if (!conversationId || conversationId.trim() === "") {
        throw redirect(307, `/${organization}/${project}/-/chat/new`);
      }

      return {
        routeType: "conversation",
        conversationId: conversationId,
      };
    }
    default: {
      throw new Error(`Unknown chat route: ${routeId}`);
    }
  }
};
