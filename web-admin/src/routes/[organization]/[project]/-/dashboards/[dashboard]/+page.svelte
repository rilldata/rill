<script lang="ts">
  import { page } from "$app/stores";
  import CustomDashboard from "@rilldata/web-common/features/custom-dashboards/CustomDashboard.svelte";
  import { useCustomDashboard } from "@rilldata/web-common/features/custom-dashboards/selectors";
  import type { V1DashboardComponent } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  $: customDashboardName = $page.params.dashboard;
  $: query = useCustomDashboard($runtime.instanceId, customDashboardName);
  $: dashboard = $query.data?.dashboard?.spec;
  $: columns = dashboard?.columns ?? 10;
  $: gap = dashboard?.gap ?? 1;
  $: charts = dashboard?.components ?? ([] as V1DashboardComponent[]);
</script>

<CustomDashboard
  snap={false}
  {gap}
  {charts}
  {columns}
  showGrid={false}
  selectedChartName=""
/>
