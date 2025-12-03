<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";

  import { dashboardChatConfig } from "@rilldata/web-common/features/dashboards/chat-context.ts";

  const { dashboardChat } = featureFlags;
</script>

{#if $dashboardChat && $chatOpen}
  <SidebarChat config={dashboardChatConfig} />
{/if}
