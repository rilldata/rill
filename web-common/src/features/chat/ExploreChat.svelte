<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";

  export let agent: string | undefined = undefined;

  const { dashboardChat, developerChat } = featureFlags;
</script>

{#if ($dashboardChat || $developerChat) && $chatOpen}
  <SidebarChat {agent} />
{/if}
