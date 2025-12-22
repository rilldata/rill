<script lang="ts">
  import ExploreChat from "@rilldata/web-common/features/chat/ExploreChat.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { activeDashboardTheme } from "@rilldata/web-common/features/themes/active-dashboard-theme";
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  <ThemeProvider theme={$activeDashboardTheme}>
    <ExploreChat />
  </ThemeProvider>
</div>
