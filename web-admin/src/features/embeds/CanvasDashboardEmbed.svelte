<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas-dashboards/CanvasDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";

  export let instanceId: string;
  export let dashboardName: string;

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
