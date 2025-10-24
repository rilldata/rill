<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    cleanupConversationManager,
    getConversationManager,
  } from "../../core/conversation-manager";
  import ChatInput from "../../core/input/ChatInput.svelte";

  export let noMargin = false;
  export let height: string | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: conversationManager = getConversationManager(instanceId, {
    conversationState: "url",
  });

  beforeNavigate(({ to }) => {
    const isStillOnHomePage = to?.route?.id === "/[organization]/[project]";
    const isGoingToChatRoute = to?.route?.id?.includes("ai");
    const shouldCleanup = !isStillOnHomePage && !isGoingToChatRoute;

    if (shouldCleanup) {
      cleanupConversationManager(instanceId);
    }
  });
</script>

<ChatInput {conversationManager} {noMargin} {height} />
