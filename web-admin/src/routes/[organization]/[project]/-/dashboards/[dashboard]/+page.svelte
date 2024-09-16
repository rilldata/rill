<script lang="ts">
  import { page } from "$app/stores";
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas-dashboards/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";

  $: instanceId = $runtime?.instanceId;
  $: dashboardName = $page.params.dashboard;

  $: dashboardQuery = useResource(
    instanceId,
    dashboardName,
    ResourceKind.Dashboard,
  );

  $: dashboard = $dashboardQuery.data?.dashboard.spec;

  $: ({
    items = [],
    columns,
    gap,
  } = dashboard || { items: [], columns: 24, gap: 2 });
</script>

<CanvasDashboardEmbed {dashboardName} {columns} {items} {gap} />
