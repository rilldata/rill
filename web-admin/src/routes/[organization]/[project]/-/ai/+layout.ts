import {
  getLastConversationId,
  setLastConversationId,
} from "@rilldata/web-common/features/chat/layouts/fullpage/fullpage-store";
import { featureFlags } from "@rilldata/web-common/features/feature-flags.js";
import { redirect } from "@sveltejs/kit";
import { get } from "svelte/store";

export const load = async ({
  params: { organization, project, conversationId },
  route,
  url,
  parent,
}) => {
  // Wait for the feature flags to load
  await parent();

  // There is a potential race condition where feature flags from instance is not loaded yet.
  // So wait until it is ready before checking for "chat"
  await featureFlags.ready;

  // Redirect to `/-/dashboards` if chat feature is disabled
  // NOTE: In the future, we'll use user-level `ai` permissions for more granular access control
  const chatEnabled = get(featureFlags.chat);
  if (!chatEnabled) {
    throw redirect(307, `/${organization}/${project}/-/dashboards`);
  }

  switch (route.id) {
    case "/[organization]/[project]/-/ai": {
      // If user explicitly wants a new conversation, clear stored ID and skip redirect logic
      const isExplicitNewConversation = url.searchParams.get("new") === "true";
      if (isExplicitNewConversation) {
        setLastConversationId(organization, project, null);
        return;
      }

      // Try to redirect to the last conversation
      const lastConversationId = getLastConversationId(organization, project);
      if (lastConversationId) {
        throw redirect(
          307,
          `/${organization}/${project}/-/ai/${lastConversationId}`,
        );
      }

      // No existing conversation found, show new conversation interface
      // This is the default case when the user first visits the chat page
      return;
    }

    case "/[organization]/[project]/-/ai/[conversationId]": {
      // If conversation ID is missing or empty, redirect to base chat
      if (!conversationId?.trim()) {
        throw redirect(307, `/${organization}/${project}/-/ai`);
      }

      // Store this conversation ID as the last accessed conversation
      setLastConversationId(organization, project, conversationId);

      // Go to the conversation
      return;
    }

    default: {
      throw new Error(`Unknown chat route: ${route.id}`);
    }
  }
};
