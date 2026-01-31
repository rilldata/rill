<script lang="ts">
  import { featureFlags } from "../feature-flags";
  import SidebarChat from "./layouts/sidebar/SidebarChat.svelte";
  import { chatOpen } from "./layouts/sidebar/sidebar-store";
  import { dashboardChatConfig } from "@rilldata/web-common/features/dashboards/chat-context.ts";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { canvasChatConfig } from "@rilldata/web-common/features/canvas/chat-context.ts";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { activeDashboardTheme } from "@rilldata/web-common/features/themes/active-dashboard-theme";

  export let kind: ResourceKind.Explore | ResourceKind.Canvas =
    ResourceKind.Explore;

  $: chatConfig =
    kind === ResourceKind.Explore ? dashboardChatConfig : canvasChatConfig;

  const { dashboardChat } = featureFlags;
</script>

{#if $dashboardChat && $chatOpen}
  <ThemeProvider theme={$activeDashboardTheme} applyLayout={false}>
    <SidebarChat config={chatConfig} />
  </ThemeProvider>
{/if}
