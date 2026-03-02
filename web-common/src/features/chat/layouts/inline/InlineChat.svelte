<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    cleanupConversationManager,
    getConversationManager,
  } from "../../core/conversation-manager";
  import ChatInput from "../../core/input/ChatInput.svelte";
  import { projectChat } from "@rilldata/web-common/features/project/chat-context.ts";

  export let noMargin = false;
  export let height: string | undefined = undefined;

  const runtimeClient = useRuntimeClient();
  $: instanceId = runtimeClient.instanceId;

  $: conversationManager = getConversationManager(runtimeClient, {
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

<ChatInput
  inline
  {conversationManager}
  {noMargin}
  {height}
  config={projectChat}
/>
