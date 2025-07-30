<!-- 
  This layout wraps the chat page to provide proper height constraints
  and passes routing data to the ChatFullPage component
-->
<script lang="ts">
  import { page } from "$app/stores";
  import Chat from "@rilldata/web-common/features/chat/Chat.svelte";
  import type { LayoutData } from "./$types";

  export let data: LayoutData;

  // Extract URL parameters
  $: ({ organization, project } = $page.params);
  $: ({ routeType, conversationId } = data);

  // Type-safe routeType for Chat component
  $: typedRouteType = routeType as "new" | "conversation" | undefined;
</script>

<div class="chat-page-wrapper">
  <Chat
    layout="fullpage"
    routeType={typedRouteType}
    {conversationId}
    {organization}
    {project}
  />
</div>

<style lang="postcss">
  .chat-page-wrapper {
    /* Calculate available height after navigation elements */
    /* Approximate: TopNav (64px) + ProjectTabs (56px) = 120px */
    height: calc(100vh - 120px);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    background: #ffffff;
  }

  /* Ensure proper flex behavior for the chat component */
  :global(.chat-page-wrapper > .chat-fullpage) {
    flex: 1;
    min-height: 0; /* Important for proper flex shrinking */
  }
</style>
