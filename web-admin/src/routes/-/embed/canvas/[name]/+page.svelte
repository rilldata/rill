<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ canvasName, instanceId, navigationEnabled } = data);
</script>

<CanvasProvider {canvasName} {instanceId} showBanner>
  <CanvasDashboardEmbed {canvasName} {navigationEnabled} />
</CanvasProvider>
