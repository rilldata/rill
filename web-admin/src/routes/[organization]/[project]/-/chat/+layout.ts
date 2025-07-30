import { getLastConversationId } from "@rilldata/web-common/features/chat/layouts/fullpage/fullpage-store";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.js";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({
  params: { organization, project, conversationId },
  route,
  url,
}) => {
  // Redirect to `/-/dashboards` if chat feature is disabled
  // NOTE: In the future, we'll use user-level `ai` permissions for more granular access control
  if (!get(featureFlags.chat)) {
    throw redirect(307, `/${organization}/${project}/-/dashboards`);
  }

  const routeId = route.id;

  switch (routeId) {
    case "/[organization]/[project]/-/chat": {
      const isExplicitNewConversation = url.searchParams.get("new") === "true";

      // If user explicitly wants a new conversation, skip redirect logic
      if (isExplicitNewConversation) {
        return { conversationId: null };
      }

      // Try to redirect to the last conversation
      const lastConversationId = getLastConversationId(organization, project);

      if (lastConversationId) {
        throw redirect(
          307,
          `/${organization}/${project}/-/chat/${lastConversationId}`,
        );
      }

      // No existing conversation found, show new conversation interface
      return { conversationId: null };
    }

    case "/[organization]/[project]/-/chat/[conversationId]": {
      // Redirect to base chat if conversation ID is missing or empty
      if (!conversationId?.trim()) {
        throw redirect(307, `/${organization}/${project}/-/chat`);
      }

      return { conversationId };
    }

    default: {
      throw new Error(`Unknown chat route: ${routeId}`);
    }
  }
};
