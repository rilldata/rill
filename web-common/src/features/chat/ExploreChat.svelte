<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";

  const { dashboardChat } = featureFlags;
</script>

{#if $dashboardChat && $chatOpen}
  <SidebarChat />
{/if}
