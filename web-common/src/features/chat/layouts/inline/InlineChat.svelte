<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import httpClient from "@rilldata/web-common/runtime-client/http-client";
  import {
    cleanupConversationManager,
    getConversationManager,
  } from "../../core/conversation-manager";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import { projectChat } from "@rilldata/web-common/features/project/chat-context.ts";

  export let noMargin = false;
  export let height: string | undefined = undefined;

  const instanceId = httpClient.getInstanceId();

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

<ChatInput {conversationManager} {noMargin} {height} config={projectChat} />
