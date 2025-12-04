<script lang="ts">
  import CanvasDashboardEmbed from "@rilldata/web-common/features/canvas/CanvasDashboardEmbed.svelte";
  import CanvasProvider from "@rilldata/web-common/features/canvas/CanvasProvider.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.js";

  export let data;

  $: ({ canvasName } = data);

  $: ({ instanceId } = $runtime);
</script>

<CanvasProvider {canvasName} {instanceId} showBanner>
  <CanvasDashboardEmbed {canvasName} />
</CanvasProvider>
