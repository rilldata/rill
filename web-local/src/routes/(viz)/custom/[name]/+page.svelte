<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ items = [], columns, gap, variables } = data.dashboard);
</script>

<CanvasDashboardEmbed
  canvasName={data.dashboardName}
  {items}
  {columns}
  {gap}
  {variables}
/>
