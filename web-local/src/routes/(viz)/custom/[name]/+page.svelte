<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas-dashboards/CanvasDashboardEmbed.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ items = [], columns, gap, variables } = data.dashboard);
</script>

<CanvasDashboardEmbed
  dashboardName={data.dashboardName}
  {items}
  {columns}
  {gap}
  {variables}
/>
