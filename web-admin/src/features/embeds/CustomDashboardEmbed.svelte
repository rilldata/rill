<script lang="ts">
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";
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
  } = dashboard || { items: [], columns: 10, gap: 2 });
</script>

<CustomDashboardEmbed {dashboardName} {columns} {items} {gap} />
