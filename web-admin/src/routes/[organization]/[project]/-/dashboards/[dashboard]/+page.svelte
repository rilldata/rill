<script lang="ts">
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";
  import {
    ResourceKind,
    useResource,
  } from "@rilldata/web-common/features/entity-management/resource-selectors.js";
  import { page } from "$app/stores";
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
  } = dashboard || { items: [], columns: 10, gap: 2 });
</script>

<CustomDashboardEmbed {columns} {items} {gap} />
