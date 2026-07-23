<script lang="ts">
  import ConnectClientProvider from "@rilldata/web-admin/features/ai/mcp/ConnectClientProvider.svelte";
  import DashboardChat from "@rilldata/web-common/features/chat/DashboardChat.svelte";
  import { RecentlyUsedDashboards } from "../../../../../features/dashboards/listing/dashboard-favourites.ts";
  import { page } from "$app/stores";

  $: ({ organization, project, dashboard } = $page.params);
  $: recentlyUsedDashboards = new RecentlyUsedDashboards(organization, project);

  // Record usage on every dashboard change. The layout persists across
  // param-only navigations, so onMount would only capture the first visit.
  $: if (dashboard) recentlyUsedDashboards.update(dashboard.toLowerCase());
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  <ConnectClientProvider>
    <DashboardChat />
  </ConnectClientProvider>
</div>
