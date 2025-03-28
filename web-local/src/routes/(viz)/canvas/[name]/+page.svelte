<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasThemeProvider from "@rilldata/web-common/features/canvas/CanvasThemeProvider.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: canvasName = data.dashboardName;
</script>

<CanvasThemeProvider {canvasName}>
  <CanvasDashboardEmbed resource={data.dashboard} />
</CanvasThemeProvider>
