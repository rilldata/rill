<script lang="ts">
  import EmbedHeader from "@rilldata/web-admin/features/embeds/EmbedHeader.svelte";
  import DashboardChat from "@rilldata/web-common/features/chat/DashboardChat.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
  import { useFeatureFlags } from "@rilldata/web-common/features/feature-flags";
  import { activeDashboardTheme } from "@rilldata/web-common/features/themes/active-dashboard-theme";

  export let activeResource: { kind: ResourceKind; name: string };
  export let navigationEnabled: boolean;
  export let onProjectPage: boolean;

  const { dashboardChat } = useFeatureFlags();

  $: showDashboardChat = $dashboardChat && !onProjectPage;
  // Resource kind can be metrics view in some cases. But internally to render it will have to have an equivalent explore.
  $: correctedKindForChat =
    activeResource?.kind === ResourceKind.MetricsView
      ? ResourceKind.Explore
      : (activeResource?.kind as
          | ResourceKind.Explore
          | ResourceKind.Canvas
          | undefined);

  $: showTopBar = navigationEnabled || showDashboardChat;
</script>

{#if showTopBar}
  <ThemeProvider theme={$activeDashboardTheme} applyLayout={false}>
    <div
      class="flex items-center w-full pr-4 py-1 min-h-[2.5rem] bg-surface-subtle"
      class:border-b={!onProjectPage}
    >
      <EmbedHeader {activeResource} {navigationEnabled} />
    </div>
  </ThemeProvider>
{/if}

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  {#if showDashboardChat && correctedKindForChat}
    <DashboardChat kind={correctedKindForChat} />
  {/if}
</div>
