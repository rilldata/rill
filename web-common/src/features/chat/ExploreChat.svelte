<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import ChatSidebar from "./layouts/sidebar/ChatSidebar.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";

  const { dashboardChat } = featureFlags;
</script>

{#if $dashboardChat && $chatOpen}
  <ChatSidebar />
{/if}
