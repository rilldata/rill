<script lang="ts">
  import ChatFullPage from "./layouts/fullpage/ChatFullPage.svelte";
  import ChatSidebar from "./layouts/sidebar/ChatSidebar.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";

  export let layout: "sidebar" | "fullpage" = "sidebar";

  // Route context for fullpage navigation (only needed for fullpage layout)
  export let routeType: "new" | "conversation" | undefined = undefined;
  export let conversationId: string | undefined = undefined;
  export let organization: string = "";
  export let project: string = "";
</script>

{#if layout === "fullpage"}
  <ChatFullPage {routeType} {conversationId} {organization} {project} />
{:else if $chatOpen}
  <ChatSidebar />
{/if}
