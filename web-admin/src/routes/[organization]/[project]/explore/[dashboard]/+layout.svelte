<script lang="ts">
  import DashboardChat from "@rilldata/web-common/features/chat/DashboardChat.svelte";
  import { RecentlyUsedDashboards } from "../../../../../features/dashboards/listing/dashboard-favourites.ts";
  import { onMount } from "svelte";
  import { page } from "$app/stores";

  $: ({ organization, project, dashboard } = $page.params);
  $: recentlyUsedDashboards = new RecentlyUsedDashboards(organization, project);

  onMount(() => {
    recentlyUsedDashboards.update(dashboard.toLowerCase());
  });
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  <DashboardChat />
</div>
