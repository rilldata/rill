<script lang="ts">
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ items = [], columns, gap, variables } = data.dashboard);
</script>

<CustomDashboardEmbed
  dashboardName={data.dashboardName}
  {items}
  {columns}
  {gap}
  {variables}
/>
