<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    getLocalConversationManager,
    cleanupLocalConversationManager,
  } from "../ai/local-conversation-manager";
  import ChatInput from "@rilldata/web-common/features/chat/core/input/ChatInput.svelte";
  import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager";
  import { projectChat } from "@rilldata/web-common/features/project/chat-context";

  export let noMargin = false;
  export let height: string | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: localManager = getLocalConversationManager(instanceId);
  $: conversationManager = localManager as unknown as ConversationManager;

  beforeNavigate(({ to }) => {
    const isStillOnHomePage = to?.route?.id === "/home";
    const isGoingToChatRoute = to?.route?.id?.includes("ai");
    const shouldCleanup = !isStillOnHomePage && !isGoingToChatRoute;

    if (shouldCleanup) {
      cleanupLocalConversationManager(instanceId);
    }
  });
</script>

<ChatInput {conversationManager} {noMargin} {height} config={projectChat} />
