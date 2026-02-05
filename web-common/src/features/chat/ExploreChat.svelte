<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { activeDashboardTheme } from "@rilldata/web-common/features/themes/active-dashboard-theme";

  import { dashboardChatConfig } from "@rilldata/web-common/features/dashboards/chat-context.ts";

  const { dashboardChat } = featureFlags;
</script>

{#if $dashboardChat && $chatOpen}
  <ThemeProvider theme={$activeDashboardTheme} applyLayout={false}>
    <SidebarChat config={dashboardChatConfig} />
  </ThemeProvider>
{/if}
